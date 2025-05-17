package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/lengzuo/fundflow/pkg/log"
	"github.com/lengzuo/fundflow/utils"
	"github.com/redis/go-redis/v9"
)

const (
	idempotencyHeader = "X-Idempotency-Key"
	initValue         = "pending"
)

type MiddlewareFunc func(next http.Handler) http.Handler

// responseWriter is a wrapper around http.ResponseWriter that captures the status code and body.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// cachedResponse is a struct to store the cached HTTP response.
type cachedResponse struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
	BodyHash   string      `json:"body_hash"`
}

func hashedAndRestoreBody(r *http.Request) (string, error) {
	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(r.Context(), "failed to read request body")
		return "", err
	}
	// Restore the request body so the next handler can read it
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Calculate the hash of the request body
	hasher := fnv.New128a()
	hasher.Write(bodyBytes)
	reqBodyHash := hasher.Sum(nil)
	return fmt.Sprintf("%x", reqBodyHash), nil
}

func renderInternalErr(w http.ResponseWriter, r *http.Request) {
	apiErr := apierr.InternalServer("please try again")
	render.Status(r, apiErr.HTTPStatusCode())
	render.JSON(w, r, apiErr)
}

func getAPIName(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	path := parsedURL.Path
	if path == "" || path == "/" {
		return ""
	}
	trimmedPath := strings.TrimRight(path, "/")
	lastSlashIndex := strings.LastIndex(trimmedPath, "/")
	if lastSlashIndex == -1 {
		return trimmedPath
	}
	return trimmedPath[lastSlashIndex+1:]
}

func Idempotency(redisClient *redis.Client) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Only apply idempotency to POST and PUT requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}

			idempotencyKey := r.Header.Get(idempotencyHeader)
			if idempotencyKey == "" {
				// No idempotency key provided, proceed without idempotency
				next.ServeHTTP(w, r)
				return
			}

			// Read the request body
			reqBodyHash, err := hashedAndRestoreBody(r)
			if err != nil {
				renderInternalErr(w, r)
				return
			}
			username := r.Header.Get("Authorization")
			apiName := getAPIName(r.URL.Path)
			// The key should be a combination of idempotency key and username
			redisKey := username + idempotencyKey + apiName
			result, err := redisClient.Get(ctx, redisKey).Result()
			// Process with idempontency when err is redis.Nil
			if err == redis.Nil {
				// Start idempotency key as status = "pending"
				err = redisClient.Set(ctx, redisKey, initValue, utils.APIRequestTimeout).Err()
				if err != nil {
					log.Error(ctx, "failed in set idempotency key with err: %s", err)
					renderInternalErr(w, r)
					return
				}

				// Wrap the response writer to capture the response
				wrappedWriter := newResponseWriter(w)

				next.ServeHTTP(wrappedWriter, r)

				// After processing, store the actual response in Redis
				responseToCache := cachedResponse{
					StatusCode: wrappedWriter.statusCode,
					Headers:    wrappedWriter.Header(),
					Body:       wrappedWriter.body.Bytes(),
					BodyHash:   reqBodyHash,
				}

				responseBytes, err := json.Marshal(responseToCache)
				if err != nil {
					log.Error(ctx, "failed in marshal response in idempotency with err: %s", err)
					renderInternalErr(w, r)
					return
				}

				// Store the actual response with the 1-hour TTL
				err = redisClient.Set(ctx, redisKey, responseBytes, 1*time.Hour).Err()
				if err != nil {
					// In any case we failed to set after the server have response. We need to tell user not to retry with idempotency
					// Allow caller to recover the tranaction from get
					log.Error(ctx, "failed in update idempotency response with err: %s", err)
					apiErr := apierr.ServiceUnavailable("idempotency service down")
					render.Status(r, apiErr.HTTPStatusCode())
					render.JSON(w, r, apiErr)
					return
				}
				return

			}

			// Handle err if redis not available
			if err != nil {
				// Handle other Redis errors, maybe log and proceed without idempotency
				log.Error(ctx, "failed in get redis err: %s", err)
				apiErr := apierr.ServiceUnavailable("idempotency service down")
				render.Status(r, apiErr.HTTPStatusCode())
				render.JSON(w, r, apiErr)
				return
			}

			// Idempontency key exists

			if result == initValue {
				apiErr := apierr.Conflict("try again with exponential")
				render.Status(r, apiErr.HTTPStatusCode())
				render.JSON(w, r, apiErr)
				return
			}
			// Key exists, this is a duplicate request
			// Retrieve the cached response and write it to the client
			var cachedResponse cachedResponse
			err = json.Unmarshal([]byte(result), &cachedResponse)
			if err != nil {
				log.Error(ctx, "failed in unmarhsal err: %s", err)
				renderInternalErr(w, r)
				return
			}
			if cachedResponse.BodyHash != reqBodyHash {
				log.Error(ctx, "idempotency key: %s reuse doesnt match", idempotencyKey)
				apiErr := apierr.Conflict("idempotency key: does not match the first request’s parameters")
				render.Status(r, apiErr.HTTPStatusCode())
				render.JSON(w, r, apiErr)
				return
			}

			// Write the cached response to the original ResponseWriter
			for key, values := range cachedResponse.Headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(cachedResponse.StatusCode)
			w.Write(cachedResponse.Body)
		})
	}
}
