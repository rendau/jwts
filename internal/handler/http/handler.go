package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/rendau/jwts/internal/errs"
	"github.com/rendau/jwts/pkg/proto/jwts_v1"
)

type Handler struct {
	jwkClient jwts_v1.JwkClient
	jwtClient jwts_v1.JwtClient
}

func New(jwkClient jwts_v1.JwkClient, jwtClient jwts_v1.JwtClient) *Handler {
	return &Handler{
		jwkClient: jwkClient,
		jwtClient: jwtClient,
	}
}

func (h *Handler) JwkGetSet(w http.ResponseWriter, r *http.Request) {
	grpcRepObj, err := h.jwkClient.Get(r.Context(), &emptypb.Empty{})
	if checkErr(err, r, w) {
		return
	}
	sendJson(grpcRepObj, w, http.StatusOK)
}

func (h *Handler) JwtCreate(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("fail to read request-body %w", err)
		checkErr(err, r, w)
		return
	}

	reqObj := map[string]any{}
	if err = json.Unmarshal(reqBody, &reqObj); err != nil {
		err = fmt.Errorf("fail to unmarshal request-body %w, body: %s", err, string(reqBody))
		checkErr(err, r, w)
		return
	}

	grpcReqObj := &jwts_v1.JwtCreateReq{
		Payload: reqBody,
	}

	var av any
	var ok bool

	if av, ok = reqObj["sub"]; ok {
		if grpcReqObj.Sub, ok = av.(string); !ok {
			sendJson(&ErrorRep{
				ErrorCode: errs.ServiceNA.Error(),
				Desc:      "sub must be string",
			}, w, http.StatusBadRequest)
			return
		}
	}

	if av, ok = reqObj["exp_seconds"]; ok { // 1296000
		expSecondsStr := fmt.Sprintf("%v", av)
		v, err := strconv.ParseFloat(expSecondsStr, 64)
		if err != nil {
			slog.Error("fail to parse exp_seconds", "exp_seconds_str", expSecondsStr, "original_exp_seconds", av, "err", err)
			sendJson(&ErrorRep{
				ErrorCode: errs.ServiceNA.Error(),
				Desc:      "fail to parse exp_seconds",
			}, w, http.StatusBadRequest)
			return
		}
		grpcReqObj.ExpSeconds = int64(v)
	}

	grpcRepObj, err := h.jwtClient.Create(r.Context(), grpcReqObj)
	if checkErr(err, r, w) {
		return
	}

	sendJson(grpcRepObj, w, http.StatusOK)
}

func (h *Handler) JwtValidate(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("fail to read request-body %w", err)
		checkErr(err, r, w)
		return
	}

	reqObj := &jwts_v1.JwtValidateReq{}
	if err = json.Unmarshal(reqBody, reqObj); err != nil {
		err = fmt.Errorf("fail to unmarshal request-body %w", err)
		checkErr(err, r, w)
		return
	}

	grpcRepObj, err := h.jwtClient.Validate(r.Context(), reqObj)
	if checkErr(err, r, w) {
		return
	}

	sendJson(&JwtValidateRep{
		Valid:  grpcRepObj.Valid,
		Claims: grpcRepObj.Claims,
	}, w, http.StatusOK)
}
