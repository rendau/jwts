package model

type JwkMain struct {
	Kty string
	E   string
	Kid string
	Alg string
	N   string
	Use string
}

type JwkSet struct {
	Keys []*JwkMain
}
