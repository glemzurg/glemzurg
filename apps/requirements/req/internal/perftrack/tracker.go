package perftrack

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"
)

// SlowThreshold is the duration above which an operation is logged with span detail.
const SlowThreshold = time.Second

type ctxKey struct{}

// Tracker records named phase durations for one operation (HTTP request or model reload).
type Tracker struct {
	name   string
	start  time.Time
	spans  map[string]time.Duration
	status int
}

// New starts tracking an operation identified by name (e.g. "GET /evenplay/model.md").
func New(name string) *Tracker {
	return &Tracker{
		name:  name,
		start: time.Now(),
		spans: make(map[string]time.Duration),
	}
}

// WithContext attaches a tracker to ctx for downstream phase recording.
func WithContext(ctx context.Context, tracker *Tracker) context.Context {
	if tracker == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, tracker)
}

// FromContext returns the tracker attached to ctx, or nil.
func FromContext(ctx context.Context) *Tracker {
	if ctx == nil {
		return nil
	}
	tracker, _ := ctx.Value(ctxKey{}).(*Tracker)
	return tracker
}

// Run executes fn and records its duration under name on the context tracker.
func Run(ctx context.Context, name string, fn func()) {
	FromContext(ctx).run(name, fn)
}

// RunOn executes fn and records its duration on the given tracker.
func RunOn(tracker *Tracker, name string, fn func()) {
	if tracker == nil {
		fn()
		return
	}
	tracker.run(name, fn)
}

// Add records a pre-measured duration (e.g. lock wait before a guarded section).
func (t *Tracker) Add(name string, d time.Duration) {
	if t != nil {
		t.spans[name] += d
	}
}

func (t *Tracker) run(name string, fn func()) {
	if t == nil {
		fn()
		return
	}
	start := time.Now()
	fn()
	t.spans[name] += time.Since(start)
}

// SetStatus records the HTTP response status for slow-request logs.
func (t *Tracker) SetStatus(status int) {
	if t != nil {
		t.status = status
	}
}

// Elapsed returns time since the tracker was created.
func (t *Tracker) Elapsed() time.Duration {
	if t == nil {
		return 0
	}
	return time.Since(t.start)
}

// LogIfSlow writes a structured log line when the operation exceeded SlowThreshold.
func (t *Tracker) LogIfSlow() {
	if t == nil {
		return
	}
	elapsed := t.Elapsed()
	if elapsed < SlowThreshold {
		return
	}

	var b strings.Builder
	b.WriteString("slow operation: ")
	b.WriteString(t.name)
	if t.status != 0 {
		b.WriteString(" status=")
		b.WriteString(itoa(t.status))
	}
	b.WriteString(" total=")
	b.WriteString(elapsed.String())
	if len(t.spans) > 0 {
		b.WriteString(" spans=")
		b.WriteString(formatSpans(t.spans))
	}
	log.Println(b.String())
}

func formatSpans(spans map[string]time.Duration) string {
	names := make([]string, 0, len(spans))
	for name := range spans {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return spans[names[i]] > spans[names[j]]
	})

	var parts []string
	for _, name := range names {
		parts = append(parts, name+"="+spans[name].String())
	}
	return strings.Join(parts, " ")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
