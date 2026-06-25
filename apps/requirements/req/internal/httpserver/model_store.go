package httpserver

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/perftrack"
)

// ModelStore manages in-memory models and their generated markdown content.
type ModelStore struct {
	mu          sync.RWMutex
	models      map[string]*core.Model       // Keyed by model name
	markdown    map[string]map[string][]byte // model -> file -> content
	css         map[string][]byte            // model -> CSS content
	svg         map[string]map[string][]byte // model -> file -> SVG content
	modelErrors map[string]string            // model -> last generation error message
	parseIssues map[string]*generate.ParseIssueIndex
}

// NewModelStore creates a new model store.
func NewModelStore() *ModelStore {
	return &ModelStore{
		models:      make(map[string]*core.Model),
		markdown:    make(map[string]map[string][]byte),
		css:         make(map[string][]byte),
		svg:         make(map[string]map[string][]byte),
		modelErrors: make(map[string]string),
		parseIssues: make(map[string]*generate.ParseIssueIndex),
	}
}

// SetModel stores a model and regenerates its content. On success it clears any
// previously recorded generation error for that model.
//
// classErrors maps a class key string to a parse-error message; those classes'
// pages render as red-bold error blocks. Pass nil when there are no failures.
func (s *ModelStore) SetModel(name string, model *core.Model, classErrors map[string]string) error {
	return s.SetModelTracked(name, model, classErrors, nil)
}

// SetModelTracked is SetModel with optional phase timings recorded on tracker.
func (s *ModelStore) SetModelTracked(name string, model *core.Model, classErrors map[string]string, tracker *perftrack.Tracker) error {
	// Generate without holding the store lock so HTTP readers can keep serving the
	// previous snapshot while a reload runs.
	var (
		mdContent   map[string][]byte
		svgContent  map[string][]byte
		cssContent  []byte
		parseIssues *generate.ParseIssueIndex
		err         error
	)
	perftrack.RunOn(tracker, "store.generate", func() {
		mdContent, svgContent, cssContent, parseIssues, err = s.generateContentTracked(model, classErrors, tracker)
	})
	if err != nil {
		return err
	}

	lockStart := time.Now()
	s.mu.Lock()
	if tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	s.models[name] = model
	s.markdown[name] = mdContent
	s.svg[name] = svgContent
	s.css[name] = cssContent
	s.parseIssues[name] = parseIssues
	delete(s.modelErrors, name) // Recovery: a successful generation clears the error.
	s.mu.Unlock()

	return nil
}

// SetModelError records a generation failure for a model. The model's previously
// generated content (if any) is left in place; renderMD shows the error instead.
func (s *ModelStore) SetModelError(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := "unknown error"
	if err != nil {
		msg = err.Error()
	}
	s.modelErrors[name] = msg
}

// GetModelError returns the recorded generation error for a model, if any.
func (s *ModelStore) GetModelError(name string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msg, ok := s.modelErrors[name]
	return msg, ok
}

// GetParseIssues returns the last recorded parse-issue index for a model, if any.
func (s *ModelStore) GetParseIssues(name string) (*generate.ParseIssueIndex, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, ok := s.parseIssues[name]
	return idx, ok
}

// GetModel returns a model by name.
func (s *ModelStore) GetModel(name string) (*core.Model, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	model, ok := s.models[name]
	return model, ok
}

// GetMarkdown returns markdown content for a specific model and file.
func (s *ModelStore) GetMarkdown(ctx context.Context, model, file string) ([]byte, bool) {
	lockStart := time.Now()
	s.mu.RLock()
	if tracker := perftrack.FromContext(ctx); tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	defer s.mu.RUnlock()

	var content []byte
	var ok bool
	perftrack.Run(ctx, "store.lookup", func() {
		if files, found := s.markdown[model]; found {
			content, ok = files[file]
		}
	})
	return content, ok
}

// GetCSS returns CSS content for a model.
func (s *ModelStore) GetCSS(ctx context.Context, model string) ([]byte, bool) {
	lockStart := time.Now()
	s.mu.RLock()
	if tracker := perftrack.FromContext(ctx); tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	defer s.mu.RUnlock()

	content, ok := s.css[model]
	return content, ok
}

// GetSVG returns SVG content for a specific model and file.
func (s *ModelStore) GetSVG(ctx context.Context, model, file string) ([]byte, bool) {
	lockStart := time.Now()
	s.mu.RLock()
	if tracker := perftrack.FromContext(ctx); tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	defer s.mu.RUnlock()

	if files, found := s.svg[model]; found {
		if content, ok := files[file]; ok {
			return content, true
		}
	}
	return nil, false
}

// ListModels returns a list of all model names.
func (s *ModelStore) ListModels(ctx context.Context) []string {
	lockStart := time.Now()
	s.mu.RLock()
	if tracker := perftrack.FromContext(ctx); tracker != nil {
		tracker.Add("store.lock_wait", time.Since(lockStart))
	}
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.models))
	for name := range s.models {
		names = append(names, name)
	}
	return names
}

func (s *ModelStore) generateContentTracked(model *core.Model, classErrors map[string]string, tracker *perftrack.Tracker) (map[string][]byte, map[string][]byte, []byte, *generate.ParseIssueIndex, error) {
	var parseIssues *generate.ParseIssueIndex
	perftrack.RunOn(tracker, "generate.parseIssues", func() {
		parseIssues = generate.BuildParseIssueIndex(model, classErrors)
	})

	collector := &ContentCollector{
		Markdown: make(map[string][]byte),
		SVG:      make(map[string][]byte),
	}

	var err error
	perftrack.RunOn(tracker, "generate.markdown", func() {
		err = generate.GenerateMdToWriter(*model, collector, classErrors)
	})
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var cssContent []byte
	perftrack.RunOn(tracker, "generate.css", func() {
		var cssBuffer bytes.Buffer
		generate.WriteCSS(&cssBuffer)
		cssContent = cssBuffer.Bytes()
	})

	return collector.Markdown, collector.SVG, cssContent, parseIssues, nil
}

// ContentCollector implements generate.ContentWriter to collect content in memory.
type ContentCollector struct {
	Markdown map[string][]byte
	SVG      map[string][]byte
	CSS      []byte
}

// WriteMarkdown stores markdown content.
func (c *ContentCollector) WriteMarkdown(filename string, content []byte) error {
	c.Markdown[filename] = content
	return nil
}

// WriteSVG stores SVG content.
func (c *ContentCollector) WriteSVG(filename string, content []byte) error {
	c.SVG[filename] = content
	return nil
}

// WriteCSS stores CSS content.
func (c *ContentCollector) WriteCSS(content []byte) error {
	c.CSS = content
	return nil
}
