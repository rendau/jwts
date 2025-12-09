package service

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"reflect"

	e_jwk "github.com/rendau/jwts/internal/service/jwk/e-jwk"
	"github.com/rendau/jwts/internal/service/jwk/model"
)

type Service struct {
	jwks *model.JwkSet

	jwtsService JwtsServiceI
	eJwkService e_jwk.EJwkServiceI
}

func New(jwtsService JwtsServiceI, eJwkService e_jwk.EJwkServiceI) *Service {
	return &Service{
		jwtsService: jwtsService,
		eJwkService: eJwkService,
	}
}

func (s *Service) CreateJwks() error {
	var err error

	s.jwks, err = s.createJwks()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) createJwks() (*model.JwkSet, error) {
	result := &model.JwkSet{}

	if !reflect.ValueOf(s.eJwkService).IsNil() {
		eKeys, err := s.eJwkService.FetchJwks(context.Background())
		if err != nil {
			return nil, err
		}

		for _, v := range eKeys.Keys {
			result.Keys = append(result.Keys, v)
		}
	}

	if s.jwtsService.GetPublicKey() == nil {
		if len(result.Keys) != 0 {
			return result, nil
		}
		return nil, nil
	}

	eBytes := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(eBytes, uint32(s.jwtsService.GetPublicKey().E))

	key := model.JwkMain{
		Kty: "RSA",
		E:   base64.RawURLEncoding.EncodeToString(eBytes[:3]),
		Kid: s.jwtsService.GetKid(),
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(s.jwtsService.GetPublicKey().N.Bytes()),
		Use: "sig",
	}

	result.Keys = append(result.Keys, &key)

	return result, nil
}

func (s *Service) GetSet() *model.JwkSet {
	return s.jwks
}
