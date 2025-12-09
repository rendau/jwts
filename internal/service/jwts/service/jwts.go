package service

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
}

func New(kid string) *Service {
	return &Service{
		kid: kid,
	}
}

func (s *Service) SetKeys(privateKeyBytes []byte, publicKeyBytes []byte) error {
	var err error

	if len(privateKeyBytes) > 0 {
		s.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
		if err != nil {
			return err
		}
	}

	if len(publicKeyBytes) > 0 {
		s.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetPrivateKey() *rsa.PrivateKey {
	return s.privateKey
}

func (s *Service) GetPublicKey() *rsa.PublicKey {
	return s.publicKey
}

func (s *Service) GetKid() string {
	return s.kid
}
