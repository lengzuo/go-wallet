package currency

import (
	"errors"
	"fmt"
	"strings"
)

type Value struct {
	Code    string
	Precise int
}

var (
	ErrUnSupportedCurrency = errors.New("unsupported currency")
	ErrEmptyCurrency       = errors.New("currency is mandatory")
)

var currencies = make(map[string]Value)

var (
	SGD = register("SGD", 2)
	JPY = register("JPY", 0)
)

func register(code string, precise int) Value {
	code = strings.ToUpper(code)
	if _, ok := currencies[code]; ok {
		panic(fmt.Sprintf("duplicated currencies found: %s", code))
	}
	currencies[code] = Value{Precise: precise, Code: code}
	return currencies[code]
}

func Supported(code string) error {
	if strings.TrimSpace(code) == "" {
		return ErrEmptyCurrency
	}
	for k := range currencies {
		if k == code {
			return nil
		}
	}
	return ErrUnSupportedCurrency
}
