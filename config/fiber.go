package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func (cfg Config) NewFiber() *fiber.App {
	validate := validator.New()

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.AppName,
		BodyLimit:    10 * 1024 * 1024,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		StructValidator: &structValidator{
			validate: validate,
		},
		ErrorHandler: NewErrorHandler(),
	})

	return app
}

func NewErrorHandler() fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "internal server error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		} else if err != nil {
			message = err.Error()
		}

		log.Error().
			Err(err).
			Int("status_code", code).
			Str("path", c.Path()).
			Str("method", c.Method()).
			Msg("handler error")

		return c.Status(code).JSON(fiber.Map{
			"message": message,
			"data":    nil,
		})
	}
}
