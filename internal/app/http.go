package app

import (
	"log/slog"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/cors"

	"github.com/rendau/jwts/internal/config"
)

func HttpMiddlewares(handler http.Handler) http.Handler {
	// add tracing-context middleware
	handler = func(h http.Handler) http.Handler {
		tracer := opentracing.GlobalTracer()

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// only include spans if there is a parent span
			wireContext, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			if err == nil {
				span := tracer.StartSpan(
					"HttpRequest",
					ext.RPCServerOption(wireContext),
				)
				defer span.Finish()

				r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
			}

			h.ServeHTTP(w, r)
		})
	}(handler)

	// add cors middleware
	if config.Conf.HttpCors {
		return cors.New(cors.Options{
			AllowOriginFunc: func(origin string) bool { return true },
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPut,
				http.MethodPost,
				http.MethodDelete,
			},
			AllowedHeaders: []string{
				"Accept",
				"Content-Type",
				"X-Requested-With",
				"Authorization",
			},
			AllowCredentials: true,
			MaxAge:           604800,
		}).Handler(handler)
	}

	// add recover middleware
	handler = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				// use always new err instance in defer
				if err := recover(); err != nil {
					slog.Error("HTTP handler recovered from panic", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}(handler)

	return handler
}
