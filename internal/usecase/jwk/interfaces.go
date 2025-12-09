package jwk

import "github.com/rendau/jwts/internal/service/jwk/model"

type JwkServiceI interface {
	GetSet() *model.JwkSet
}
