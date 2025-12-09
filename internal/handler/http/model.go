package http

import (
	"encoding/json"
)

type ErrorRep struct {
	ErrorCode string `json:"error_code"`
	Desc      string `json:"desc"`
}

type JwtValidateRep struct {
	Valid  bool            `json:"valid"`
	Claims json.RawMessage `json:"claims"`
}
