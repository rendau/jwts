package jwt

import "github.com/rendau/jwts/internal/service/jwt/model"

type JwtServiceI interface {
	Create(obj *model.JwtCreateReq) (model.JwtCreateRep, error)
	Validate(obj *model.JwtValidateReq) (*model.JwtValidateRep, error)
}
