package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/status"

	"github.com/rendau/jwts/internal/errs"
	"github.com/rendau/jwts/pkg/proto/common"
)

func checkErr(err error, r *http.Request, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}

	// grpc error
	st, ok := status.FromError(err)
	if ok {
		if len(st.Details()) > 0 {
			stDetail := st.Details()[0]
			errObj, ok := stDetail.(*common.ErrorRep)
			if ok {
				sendJson(&ErrorRep{
					ErrorCode: errObj.Code,
					Desc:      errObj.Message,
				}, w, http.StatusBadRequest)
				return true
			} else {
				sendJson(&ErrorRep{
					ErrorCode: errs.ServiceNA.Error(),
					Desc:      st.String(),
				}, w, http.StatusBadRequest)
				return true
			}
		} else {
			sendJson(&ErrorRep{
				ErrorCode: errs.ServiceNA.Error(),
				Desc:      st.String(),
			}, w, http.StatusBadRequest)
			return true
		}
	}

	// errs.Err
	var errBase errs.Err
	if errors.As(err, &errBase) {
		sendJson(&ErrorRep{
			ErrorCode: errBase.Error(),
			Desc:      err.Error(),
		}, w, http.StatusBadRequest)
		return true
	}

	// errs.ErrFull
	var errFull errs.ErrFull
	if errors.As(err, &errFull) {
		sendJson(&ErrorRep{
			ErrorCode: errFull.Err.Error(),
			Desc:      errFull.Desc,
		}, w, http.StatusBadRequest)
		return true
	}

	// log error
	if r.Context().Err() == nil {
		slog.Info(
			"Http handler error",
			"error", err.Error(),
			"method", r.Method,
			"path", r.URL.Path,
		)
	}

	// unknown error
	sendJson(&ErrorRep{
		ErrorCode: errs.ServiceNA.Error(),
		Desc:      err.Error(),
	}, w, http.StatusBadRequest)

	return true
}

func sendJson(obj any, w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(obj)
}
