package jwt

import (
	"fmt"

	"github.com/rendau/jwts/internal/service/jwt/model"
)

type Usecase struct {
	srv JwtServiceI
}

func New(
	srv JwtServiceI,
) *Usecase {
	return &Usecase{
		srv: srv,
	}
}

func (u *Usecase) Create(obj *model.JwtCreateReq) (model.JwtCreateRep, error) {
	result, err := u.srv.Create(obj)
	if err != nil {
		err = fmt.Errorf("srv.Create: %w", err)
	}

	return result, err
}

func (u *Usecase) Validate(obj *model.JwtValidateReq) (*model.JwtValidateRep, error) {
	result, err := u.srv.Validate(obj)
	if err != nil {
		err = fmt.Errorf("srv.Validate: %w", err)
	}

	return result, err
}
