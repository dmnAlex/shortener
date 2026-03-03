package tlsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/pkg/errors"
)

func GenerateSelfSignedCert() (string, string, error) {
	pKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", errors.Wrap(err, "generate key")
	}

	ca := x509.Certificate{
		SerialNumber: big.NewInt(2026),
		Subject: pkix.Name{
			Organization: []string{"Shortener Inc"},
			Country:      []string{"RU"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &pKey.PublicKey, pKey)
	if err != nil {
		return "", "", errors.Wrap(err, "create certificate")
	}

	certOut, err := os.CreateTemp("", "cert-*.pem")
	if err != nil {
		return "", "", errors.Wrap(err, "create temporary cert file")
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return "", "", errors.Wrap(err, "encode cert")
	}

	keyOut, err := os.CreateTemp("", "key-*.pem")
	if err != nil {
		return "", "", errors.Wrap(err, "create temporary key file")
	}
	defer keyOut.Close()

	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pKey)}); err != nil {
		return "", "", errors.Wrap(err, "encode key")
	}

	return certOut.Name(), keyOut.Name(), nil
}
