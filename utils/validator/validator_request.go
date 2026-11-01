package validator

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/rs/zerolog/log"
)

type Validator struct {
	Validator  *validator.Validate
	Translator ut.Translator
}

func NewValidator() *Validator {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, found := uni.GetTranslator("en")
	if !found {
		log.Error().
			Str("source", "utils.validator.NewValidator").
			Msg("Translator not found")
	}

	validate := validator.New()
	_ = entranslations.RegisterDefaultTranslations(validate, trans)

	return &Validator{
		Validator:  validate,
		Translator: trans,
	}
}

func (v *Validator) Validate(i any) error {
	err := v.Validator.Struct(i)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok {
			for _, e := range validationErrors {
				return errors.New(e.Translate(v.Translator))
			}
		}
		return err
	}
	return nil
}
