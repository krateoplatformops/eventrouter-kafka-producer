package middlewares

import (
	"context"
	"net/http"

	"github.com/krateoplatformops/eventrouter-kafka-producer/internal/helpers/uuid"
	"github.com/rs/zerolog"
)

const (
	correlationIdHeader = "X-Request-Id"
	correlationIdKey    = "requestId"
)

// CorrelationID returns a middleware that add a
// correlation identifier to the HTTP request.
func CorrelationID(next http.Handler) http.Handler {
	uidGen := uuid.NewGen()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.Header.Get(correlationIdHeader)
		if id == "" {
			// generate new version 4 uuid
			id = uidGen.NewV4().String()
		}

		// set the id to the request context
		ctx = context.WithValue(ctx, correlationIdKey, id)
		r = r.WithContext(ctx)

		// fetch the logger from context and update the context
		// with the correlation id value
		log := zerolog.Ctx(ctx)
		log.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str(correlationIdKey, id)
		})

		// set the response header
		w.Header().Set(correlationIdHeader, id)
		next.ServeHTTP(w, r)
	})
}
