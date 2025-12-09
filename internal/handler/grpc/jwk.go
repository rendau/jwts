package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	usecase "github.com/rendau/jwts/internal/usecase/jwk"
	"github.com/rendau/jwts/pkg/proto/jwts_v1"
)

type Jwk struct {
	jwts_v1.UnsafeJwkServer
	usecase *usecase.Usecase
}

func NewJwk(usecase *usecase.Usecase) *Jwk {
	return &Jwk{
		usecase: usecase,
	}
}

func (h *Jwk) Get(ctx context.Context, pars *emptypb.Empty) (*jwts_v1.JwkSet, error) {
	res := h.usecase.GetSet()

	keys := make([]*jwts_v1.JwkMain, len(res.Keys))
	for i, key := range res.Keys {
		keys[i] = &jwts_v1.JwkMain{
			Kty: key.Kty,
			E:   key.E,
			Kid: key.Kid,
			Alg: key.Alg,
			N:   key.N,
			Use: key.Use,
		}
	}

	return &jwts_v1.JwkSet{
		Keys: keys,
	}, nil
}
