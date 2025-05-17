package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/stretchr/testify/assert"
)

type mockParams2 struct {
	Field1 string `json:"field1"`
	Field2 int    `schema:"field2"`
}

func (r mockParams2) Validate() apierr.JSON {
	return apierr.BadRequest("err")
}

type mockResponse2 struct {
	Status string `json:"status"`
}

func (m mockResponse2) StatusCode() int {
	return 200
}

type mockParams struct {
	Field1 string `json:"field1"`
	Field2 int    `schema:"field2"`
}

func (r mockParams) Validate() apierr.JSON {
	return nil
}

type mockResponse struct {
	Status string `json:"status"`
}

func (m mockResponse) StatusCode() int {
	return 200
}
func TestHandle(t *testing.T) {
	t.Run("successful requesa in handle", func(t *testing.T) {
		mockFunc := func(ctx context.Context, in mockParams) (*mockResponse, apierr.JSON) {
			assert.Equal(t, "value1", in.Field1)
			assert.Equal(t, 123, in.Field2)
			return &mockResponse{Status: "success"}, nil
		}
		reqBody, _ := json.Marshal(map[string]string{"field1": "value1"})
		req, _ := http.NewRequest("POST", "/test?field2=123", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler := Handle(mockFunc)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var responseBody mockResponse
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, "success", responseBody.Status)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("failed bad request in handle", func(t *testing.T) {
		mockFunc := func(ctx context.Context, in mockParams2) (*mockResponse2, apierr.JSON) {
			t.Fatal("shouldn't called")
			return nil, nil
		}
		a := new(mockParams2)
		reqBody, _ := json.Marshal(a)
		req, _ := http.NewRequest("POST", "/test?field2=123", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler := Handle(mockFunc)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody map[string]string
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.EqualValues(t, map[string]string{
			"code":    "VALIDATION_FAILED",
			"message": "err",
		}, responseBody)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("failed invalid json in handle", func(t *testing.T) {
		mockFunc := func(ctx context.Context, in mockParams) (*mockResponse, apierr.JSON) {
			t.Fatal("shouldn't called")
			return nil, nil
		}
		reqBody := bytes.NewReader([]byte(`{"field1": "value1",}`))
		req, _ := http.NewRequest("POST", "/test?field2=123", reqBody)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler := Handle(mockFunc)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody map[string]string
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.EqualValues(t, map[string]string{
			"code":    "VALIDATION_FAILED",
			"message": "invalid json",
		}, responseBody)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("failed invalid json in handle", func(t *testing.T) {
		mockFunc := func(ctx context.Context, in mockParams) (*mockResponse, apierr.JSON) {
			t.Fatal("shouldn't called")
			return nil, nil
		}
		req, _ := http.NewRequest("GET", "/test?field2=abc", nil)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler := Handle(mockFunc)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var responseBody map[string]string
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.EqualValues(t, map[string]string{
			"code":    "VALIDATION_FAILED",
			"message": "invalid querystring",
		}, responseBody)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("failed unauthenticated in handle", func(t *testing.T) {
		// Mock RestfulFunc that returns an apierr.JSON error
		mockFuncWithError := func(ctx context.Context, in mockParams) (*mockResponse, apierr.JSON) {
			return nil, apierr.Unauthenticated()
		}
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler := Handle(mockFuncWithError)
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var responseBody map[string]string
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.EqualValues(t, map[string]string{
			"code":    "UNAUTHENTICATED",
			"message": "unauthenticated",
		}, responseBody)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})
}
