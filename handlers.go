package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

type user struct {
	Name  string `json:"name" form:"name" query:"name"`
	Email string `json:"email" form:"email" query:"email"`
}

type userDTO struct {
	Name    string
	Email   string
	IsAdmin bool
}

// http localhost:3000/user name=ronen email=konfortes@gmail.com
func createUser(c echo.Context) error {
	u := new(user)
	if err := c.Bind(u); err != nil {
		c.Logger().Errorf("failed to create user: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Request")
	}
	// To avoid security flaws try to avoid passing binded structs directly to other methods
	// if these structs contain fields that should not be bindable.
	user := userDTO{
		Name:    u.Name,
		Email:   u.Email,
		IsAdmin: false, // because you could accidentally expose fields that should not be bind
	}
	// executeSomeBusinessLogic(user)

	return c.JSON(http.StatusOK, user)
}
