package perftrack

import (
	"net/http"
	"strings"
)

// IsLongLivedStream reports paths whose handlers intentionally block until the
// client disconnects (e.g. Server-Sent Events). Total duration is not a useful
// latency signal for these requests.
func IsLongLivedStream(path string) bool {
	return strings.HasPrefix(path, "/events/")
}

// Middleware wraps an HTTP handler to record per-request phase timings and log
// operations that exceed SlowThreshold.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if IsLongLivedStream(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		tracker := New(r.Method + " " + r.URL.Path)
		ctx := WithContext(r.Context(), tracker)
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r.WithContext(ctx))

		tracker.SetStatus(recorder.status)
		tracker.LogIfSlow()
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
