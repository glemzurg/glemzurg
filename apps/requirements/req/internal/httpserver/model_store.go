package httpserver

import (
	"bytes"
	"sync"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
)

// ModelStore manages in-memory models and their generated markdown content.
type ModelStore struct {
	mu          sync.RWMutex
	models      map[string]*core.Model       // Keyed by model name
	markdown    map[string]map[string][]byte // model -> file -> content
	css         map[string][]byte            // model -> CSS content
	svg         map[string]map[string][]byte // model -> file -> SVG content
	modelErrors map[string]string            // model -> last generation error message
}

// NewModelStore creates a new model store.
func NewModelStore() *ModelStore {
	return &ModelStore{
		models:      make(map[string]*core.Model),
		markdown:    make(map[string]map[string][]byte),
		css:         make(map[string][]byte),
		svg:         make(map[string]map[string][]byte),
		modelErrors: make(map[string]string),
	}
}

// SetModel stores a model and regenerates its content. On success it clears any
// previously recorded generation error for that model.
//
// classErrors maps a class key string to a parse-error message; those classes'
// pages render as red-bold error blocks. Pass nil when there are no failures.
func (s *ModelStore) SetModel(name string, model *core.Model, classErrors map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate markdown content in memory
	mdContent, svgContent, cssContent, err := s.generateContent(model, classErrors)
	if err != nil {
		return err
	}

	s.models[name] = model
	s.markdown[name] = mdContent
	s.svg[name] = svgContent
	s.css[name] = cssContent
	delete(s.modelErrors, name) // Recovery: a successful generation clears the error.

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

// GetModel returns a model by name.
func (s *ModelStore) GetModel(name string) (*core.Model, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	model, ok := s.models[name]
	return model, ok
}

// GetMarkdown returns markdown content for a specific model and file.
func (s *ModelStore) GetMarkdown(model, file string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if files, ok := s.markdown[model]; ok {
		if content, ok := files[file]; ok {
			return content, true
		}
	}
	return nil, false
}

// GetCSS returns CSS content for a model.
func (s *ModelStore) GetCSS(model string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	content, ok := s.css[model]
	return content, ok
}

// GetSVG returns SVG content for a specific model and file.
func (s *ModelStore) GetSVG(model, file string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if files, ok := s.svg[model]; ok {
		if content, ok := files[file]; ok {
			return content, true
		}
	}
	return nil, false
}

// ListModels returns a list of all model names.
func (s *ModelStore) ListModels() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.models))
	for name := range s.models {
		names = append(names, name)
	}
	return names
}

// generateContent generates markdown, SVG, and CSS content for a model.
func (s *ModelStore) generateContent(model *core.Model, classErrors map[string]string) (map[string][]byte, map[string][]byte, []byte, error) {
	// Use the generate package to create content in memory
	collector := &ContentCollector{
		Markdown: make(map[string][]byte),
		SVG:      make(map[string][]byte),
	}

	err := generate.GenerateMdToWriter(*model, collector, classErrors)
	if err != nil {
		return nil, nil, nil, err
	}

	mdContent := collector.Markdown
	svgContent := collector.SVG

	// Generate CSS
	var cssBuffer bytes.Buffer
	generate.WriteCSS(&cssBuffer)
	cssContent := cssBuffer.Bytes()

	return mdContent, svgContent, cssContent, nil
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
