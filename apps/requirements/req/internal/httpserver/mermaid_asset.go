package httpserver

import (
	"bytes"
	_ "embed"
	"net/http"
)

// Mermaid 11.x bundle served by the requirements HTTP server so diagram pages
// work offline and browsers can cache the library across reloads.
//
//go:embed assets/mermaid.min.js
var mermaidJS []byte

const mermaidJSPath = "/mermaid.min.js"

// mermaidCacheControl lets browsers reuse the bundle across page reloads while
// still picking up a new embed when the server binary is redeployed.
const mermaidCacheControl = "public, max-age=86400"

func markdownHasMermaid(data []byte) bool {
	return bytes.Contains(data, []byte("```mermaid"))
}

func (s *Server) serveMermaidJS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", mermaidCacheControl)
	_, _ = w.Write(mermaidJS) //nolint:gosec // embedded static asset
}
