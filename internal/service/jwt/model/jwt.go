package model

type JwtCreateReq struct {
	Sub        string
	ExpSeconds int64
	Payload    map[string]any
}

type JwtCreateRep struct {
	Token string
}

type JwtValidateReq struct {
	Token string
}

type JwtValidateRep struct {
	Valid  bool
	Claims map[string]any
}
