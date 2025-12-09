package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/rendau/jwts/internal/config"
	"github.com/rendau/jwts/internal/constant"
	handlerGrpcP "github.com/rendau/jwts/internal/handler/grpc"
	handlerHttpP "github.com/rendau/jwts/internal/handler/http"
	"github.com/rendau/jwts/internal/service/jwk/e-jwk/kc"
	jwkServiceP "github.com/rendau/jwts/internal/service/jwk/service"
	jwtServiceP "github.com/rendau/jwts/internal/service/jwt/service"
	jwtsServiceP "github.com/rendau/jwts/internal/service/jwts/service"
	jwkUsecaseP "github.com/rendau/jwts/internal/usecase/jwk"
	jwtUsecaseP "github.com/rendau/jwts/internal/usecase/jwt"
	"github.com/rendau/jwts/pkg/proto/jwts_v1"
)

type App struct {
	jwkService *jwkServiceP.Service

	grpcServer *grpc.Server
	httpServer *http.Server

	// globalTracer
	globalTracerCloser io.Closer

	exitCode int
}

func (a *App) Init() {
	var err error

	var jwtsService *jwtsServiceP.Service

	var jwkHandlerGrpc *handlerGrpcP.Jwk
	var jwtHandlerGrpc *handlerGrpcP.Jwt

	// logger
	{
		if !config.Conf.Debug {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			slog.SetDefault(logger)
		}
	}

	// globalTracer
	{
		if config.Conf.WithTracing && config.Conf.JaegerAddress != "" {
			slog.Info("tracing enabled")
			_, a.globalTracerCloser, err = tracerInitGlobal(config.Conf.JaegerAddress, constant.ServiceName)
			errCheck(err, "tracerInitGlobal")
		}
	}

	// jwts
	{
		jwtsService = jwtsServiceP.New(config.Conf.Kid)
		if config.Conf.PublicPem != "" || config.Conf.PrivatePem != "" {
			var privatePem []byte
			var publicPem []byte

			if privatePemPath := config.Conf.PrivatePem; privatePemPath != "" {
				privatePem, err = os.ReadFile(privatePemPath)
				if err != nil {
					log.Fatal(err)
				}
			}

			if publicPemPath := config.Conf.PublicPem; publicPemPath != "" {
				publicPem, err = os.ReadFile(publicPemPath)
				if err != nil {
					log.Fatal(err)
				}
			}

			// set keys
			err = jwtsService.SetKeys(privatePem, publicPem)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// jwk
	{
		eJwkKC := kc.New(config.Conf.KcURL, config.Conf.KcRealmName)

		a.jwkService = jwkServiceP.New(jwtsService, eJwkKC)
		usecase := jwkUsecaseP.New(a.jwkService)
		jwkHandlerGrpc = handlerGrpcP.NewJwk(usecase)
	}

	// jwt
	{
		jwtService := jwtServiceP.New(jwtsService, config.Conf.DefaultIssuer)
		usecase := jwtUsecaseP.New(jwtService)
		jwtHandlerGrpc = handlerGrpcP.NewJwt(usecase)
	}

	// grpc server
	{
		interceptors := make([]grpc.UnaryServerInterceptor, 0, 3)

		// tracing
		interceptors = append(interceptors, GrpcInterceptorTracing(opentracing.GlobalTracer()))

		// metrics
		if config.Conf.WithMetrics {
			slog.Info("metrics enabled")
			interceptors = append(interceptors, GrpcInterceptorMetrics(config.Conf.Namespace, constant.ServiceName))
		}

		// error
		interceptors = append(interceptors, GrpcInterceptorError())

		// server
		a.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
			interceptors...,
		))

		// register grpc handlers
		jwts_v1.RegisterJwkServer(a.grpcServer, jwkHandlerGrpc)
		jwts_v1.RegisterJwtServer(a.grpcServer, jwtHandlerGrpc)

		// register grpc reflection
		reflection.Register(a.grpcServer)
	}

	// http server
	{
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024)),
			grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(
				opentracing.GlobalTracer(),
				otgrpc.IncludingSpans(func(parentSpanCtx opentracing.SpanContext, method string, req, resp any) bool {
					return parentSpanCtx != nil // only include spans if there is a parent span
				}),
			)),
		}
		conn, err := grpc.DialContext(context.Background(), "localhost:"+config.Conf.GrpcPort, opts...)
		errCheck(err, "grpc.DialContext")

		grpcJwkClient := jwts_v1.NewJwkClient(conn)
		grpcJwtClient := jwts_v1.NewJwtClient(conn)

		handlerHttp := handlerHttpP.New(grpcJwkClient, grpcJwtClient)

		mux := http.NewServeMux()

		// app handlers
		mux.HandleFunc("GET /jwk/set", handlerHttp.JwkGetSet)
		mux.HandleFunc("POST /jwt", handlerHttp.JwtCreate)
		mux.HandleFunc("PUT /jwt/validate", handlerHttp.JwtValidate)

		// metrics
		mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
			promhttp.Handler().ServeHTTP(w, r)
		})

		// healthcheck
		mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {})

		a.httpServer = &http.Server{
			Addr:              ":" + config.Conf.HttpPort,
			Handler:           HttpMiddlewares(mux),
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       time.Minute,
			MaxHeaderBytes:    300 * 1024,
		}
	}
}

func (a *App) PreStartHook() {
	slog.Info("PreStartHook")

	err := a.jwkService.CreateJwks()
	errCheck(err, "jwkService.CreateJwks")
}

func (a *App) Start() {
	slog.Info("Starting")

	// services
	{
	}

	// grpc server
	{
		lis, err := net.Listen("tcp", ":"+config.Conf.GrpcPort)
		errCheck(err, "fail to listen")
		go func() {
			err = a.grpcServer.Serve(lis)
			errCheck(err, "grpc-server stopped")
		}()
		slog.Info("grpc-server started " + lis.Addr().String())
	}

	// http server
	{
		go func() {
			err := a.httpServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCheck(err, "http-server stopped")
			}
		}()
		slog.Info("http-server started " + a.httpServer.Addr)
	}
}

func (a *App) Listen() {
	signalCtx, signalCtxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer signalCtxCancel()

	// wait signal
	<-signalCtx.Done()
}

func (a *App) Stop() {
	slog.Info("Shutting down...")

	// http server
	{
		ctx, ctxCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer ctxCancel()

		if err := a.httpServer.Shutdown(ctx); err != nil {
			slog.Error("http-server shutdown error", "error", err)
			a.exitCode = 1
		}
	}

	// grpc server
	{
		a.grpcServer.GracefulStop()
	}
}

func (a *App) WaitJobs() {
	slog.Info("waiting jobs")
}

func (a *App) Exit() {
	slog.Info("Exit")

	if a.globalTracerCloser != nil {
		_ = a.globalTracerCloser.Close()
	}

	os.Exit(a.exitCode)
}

func errCheck(err error, msg string) {
	if err != nil {
		if msg != "" {
			err = fmt.Errorf("%s: %w", msg, err)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}
