package jwk

import (
	"github.com/rendau/jwts/internal/service/jwk/model"
)

type Usecase struct {
	srv JwkServiceI
}

func New(
	srv JwkServiceI,
) *Usecase {
	return &Usecase{
		srv: srv,
	}
}

func (u *Usecase) GetSet() *model.JwkSet {
	return u.srv.GetSet()
}
