package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/lengzuo/fundflow/pkg/log"
	"github.com/lengzuo/fundflow/utils"
)

//go:generate mockery --name Params --output ./mocks --outpkg mocks --case=underscore
type Params interface {
	Validate() apierr.JSON
}

//go:generate mockery --name Responder --output ./mocks --outpkg mocks --case=underscore
type Responder interface {
	StatusCode() int
}

type RestfulFunc[In Params, Out Responder] func(context.Context, In) (Out, apierr.JSON)

func Handle[In Params, Out Responder](f RestfulFunc[In, Out]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in In

		// Retrieve data from request.
		if r.Body != nil && r.Body != http.NoBody {
			err := json.NewDecoder(r.Body).Decode(&in)
			if err != nil {
				jsonErr := apierr.BadRequest("invalid json")
				render.Status(r, jsonErr.HTTPStatusCode())
				render.JSON(w, r, jsonErr)
				return
			}
		}

		// Parse any querystring
		if err := utils.Decoder.Decode(&in, r.URL.Query()); err != nil {
			jsonErr := apierr.BadRequest("invalid querystring")
			render.Status(r, jsonErr.HTTPStatusCode())
			render.JSON(w, r, jsonErr)
			return
		}

		// Perform simple validation
		if err := in.Validate(); err != nil {
			render.Status(r, err.HTTPStatusCode())
			render.JSON(w, r, err)
			return
		}

		// Call out to target function
		out, err := f(r.Context(), in)
		if err != nil {
			var jsonErr apierr.JSON
			if errors.As(err, &jsonErr) {
				// Format error response
				render.Status(r, jsonErr.HTTPStatusCode())
				render.JSON(w, r, jsonErr)
				return
			}
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, "unable to process response")
			return
		}

		// Format and write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(out.StatusCode())
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Error(r.Context(), "failed to encode json: %v", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, "unable to process response")
			return
		}
	})
}
