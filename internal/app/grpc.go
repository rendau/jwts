package app

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang/protobuf/proto"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rendau/jwts/internal/errs"
	"github.com/rendau/jwts/pkg/proto/common"
)

func GrpcInterceptorTracing(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(
		tracer,
		otgrpc.IncludingSpans(func(parentSpanCtx opentracing.SpanContext, method string, req, resp any) bool {
			return parentSpanCtx != nil // only include spans if there is a parent span
		}),
		otgrpc.SpanDecorator(func(ctx context.Context, span opentracing.Span, method string, req, resp any, err error) {
			if err != nil {
				span.SetTag("error", true)
			}
		}),
	)
}

func GrpcInterceptorMetrics(namespace, service string) grpc.UnaryServerInterceptor {
	responseDurationSummary := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: "grpc",
		Name:      service + "_response_duration_seconds",
		Objectives: map[float64]float64{
			0.95: 0.001,
		},
		MaxAge: time.Minute,
	}, []string{
		"method",
	})

	requestCounter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "grpc",
		Name:      service + "_request_count",
	}, []string{
		"status",
		"method",
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		h, err := handler(ctx, req)

		st := "ok"
		if err != nil {
			st = "error"
		}

		responseDurationSummary.WithLabelValues(st, info.FullMethod).Observe(time.Since(start).Seconds())

		requestCounter.WithLabelValues(st, info.FullMethod).Inc()

		return h, err
	}
}

func GrpcInterceptorError() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		h, err := handler(ctx, req)
		if err == nil {
			return h, nil
		}

		var ei proto.Message
		errStr := err.Error()

		var errBase errs.Err
		if errors.As(err, &errBase) { // errs.Err
			ei = &common.ErrorRep{
				Code:    errBase.Error(),
				Message: errStr,
			}
		} else {
			var errFull errs.ErrFull
			if errors.As(err, &errFull) { // errs.ErrFull
				ei = &common.ErrorRep{
					Code:    errFull.Err.Error(),
					Message: errFull.Desc,
					Fields:  errFull.Fields,
				}
			}
		}
		if ei == nil {
			ei = &common.ErrorRep{
				Code:    errs.ServiceNA.Error(),
				Message: errStr,
			}

			if ctx.Err() == nil {
				slog.Info(
					"GRPC handler error",
					slog.String("error", errStr),
					slog.String("method", info.FullMethod),
				)
			}
		}

		st := status.New(codes.InvalidArgument, errStr)
		st, err = st.WithDetails(ei)
		if err != nil {
			slog.Error(
				"error while creating status with details",
				slog.String("error", errStr),
				slog.String("method", info.FullMethod),
			)
			st = status.New(codes.InvalidArgument, errStr)
		}

		return h, st.Err()
	}
}
