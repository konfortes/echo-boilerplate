package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type serverConfig struct {
	port     string
	env      string
	logLevel string
	appName  string
	admin    struct {
		user     string
		password string
	}
	tracingEnabled string
}

var (
	config        serverConfig
	shutdownHooks []func()
)

func getEnvOr(name string, defaultValue string) string {
	val, found := os.LookupEnv(name)

	if !found {
		val = defaultValue
	}

	return val
}

func init() {
	_ = godotenv.Load()

	// TODO: use Viper
	config.port = getEnvOr("PORT", "3000")
	config.env = getEnvOr("GO_ENV", "development")
	config.appName = getEnvOr("APP_NAME", "my_app")
	config.logLevel = getEnvOr("LOG_LEVEL", "info")
	config.admin.user = getEnvOr("ADMIN_USER", "")
	config.admin.password = getEnvOr("ADMIN_PASSWORD", "")
	config.tracingEnabled = getEnvOr("TRACING_ENABLED", "false")
}

func generateSkipper(skipForPaths []string) func(echo.Context) bool {
	return func(c echo.Context) bool {
		for _, path := range skipForPaths {
			if path == c.Path() {
				return true
			}
		}
		return false
	}
}

func setMiddlewares(e *echo.Echo) {
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: generateSkipper([]string{"/health"}),
	}))
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: generateSkipper([]string{"/health"}),
	}))

	p := prometheus.NewPrometheus(config.appName, nil)
	p.Use(e)

	if config.tracingEnabled == "true" {
		// Use JAEGER_AGENT_HOST and JAEGER_AGENT_PORT to configure agent (or collector)
		c := jaegertracing.New(e, nil)
		e.Server.RegisterOnShutdown(func() {
			if err := c.Close; err != nil {
				e.Logger.Error(err)
			}
		})

	}
}

func newEcho() *echo.Echo {
	e := echo.New()

	if config.env == "production" {
		e.HideBanner = true
	}

	// TODO: handle log level
	// e.Logger.SetLevel()
	return e
}

func setAdminRoutes(e *echo.Echo) {
	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == config.admin.user && password == config.admin.password {
			return true, nil
		}
		return false, nil
	}))

	// base64(admin:admin)=YWRtaW46YWRtaW4=
	// curl -H'Authorization: Basic YWRtaW46YWRtaW4=' localhost:3000/admin/routes
	adminGroup.GET("/routes", func(c echo.Context) error {
		routes := e.Routes()
		return c.JSON(http.StatusOK, routes)
	})
}

func main() {
	e := newEcho()

	setMiddlewares(e)

	e.GET("/", hello)
	e.GET("/health", health)
	e.POST("/user", createUser)

	setAdminRoutes(e)

	listenAddress := ":" + config.port

	go func() {
		if err := e.Start(listenAddress); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	// using grace or graceful libraries does not invoke Echo#Start, therefore skips Echo#ConfigureServer
	gracefulShutdown(e, 5*time.Second)
}

func gracefulShutdown(e *echo.Echo, graceDuration time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), graceDuration)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
