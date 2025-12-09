package service

import "crypto/rsa"

type JwtsServiceI interface {
	GetPublicKey() *rsa.PublicKey
	GetKid() string
}
