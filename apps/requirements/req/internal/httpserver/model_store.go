package httpserver

import (
	"bytes"
	"sync"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

// ModelStore manages in-memory models and their generated markdown content.
type ModelStore struct {
	mu       sync.RWMutex
	models   map[string]*req_model.Model  // Keyed by model name
	markdown map[string]map[string][]byte // model -> file -> content
	css      map[string][]byte            // model -> CSS content
	svg      map[string]map[string][]byte // model -> file -> SVG content
}

// NewModelStore creates a new model store.
func NewModelStore() *ModelStore {
	return &ModelStore{
		models:   make(map[string]*req_model.Model),
		markdown: make(map[string]map[string][]byte),
		css:      make(map[string][]byte),
		svg:      make(map[string]map[string][]byte),
	}
}

// SetModel stores a model and regenerates its content.
func (s *ModelStore) SetModel(name string, model *req_model.Model) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.models[name] = model

	// Generate markdown content in memory
	mdContent, svgContent, cssContent, err := s.generateContent(name, model)
	if err != nil {
		return err
	}

	s.markdown[name] = mdContent
	s.svg[name] = svgContent
	s.css[name] = cssContent

	return nil
}

// GetModel returns a model by name.
func (s *ModelStore) GetModel(name string) (*req_model.Model, bool) {
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
func (s *ModelStore) generateContent(name string, model *req_model.Model) (map[string][]byte, map[string][]byte, []byte, error) {
	mdContent := make(map[string][]byte)
	svgContent := make(map[string][]byte)

	// Use the generate package to create content in memory
	collector := &ContentCollector{
		Markdown: make(map[string][]byte),
		SVG:      make(map[string][]byte),
	}

	err := generate.GenerateMdToWriter(false, *model, collector)
	if err != nil {
		return nil, nil, nil, err
	}

	mdContent = collector.Markdown
	svgContent = collector.SVG

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
