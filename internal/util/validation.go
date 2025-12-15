package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"gorm.io/gorm"
)

var (
	once                sync.Once
	UniversalTranslator *ut.UniversalTranslator
)

func AddValidation(db *gorm.DB) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	once.Do(func() {
		en := en.New()
		UniversalTranslator = ut.New(en, en)
	})

	trans, _ := UniversalTranslator.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(v, trans)

	// register custom validation
	v.RegisterValidation("unique_db", uniqueValidator(db))
	v.RegisterTranslation("unique_db", trans, func(ut ut.Translator) error {
		return ut.Add("unique_db", "{0} already taken", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("unique_db", fe.Field())
		return t
	})

	// register password validator
	v.RegisterValidation("validate_password", passwordValidator())
	v.RegisterTranslation("validate_password", trans, func(ut ut.Translator) error {
		return ut.Add("validate_password", "password must be at least 8 characters long and contain uppercase, lowercase, number, and special character", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("validate_password", fe.Field())
		return t
	})

	// register currency validator
	v.RegisterValidation("currency", currencyValidator())
	v.RegisterTranslation("currency", trans, func(ut ut.Translator) error {
		return ut.Add("currency", "currency must be a supported currency", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("currency", fe.Field())
		return t
	})

	// set tag name from json tag
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// register additional custom validators (global)
	RegisterCustomValidators(v) // <=== added
}

// uniqueValidator validate that new value is not exist on db / unique
// example usage: unique_db=users:email
func uniqueValidator(db *gorm.DB) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		param := fl.Param()
		params := strings.Split(param, ":")
		if len(params) != 2 {
			return false
		}

		tableName := params[0]
		columnName := params[1]
		fieldValue := fl.Field().String()
		var count int64

		query := fmt.Sprintf("%s = ?", columnName)
		err := db.Table(tableName).Where(query, fieldValue).Where("deleted_at IS NULL").Count(&count).Error
		if err != nil {
			return false
		}

		return count == 0
	}
}

// passwordValidator validates that password meets security requirements
func passwordValidator() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return IsValidPassword(password)
	}
}

// currencyValidator validates if currency is supported
func currencyValidator() func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		currency := fl.Field().String()
		if currency == "" {
			return true // Allow empty currency (will use default)
		}

		// Get supported currencies from global validator
		supportedCurrencies := GetSupportedCurrencies()

		for _, supported := range supportedCurrencies {
			if strings.EqualFold(currency, supported) {
				return true
			}
		}

		return false
	}
}

func isUpperCase(c rune) bool {
	return c >= 'A' && c <= 'Z'
}

func isLowerCase(c rune) bool {
	return c >= 'a' && c <= 'z'
}

func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func isPunctOrSymbol(c rune) bool {
	return (c >= 33 && c <= 47) || (c >= 58 && c <= 64) ||
		(c >= 91 && c <= 96) || (c >= 123 && c <= 126)
}

type CurrencyValidator struct {
	supportedCurrencies []string
	mu                  sync.RWMutex
}

func NewCurrencyValidator() *CurrencyValidator {
	return &CurrencyValidator{
		supportedCurrencies: []string{"USD", "EUR", "GBP", "SGD", "MYR", "IDR"},
	}
}

func (cv *CurrencyValidator) SetSupportedCurrencies(currencies []string) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.supportedCurrencies = currencies
}

func (cv *CurrencyValidator) GetSupportedCurrencies() []string {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return append([]string{}, cv.supportedCurrencies...)
}

func (cv *CurrencyValidator) ValidateCurrency(currency string) error {
	if currency == "" {
		return nil // Allow empty currency (will use default)
	}

	cv.mu.RLock()
	defer cv.mu.RUnlock()

	for _, supported := range cv.supportedCurrencies {
		if strings.EqualFold(currency, supported) {
			return nil
		}
	}

	return fmt.Errorf("currency '%s' is not supported. Supported currencies: %s",
		currency, strings.Join(cv.supportedCurrencies, ", "))
}

type CustomValidator struct {
	validate          *validator.Validate
	currencyValidator *CurrencyValidator
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	cv := &CustomValidator{
		validate:          v,
		currencyValidator: NewCurrencyValidator(),
	}

	cv.registerCustomValidators()

	return cv
}

func (cv *CustomValidator) SetSupportedCurrencies(currencies []string) {
	cv.currencyValidator.SetSupportedCurrencies(currencies)
}

func (cv *CustomValidator) registerCustomValidators() {
	cv.validate.RegisterValidation("currency", cv.validateCurrency)
	// Optional: also expose oneof_ci on the custom validator instance
	_ = cv.validate.RegisterValidation("oneof_ci", func(fl validator.FieldLevel) bool { // <=== added
		raw := fl.Field().String()
		if strings.TrimSpace(raw) == "" {
			return false
		}
		opts := strings.Fields(fl.Param())
		val := strings.ToUpper(strings.TrimSpace(raw))
		for _, o := range opts {
			if strings.ToUpper(o) == val {
				return true
			}
		}
		return false
	})
}

func (cv *CustomValidator) validateCurrency(fl validator.FieldLevel) bool {
	currency := fl.Field().String()
	return cv.currencyValidator.ValidateCurrency(currency) == nil
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validate.Struct(i)
}

func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validate.Var(field, tag)
}

func (cv *CustomValidator) GetValidationErrors(err error) []string {
	var errors []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			switch tag {
			case "currency":
				supported := cv.currencyValidator.GetSupportedCurrencies()
				errors = append(errors, fmt.Sprintf("field '%s' must be a supported currency: %s",
					field, strings.Join(supported, ", ")))
			case "required":
				errors = append(errors, fmt.Sprintf("field '%s' is required", field))
			case "max":
				errors = append(errors, fmt.Sprintf("field '%s' must not exceed %s characters", field, param))
			case "uuid":
				errors = append(errors, fmt.Sprintf("field '%s' must be a valid UUID", field))
			default:
				errors = append(errors, fmt.Sprintf("field '%s' failed validation: %s", field, tag))
			}
		}
	}

	return errors
}

func (cv *CustomValidator) IsValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}

var globalValidator *CustomValidator
var validatorOnce sync.Once

func GetGlobalValidator() *CustomValidator {
	validatorOnce.Do(func() {
		globalValidator = NewCustomValidator()
	})
	return globalValidator
}

func SetGlobalSupportedCurrencies(currencies []string) {
	GetGlobalValidator().SetSupportedCurrencies(currencies)
}

func ValidateStruct(i interface{}) error {
	return GetGlobalValidator().Validate(i)
}

func ValidateStructWithErrors(i interface{}) []string {
	v := GetGlobalValidator()
	err := v.Validate(i)
	if err != nil {
		return v.GetValidationErrors(err)
	}
	return nil
}

func ValidateCurrency(currency string) error {
	return GetGlobalValidator().currencyValidator.ValidateCurrency(currency)
}

func GetSupportedCurrencies() []string {
	return GetGlobalValidator().currencyValidator.GetSupportedCurrencies()
}

func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("alphanumdash", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		re := regexp.MustCompile(`^[A-Z0-9\-]+$`)
		return re.MatchString(s)
	})

	// oneof_ci: case-insensitive version of oneof (space-separated options)
	_ = v.RegisterValidation("oneof_ci", func(fl validator.FieldLevel) bool { // <=== added
		raw := fl.Field().String()
		if strings.TrimSpace(raw) == "" {
			return false
		}
		opts := strings.Fields(fl.Param()) // e.g. "DOCTOR NURSE RECEPTIONIST BOD ADMIN"
		val := strings.ToUpper(strings.TrimSpace(raw))
		for _, o := range opts {
			if strings.ToUpper(o) == val {
				return true
			}
		}
		return false
	})
}
