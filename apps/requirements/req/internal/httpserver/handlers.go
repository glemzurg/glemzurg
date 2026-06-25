package httpserver

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"
	"sync"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/perftrack"
	"github.com/gomarkdown/markdown"
)

// Server handles HTTP requests for serving model documentation.
type Server struct {
	store    *ModelStore
	registry *BrokerRegistry
}

// NewServer creates a new HTTP server with the given model store.
func NewServer(store *ModelStore) *Server {
	return &Server{
		store:    store,
		registry: NewBrokerRegistry(),
	}
}

// Handler returns the main HTTP handler function.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.mainHandler)
	mux.HandleFunc("/events/", s.eventsHandler)
	return perftrack.Middleware(mux)
}

// NotifyModel sends refresh notifications for all pages of a model.
func (s *Server) NotifyModel(model string) {
	s.registry.NotifyModel(model)
}

// NotifyAll sends refresh notifications to all connected clients.
func (s *Server) NotifyAll() {
	s.registry.NotifyAll()
}

// eventsHandler handles Server-Sent Events for refresh notifications.
func (s *Server) eventsHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/events/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	key := strings.TrimSuffix(path, "/")
	broker := s.registry.GetBroker(key)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	clientChan := make(chan []byte)
	broker.Register(clientChan)

	var once sync.Once
	closeFunc := func() { broker.Unregister(clientChan) }
	defer once.Do(closeFunc)

	notify := r.Context().Done()
	go func() {
		<-notify
		once.Do(closeFunc)
	}()

	for {
		msg, open := <-clientChan
		if !open {
			break
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", msg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// mainHandler handles all non-SSE requests.
func (s *Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		s.homeHandler(ctx, w)
		return
	}
	if path == strings.TrimPrefix(mermaidJSPath, "/") {
		s.serveMermaidJS(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}

	model := parts[0]
	if len(parts) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/%s/model.md", model), http.StatusFound) //nolint:gosec // model name comes from the local store, not user input
		return
	}

	if len(parts) == 2 {
		file := parts[1]
		switch {
		case strings.HasSuffix(file, ".md"):
			s.renderMD(ctx, model, file, w)
			return
		case strings.HasSuffix(file, ".svg"):
			s.serveSVG(ctx, model, file, w, r)
			return
		case strings.HasSuffix(file, ".css"):
			s.serveCSS(ctx, model, w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// homeHandler displays a list of available models.
func (s *Server) homeHandler(ctx context.Context, w http.ResponseWriter) {
	var page string
	perftrack.Run(ctx, "home.build", func() {
		models := s.store.ListModels(ctx)

		var list strings.Builder
		list.WriteString("<ul>")
		for _, model := range models {
			escaped := html.EscapeString(model)
			marker := ""
			if idx, ok := s.store.GetParseIssues(model); ok && idx.HasIssues() {
				marker = ` <span class="parse-error-marker" title="Parse errors">&#9888;</span>`
			} else if _, ok := s.store.GetModelError(model); ok {
				marker = ` <span class="parse-error-marker" title="Generation failed">&#9888;</span>`
			}
			fmt.Fprintf(&list, "<li><a href=\"/%s/model.md\">%s</a>%s</li>", escaped, escaped, marker)
		}
		list.WriteString("</ul>")

		page = "<html><head><style>.parse-error-marker{color:#cc0000;font-weight:bold;}</style></head><body><h1>Models</h1>" + list.String() + "</body></html>"
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	perftrack.Run(ctx, "response.write", func() {
		_, _ = w.Write([]byte(page))
	})
}

// renderMD renders a Markdown file from the in-memory store as HTML.
func (s *Server) renderMD(ctx context.Context, model, file string, w http.ResponseWriter) {
	if msg, ok := s.store.GetModelError(model); ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		perftrack.Run(ctx, "response.write", func() {
			_, _ = w.Write(generate.ErrorPageHTML(model, file, errors.New(msg))) //nolint:gosec // error page HTML is escaped in generate.ErrorPageHTML
		})
		return
	}

	var data []byte
	var found bool
	perftrack.Run(ctx, "store.getMarkdown", func() {
		data, found = s.store.GetMarkdown(ctx, model, file)
	})
	if !found {
		http.NotFound(w, nil)
		return
	}

	var mdHTML []byte
	perftrack.Run(ctx, "markdown.toHTML", func() {
		mdHTML = markdown.ToHTML(data, nil, nil)
	})

	var body []byte
	perftrack.Run(ctx, "html.build", func() {
		body = buildMDPageHTML(model, file, data, mdHTML)
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	perftrack.Run(ctx, "response.write", func() {
		_, _ = w.Write(body)
	})
}

func buildMDPageHTML(model, file string, mdSource, mdHTML []byte) []byte {
	escapedModel := html.EscapeString(model)
	escapedFile := html.EscapeString(file)

	var buf strings.Builder
	buf.WriteString(`<html><head><link rel="stylesheet" href="/`)
	buf.WriteString(escapedModel)
	buf.WriteString(`/style.css"><script>const evtSource = new EventSource("/events/`)
	buf.WriteString(escapedModel)
	buf.WriteString(`/`)
	buf.WriteString(escapedFile)
	buf.WriteString(`");evtSource.onmessage = () => location.reload();</script>`)
	if markdownHasMermaid(mdSource) {
		buf.WriteString(`<script src="`)
		buf.WriteString(mermaidJSPath)
		buf.WriteString(`"></script>`)
	}
	buf.WriteString(`</head><body>`)
	buf.Write(mdHTML)
	if markdownHasMermaid(mdSource) {
		buf.WriteString(`<script>`)
		buf.WriteString(`document.querySelectorAll('pre code.language-mermaid').forEach(function(el){`)
		buf.WriteString(`var d=document.createElement('div');d.className='mermaid';`)
		buf.WriteString(`d.textContent=el.textContent;el.parentElement.replaceWith(d);});`)
		buf.WriteString(`mermaid.initialize({startOnLoad:false,securityLevel:'loose'});mermaid.run();`)
		buf.WriteString(`</script>`)
	}
	buf.WriteString(`</body></html>`)
	return []byte(buf.String())
}

// serveSVG serves an SVG file from the in-memory store.
func (s *Server) serveSVG(ctx context.Context, model, file string, w http.ResponseWriter, r *http.Request) {
	var data []byte
	var ok bool
	perftrack.Run(ctx, "store.getSVG", func() {
		data, ok = s.store.GetSVG(ctx, model, file)
	})
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	perftrack.Run(ctx, "response.write", func() {
		_, _ = w.Write(data)
	})
}

// serveCSS serves the CSS file from the in-memory store.
func (s *Server) serveCSS(ctx context.Context, model string, w http.ResponseWriter, r *http.Request) {
	var data []byte
	var ok bool
	perftrack.Run(ctx, "store.getCSS", func() {
		data, ok = s.store.GetCSS(ctx, model)
	})
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	perftrack.Run(ctx, "response.write", func() {
		_, _ = w.Write(data)
	})
}
