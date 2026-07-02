package httpserver

import (
	"bytes"
	"context"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/perftrack"
)

// ModelStore manages in-memory models and their generated markdown content.
type ModelStore struct {
	state storeState
}

// NewModelStore creates a new model store.
func NewModelStore() *ModelStore {
	return &ModelStore{state: newStoreState()}
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
	snapshot, err := s.generateSnapshot(model, classErrors, tracker)
	if err != nil {
		return err
	}
	s.state.publish(name, snapshot, tracker)
	return nil
}

// SetModelError records a generation failure for a model. The model's previously
// generated content (if any) is left in place; renderMD shows the error instead.
func (s *ModelStore) SetModelError(name string, err error) {
	s.state.setModelError(name, err)
}

// GetModelError returns the recorded generation error for a model, if any.
func (s *ModelStore) GetModelError(name string) (string, bool) {
	return s.state.modelError(name)
}

// GetParseIssues returns the last recorded parse-issue index for a model, if any.
func (s *ModelStore) GetParseIssues(name string) (*generate.ParseIssueIndex, bool) {
	return s.state.parseIssuesFor(name)
}

// GetModel returns a model by name.
func (s *ModelStore) GetModel(name string) (*core.Model, bool) {
	return s.state.model(name)
}

// GetMarkdown returns markdown content for a specific model and file.
func (s *ModelStore) GetMarkdown(ctx context.Context, model, file string) ([]byte, bool) {
	return s.state.markdownFile(ctx, model, file)
}

// GetCSS returns CSS content for a model.
func (s *ModelStore) GetCSS(ctx context.Context, model string) ([]byte, bool) {
	return s.state.cssFor(ctx, model)
}

// GetSVG returns SVG content for a specific model and file.
func (s *ModelStore) GetSVG(ctx context.Context, model, file string) ([]byte, bool) {
	return s.state.svgFile(ctx, model, file)
}

// ListModels returns a list of all model names.
func (s *ModelStore) ListModels(ctx context.Context) []string {
	return s.state.modelNames(ctx)
}

func (s *ModelStore) generateSnapshot(model *core.Model, classErrors map[string]string, tracker *perftrack.Tracker) (publishedSnapshot, error) {
	var (
		mdContent   map[string][]byte
		svgContent  map[string][]byte
		cssContent  []byte
		parseIssues *generate.ParseIssueIndex
		err         error
	)
	perftrack.RunOn(tracker, "store.generate", func() {
		mdContent, svgContent, cssContent, parseIssues, err = generateModelContent(model, classErrors, tracker)
	})
	if err != nil {
		return publishedSnapshot{}, err
	}
	return publishedSnapshot{
		model:       model,
		markdown:    mdContent,
		svg:         svgContent,
		css:         cssContent,
		parseIssues: parseIssues,
	}, nil
}

func generateModelContent(model *core.Model, classErrors map[string]string, tracker *perftrack.Tracker) (map[string][]byte, map[string][]byte, []byte, *generate.ParseIssueIndex, error) {
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

// DiagramMode returns DiagramInlineMermaid so the HTTP server can render fences in-browser.
func (c *ContentCollector) DiagramMode() generate.DiagramOutputMode {
	return generate.DiagramInlineMermaid
}
