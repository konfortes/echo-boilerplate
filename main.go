package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type serverConfig struct {
	port     string
	env      string
	logLevel string
	admin    struct {
		user     string
		password string
	}
}

var (
	config serverConfig
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

	config.port = getEnvOr("PORT", "3000")
	config.env = getEnvOr("GO_ENV", "development")
	config.logLevel = getEnvOr("LOG_LEVEL", "info")
	config.admin.user = getEnvOr("ADMIN_USER", "")
	config.admin.password = getEnvOr("ADMIN_PASSWORD", "")
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
	// TODO: jaeger and prometheus
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: generateSkipper([]string{"/health"}),
	}))
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: generateSkipper([]string{"/health"}),
	}))
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

	e.Logger.Fatal(e.Start(":" + config.port))
}
