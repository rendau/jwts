package tests

import (
	_ "embed"
	"log"
	"os"
	"testing"

	dopLoggerZap "github.com/rendau/dop/adapters/logger/zap"
	"github.com/rendau/jwts/internal/domain/core"
	"github.com/spf13/viper"
)

//go:embed private.pem
var privatePem []byte

//go:embed public.pem
var publicPem []byte

func TestMain(m *testing.M) {
	var err error

	viper.SetConfigFile("test_conf.yml")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	app.lg = dopLoggerZap.New("info", true)

	app.core = core.New(app.lg)

	err = app.core.SetKeys(privatePem, publicPem, "key1")
	if err != nil {
		log.Fatal(err)
	}

	// Start tests
	code := m.Run()

	os.Exit(code)
}
