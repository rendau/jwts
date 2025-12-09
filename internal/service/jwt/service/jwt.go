package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/rendau/jwts/internal/constant"
	"github.com/rendau/jwts/internal/errs"
	"github.com/rendau/jwts/internal/service/jwt/model"
)

type Service struct {
	jwtsService   JwtsServiceI
	defaultIssuer string
}

func New(jwtsService JwtsServiceI, defaultIssuer string) *Service {
	return &Service{
		jwtsService:   jwtsService,
		defaultIssuer: defaultIssuer,
	}
}

func (s *Service) Create(obj *model.JwtCreateReq) (model.JwtCreateRep, error) {
	var err error

	result := model.JwtCreateRep{}

	if s.jwtsService.GetPrivateKey() == nil {
		return result, nil
	}

	claims := jwt.MapClaims{
		"iss": s.defaultIssuer, // issuer
	}

	for k, v := range obj.Payload {
		claims[k] = v
	}

	now := time.Now()

	if obj.ExpSeconds > 0 {
		claims["exp"] = now.Unix() + obj.ExpSeconds // expiration time
	}

	claims["iat"] = now.Add(-5 * time.Second).Unix() // issued at
	claims["sub"] = obj.Sub                          // subject (user id)

	t := jwt.NewWithClaims(jwt.GetSigningMethod(constant.JwtSigningMethod), claims)

	if s.jwtsService.GetKid() != "" {
		t.Header["kid"] = s.jwtsService.GetKid()
	}

	result.Token, err = t.SignedString(s.jwtsService.GetPrivateKey())
	if err != nil {
		return result, fmt.Errorf("t.SignedString: %w", err)
	}

	return result, nil
}

func (s *Service) Validate(obj *model.JwtValidateReq) (*model.JwtValidateRep, error) {
	result := &model.JwtValidateRep{}

	if s.jwtsService.GetPublicKey() == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(obj.Token, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errs.InvalidToken
		}
		return s.jwtsService.GetPublicKey(), nil
	})
	result.Valid = err == nil

	result.Claims = claims

	return result, nil
}
