package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/dmnAlex/shortener/internal/audit"
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/service"
	pb "github.com/dmnAlex/shortener/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type shortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	service  service.URLService
	config   *config.Config
	auditMgr *audit.AuditManager
}

func NewShortenerServer(svc service.URLService, cfg *config.Config, auditMgr *audit.AuditManager) pb.ShortenerServiceServer {
	return &shortenerServer{
		service:  svc,
		config:   cfg,
		auditMgr: auditMgr,
	}
}

func (s *shortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	caller := getCaller(ctx)
	if caller == nil {
		return nil, status.Error(codes.Unauthenticated, errx.ErrUnauthorized.Error())
	}
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, errx.ErrBadRequest.Error())
	}

	shortID, err := s.service.Shorten(caller.UserID, req.GetUrl())
	shortURL := fmt.Sprintf("%s/%s", s.config.ShortenAddress, shortID)

	if err != nil && !errors.Is(err, errx.ErrConflict) {
		logger.Log.Error(err.Error())
		return nil, status.Error(codes.Internal, errx.ErrInternalError.Error())
	}

	s.notifyAudit(ctx, model.AuditActionShorten, req.GetUrl())

	res := pb.URLShortenResponse_builder{
		Result: proto.String(shortURL),
	}

	return res.Build(), nil
}

func (s *shortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, errx.ErrBadRequest.Error())
	}

	originalURL, err := s.service.Expand(req.GetId())
	if err != nil {
		if errors.Is(err, errx.ErrNotFound) {
			return nil, status.Error(codes.NotFound, errx.ErrNotFound.Error())
		}
		if errors.Is(err, errx.ErrGone) {
			return nil, status.Error(codes.FailedPrecondition, errx.ErrGone.Error())
		}
		logger.Log.Error(err.Error())
		return nil, status.Error(codes.Internal, errx.ErrInternalError.Error())
	}

	s.notifyAudit(ctx, model.AuditActionFollow, originalURL)

	res := pb.URLExpandResponse_builder{
		Result: proto.String(originalURL),
	}

	return res.Build(), nil
}

func (s *shortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	caller := getCaller(ctx)
	if caller == nil {
		return nil, status.Error(codes.Unauthenticated, errx.ErrUnauthorized.Error())
	}

	urls, err := s.service.UserURLs(caller.UserID)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, status.Error(codes.Internal, errx.ErrInternalError.Error())
	}

	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", s.config.ShortenAddress, urls[i].ShortURL)
	}

	pbData := make([]*pb.URLData, len(urls))
	for i, u := range urls {
		pbData[i] = pb.URLData_builder{
			ShortUrl:    proto.String(u.ShortURL),
			OriginalUrl: proto.String(u.OriginalURL),
		}.Build()
	}

	res := pb.UserURLsResponse_builder{
		Url: pbData,
	}

	return res.Build(), nil
}

func getCaller(ctx context.Context) *model.Caller {
	if caller, ok := ctx.Value(model.CallerKey).(*model.Caller); ok {
		return caller
	}

	return nil
}

func (s *shortenerServer) notifyAudit(ctx context.Context, action model.AuditAction, url string) {
	if s.auditMgr == nil {
		return
	}
	caller := getCaller(ctx)
	if caller == nil {
		return
	}

	s.auditMgr.Notify(model.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    caller.UserID,
		URL:       url,
	})
}
