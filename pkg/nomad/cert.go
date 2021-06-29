package nomad

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

type Certificate struct {
	Pem []byte
	Key []byte
}

type Issuer struct {
	CA *Certificate
}

func NewIssuer() (*Issuer, error) {
	issuer := new(Issuer)
	ca, err := createCA()
	if err != nil {
		return nil, err
	}
	issuer.CA = ca

	return issuer, nil
}

func createCA() (*Certificate, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization:  []string{"GS-Labs"},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{"Saint-Petersburg"},
			StreetAddress: []string{"Helsingforskaya"},
			PostalCode:    []string{"4"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	return &Certificate{Pem: caPEM.Bytes(), Key: caPrivKeyPEM.Bytes()}, nil
}

func (i *Issuer) IssueCA() (*Certificate, error) {
	return nil, nil
}
