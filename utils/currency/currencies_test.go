package currency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_register(t *testing.T) {
	t.Run("currency register should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			register("sgd", 2)
		})
	})
	t.Run("currency register ok, no panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			register("THB", 2)
		})
	})
}
