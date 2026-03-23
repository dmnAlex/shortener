package grpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/middleware"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
	pb "github.com/dmnAlex/shortener/proto"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

const buffSize = 1024 * 1024

func TestGRPCShortener(t *testing.T) {
	listen := bufconn.Listen(buffSize)
	defer listen.Close()

	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
		JWTSecret:      "test-secret",
	}

	repo, err := repository.NewFileRepo("")
	require.NoError(t, err)
	defer repo.Close()

	svc := service.NewURLService(repo)

	srv := grpc.NewServer(grpc.UnaryInterceptor(middleware.GRPCAuth(cfg)))
	pb.RegisterShortenerServiceServer(srv, NewShortenerServer(svc, cfg, nil))

	go srv.Serve(listen)
	defer srv.GracefulStop()

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return listen.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewShortenerServiceClient(conn)

	claims := &model.Claims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tkn.SignedString([]byte(cfg.JWTSecret))
	require.NoError(t, err)

	md := map[string]string{"authorization": "Bearer " + token}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	exampleURL1 := "https://example-url-1.com"
	exampleURL2 := "https://example-url-2.com"
	r1, err := client.ShortenURL(ctx, pb.URLShortenRequest_builder{Url: proto.String(exampleURL1)}.Build())
	require.NoError(t, err)
	r2, err := client.ShortenURL(ctx, pb.URLShortenRequest_builder{Url: proto.String(exampleURL2)}.Build())
	require.NoError(t, err)

	id1 := strings.TrimPrefix(r1.GetResult(), cfg.ShortenAddress+"/")
	id2 := strings.TrimPrefix(r2.GetResult(), cfg.ShortenAddress+"/")

	exp, err := client.ExpandURL(ctx, pb.URLExpandRequest_builder{Id: proto.String(id1)}.Build())
	require.NoError(t, err)
	require.Equal(t, exampleURL1, exp.GetResult())

	exp, err = client.ExpandURL(ctx, pb.URLExpandRequest_builder{Id: proto.String(id2)}.Build())
	require.NoError(t, err)
	require.Equal(t, exampleURL2, exp.GetResult())

	list, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, list.GetUrl(), 2)
}
