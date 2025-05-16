package apierr

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrStatusCodeJSON(t *testing.T) {
	t.Run("bad request error", func(t *testing.T) {
		got := BadRequest("error")
		assert.Equal(t, "error", got.Error())
		assert.Equal(t, http.StatusBadRequest, got.HTTPStatusCode())
	})

	t.Run("internal error", func(t *testing.T) {
		got := InternalServer("error")
		assert.Equal(t, "error", got.Error())
		assert.Equal(t, http.StatusInternalServerError, got.HTTPStatusCode())
	})

	t.Run("unauthorized error", func(t *testing.T) {
		got := Unauthorized()
		assert.Equal(t, "unauthenticated", got.Error())
		assert.Equal(t, http.StatusUnauthorized, got.HTTPStatusCode())
	})

	t.Run("unprocessable error", func(t *testing.T) {
		got := Unprocessable("error11")
		assert.Equal(t, "error11", got.Error())
		assert.Equal(t, http.StatusUnprocessableEntity, got.HTTPStatusCode())
	})
}
