package httpserver

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

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
	return mux
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
		fmt.Fprintf(w, "data: %s\n\n", msg)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// mainHandler handles all non-SSE requests.
func (s *Server) mainHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		s.homeHandler(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.NotFound(w, r)
		return
	}

	model := parts[0]
	if len(parts) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/%s/model.md", model), http.StatusFound)
		return
	}

	if len(parts) == 2 {
		file := parts[1]
		if strings.HasSuffix(file, ".md") {
			s.renderMD(model, file, w)
			return
		} else if strings.HasSuffix(file, ".svg") {
			s.serveSVG(model, file, w, r)
			return
		} else if strings.HasSuffix(file, ".css") {
			s.serveCSS(model, w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// homeHandler displays a list of available models.
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	models := s.store.ListModels()

	var list strings.Builder
	list.WriteString("<ul>")
	for _, model := range models {
		list.WriteString(fmt.Sprintf("<li><a href=\"/%s/model.md\">%s</a></li>", model, model))
	}
	list.WriteString("</ul>")

	html := "<html><body><h1>Models</h1>" + list.String() + "</body></html>"
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// renderMD renders a Markdown file from the in-memory store as HTML.
func (s *Server) renderMD(model, file string, w http.ResponseWriter) {
	data, ok := s.store.GetMarkdown(model, file)
	if !ok {
		http.NotFound(w, nil)
		return
	}

	mdHTML := markdown.ToHTML(data, nil, nil)

	script := fmt.Sprintf(`
<script>
const evtSource = new EventSource("/events/%s/%s");
evtSource.onmessage = () => location.reload();
</script>
`, model, file)

	html := fmt.Sprintf(`<html><head><link rel="stylesheet" href="/%s/style.css">%s</head><body>%s</body></html>`, model, script, mdHTML)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// serveSVG serves an SVG file from the in-memory store.
func (s *Server) serveSVG(model, file string, w http.ResponseWriter, r *http.Request) {
	data, ok := s.store.GetSVG(model, file)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write(data)
}

// serveCSS serves the CSS file from the in-memory store.
func (s *Server) serveCSS(model string, w http.ResponseWriter, r *http.Request) {
	data, ok := s.store.GetCSS(model)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(data)
}
