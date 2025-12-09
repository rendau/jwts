package config

import (
	"github.com/caarlos0/env/v9"
	_ "github.com/joho/godotenv/autoload"
)

var Conf = struct {
	Namespace     string `env:"NAMESPACE" envDefault:"example.com"`
	Debug         bool   `env:"DEBUG" envDefault:"false"`
	GrpcPort      string `env:"GRPC_PORT" envDefault:"5050"`
	HttpPort      string `env:"HTTP_PORT" envDefault:"80"`
	HttpCors      bool   `env:"HTTP_CORS" envDefault:"false"`
	WithMetrics   bool   `env:"WITH_METRICS" envDefault:"false"`
	WithTracing   bool   `env:"WITH_TRACING" envDefault:"false"`
	JaegerAddress string `env:"JAEGER_ADDRESS"`
	Kid           string `env:"KID"`
	DefaultIssuer string `env:"DEFAULT_ISSUER"`
	PrivatePem    string `env:"PRIVATE_PEM"`
	PublicPem     string `env:"PUBLIC_PEM"`
	KcURL         string `env:"KC_URL"`
	KcRealmName   string `env:"KC_REALM_NAME"`
}{}

func init() {
	if err := env.Parse(&Conf); err != nil {
		panic(err)
	}
}
