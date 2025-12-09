package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rendau/jwts/internal/service/jwt/model"
	usecase "github.com/rendau/jwts/internal/usecase/jwt"
	"github.com/rendau/jwts/pkg/proto/jwts_v1"
)

type Jwt struct {
	jwts_v1.UnsafeJwtServer
	usecase *usecase.Usecase
}

func NewJwt(usecase *usecase.Usecase) *Jwt {
	return &Jwt{
		usecase: usecase,
	}
}

func (h *Jwt) Create(ctx context.Context, req *jwts_v1.JwtCreateReq) (*jwts_v1.JwtCreateRep, error) {
	payload := map[string]any{}
	if req.Payload != nil && len(req.Payload) > 0 {
		err := json.Unmarshal(req.Payload, &payload)
		if err != nil {
			return nil, fmt.Errorf("json.Unmarshal payload: %w", err)
		}
	}

	res, err := h.usecase.Create(&model.JwtCreateReq{
		Sub:        req.Sub,
		ExpSeconds: req.ExpSeconds,
		Payload:    payload,
	})
	if err != nil {
		return nil, err
	}

	return &jwts_v1.JwtCreateRep{
		Token: res.Token,
	}, nil
}

func (h *Jwt) Validate(ctx context.Context, req *jwts_v1.JwtValidateReq) (*jwts_v1.JwtValidateRep, error) {
	res, err := h.usecase.Validate(&model.JwtValidateReq{
		Token: req.Token,
	})
	if err != nil {
		return nil, err
	}

	jsonClaims := make([]byte, 0)
	if res.Claims != nil {
		jsonClaims, err = json.Marshal(res.Claims)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal claims: %w", err)
		}
	}

	return &jwts_v1.JwtValidateRep{
		Valid:  res.Valid,
		Claims: jsonClaims,
	}, nil
}
