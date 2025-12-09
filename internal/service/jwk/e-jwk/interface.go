package e_jwk

import (
	"context"

	"github.com/rendau/jwts/internal/service/jwk/model"
)

type EJwkServiceI interface {
	FetchJwks(ctx context.Context) (*model.JwkSet, error)
}
