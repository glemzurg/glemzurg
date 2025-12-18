// handlers.go
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gomarkdown/markdown"
)

func eventsHandler(w http.ResponseWriter, r *http.Request) { // Handles Server-Sent Events for refresh notifications.
	path := strings.TrimPrefix(r.URL.Path, "/events/") // Extracts path after /events/.
	if path == "" {                                    // Checks if path is empty.
		http.NotFound(w, r) // Returns 404 if no key specified.
		return
	}

	key := strings.TrimSuffix(path, "/") // Removes trailing slash if present, key is model/file.md.
	broker := getBroker(key)             // Gets the broker for the key.

	w.Header().Set("Content-Type", "text/event-stream") // Sets SSE content type.
	w.Header().Set("Cache-Control", "no-cache")         // Disables caching.
	w.Header().Set("Connection", "keep-alive")          // Keeps connection open.
	w.Header().Set("Transfer-Encoding", "chunked")      // Enables chunked transfer.

	clientChan := make(chan []byte) // Creates a channel for this client.
	broker.newClients <- clientChan // Registers the client with the broker.

	var once sync.Once                                          // Ensures cleanup is done only once to prevent double close.
	closeFunc := func() { broker.closingClients <- clientChan } // Function to unregister client.
	defer once.Do(closeFunc)                                    // Defers cleanup, executed only once.

	notify := r.Context().Done() // Gets context done channel for cancellation.
	go func() {                  // Starts goroutine to handle context cancellation.
		<-notify           // Waits for context done.
		once.Do(closeFunc) // Unregisters client only once.
	}()

	for { // Loop to send events.
		msg, open := <-clientChan // Receives message from client channel.
		if !open {                // Checks if channel closed.
			break // Exits loop.
		}
		fmt.Fprintf(w, "data: %s\n\n", msg) // Writes SSE data format.
		if f, ok := w.(http.Flusher); ok {  // Checks if response writer supports flushing.
			f.Flush() // Flushes to send immediately.
		}
	}
}

func Handler(rootMdPath string) func(http.ResponseWriter, *http.Request) { // Returns the main request handler closure with rootMdPath captured.
	return func(w http.ResponseWriter, r *http.Request) { // Main request handler for all paths.
		path := strings.TrimPrefix(r.URL.Path, "/") // Removes leading slash from path.
		if path == "" {                             // Checks for root path.
			homeHandler(rootMdPath, w, r) // Calls home handler.
			return
		}

		parts := strings.Split(path, "/") // Splits path into components.
		if len(parts) == 0 {              // Invalid path check.
			http.NotFound(w, r) // Returns 404.
			return
		}

		model := parts[0]    // Extracts model name.
		if len(parts) == 1 { // Handles /model or /model/ path.
			http.Redirect(w, r, fmt.Sprintf("/%s/model.md", model), http.StatusFound) // Redirects to /model/model.md for consistency.
			return
		}

		if len(parts) == 2 { // Handles /model/file paths.
			file := parts[1]                    // Extracts file name.
			if strings.HasSuffix(file, ".md") { // Checks if Markdown file.
				renderMD(rootMdPath, model, file, file == "model.md", w) // Renders Markdown.
				return
			} else { // Handles other files like SVG, CSS.
				serveFile(rootMdPath, model, file, w, r) // Serves static file.
				return
			}
		}

		http.NotFound(w, r) // Returns 404 for invalid paths.
	}
}

func homeHandler(rootMdPath string, w http.ResponseWriter, r *http.Request) { // Handles root path to list models.
	dirs, err := os.ReadDir(rootMdPath) // Reads models directory.
	if err != nil {                     // Checks for read error.
		http.Error(w, "Error reading models directory", http.StatusInternalServerError) // Returns 500 error.
		return
	}

	var list strings.Builder // Builder for efficient string concatenation.
	list.WriteString("<ul>") // Starts unordered list.
	for _, d := range dirs { // Iterates over directory entries.
		if d.IsDir() { // Checks if entry is a directory (model).
			list.WriteString(fmt.Sprintf("<li><a href=\"/%s/model.md\">%s</a></li>", d.Name(), d.Name())) // Adds link to model.md.
		}
	}
	list.WriteString("</ul>") // Closes list.

	html := "<html><body><h1>Models</h1>" + list.String() + "</body></html>" // Constructs full HTML.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")               // Sets response type.
	_, _ = w.Write([]byte(html))                                             // Writes response.
}

func renderMD(rootMdPath, model, file string, isMain bool, w http.ResponseWriter) { // Renders Markdown file as HTML.
	path := filepath.Join(rootMdPath, model, file) // Constructs full file path.
	data, err := os.ReadFile(path)                 // Reads file content.
	if err != nil {                                // Checks for read error.
		http.NotFound(w, nil) // Returns 404 if file not found.
		return
	}

	mdHTML := markdown.ToHTML(data, nil, nil) // Converts Markdown to HTML.

	header := "" // Initializes header string.

	script := fmt.Sprintf(`
<script>
const evtSource = new EventSource("/events/%s/%s");
evtSource.onmessage = () => location.reload();
</script>
`, model, file)

	html := fmt.Sprintf(`<html><head><link rel="stylesheet" href="/%s/style.css">%s</head><body>%s%s</body></html>`, model, script, header, mdHTML) // Constructs full HTML with CSS, script, header, and content.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")                                                                                      // Sets response type.
	_, _ = w.Write([]byte(html))                                                                                                                    // Writes response.
}

func serveFile(rootMdPath, model, file string, w http.ResponseWriter, r *http.Request) { // Serves static files like SVG or CSS.
	path := filepath.Join(rootMdPath, model, file) // Constructs full file path.
	http.ServeFile(w, r, path)                     // Serves the file directly.
}
