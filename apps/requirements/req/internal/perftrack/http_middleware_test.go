package perftrack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsLongLivedStream(t *testing.T) {
	assert.True(t, IsLongLivedStream("/events/evenplay/model.md"))
	assert.False(t, IsLongLivedStream("/evenplay/model.md"))
}

func TestMiddlewareSkipsLongLivedStreams(t *testing.T) {
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, FromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/events/evenplay/model.md", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddlewarePassesContextTracker(t *testing.T) {
	var got *Tracker
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = FromContext(r.Context())
		w.WriteHeader(http.StatusTeapot)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.NotNil(t, got)
	assert.Equal(t, "GET /test", got.name)
	assert.Equal(t, http.StatusTeapot, got.status)
}
