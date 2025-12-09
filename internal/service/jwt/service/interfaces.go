package service

import "crypto/rsa"

type JwtsServiceI interface {
	GetPrivateKey() *rsa.PrivateKey
	GetPublicKey() *rsa.PublicKey
	GetKid() string
}
