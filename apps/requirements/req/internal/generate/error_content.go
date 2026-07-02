package generate

import (
	"fmt"
	"html"
)

// errorTextColor is the CSS color used to render generation errors.
const errorTextColor = "#cc0000"

// errorText returns the plain error message shown to the user.
func errorText(err error) string {
	if err == nil {
		return "unknown error"
	}
	return err.Error()
}

// errorMarkdownDoc builds a markdown error document with the given heading and
// a red-bold error block (inline HTML so it stands out raw and when served).
func errorMarkdownDoc(heading, message string) []byte {
	doc := fmt.Sprintf("# %s\n\n"+
		"<p style=\"color:%s;font-weight:bold;\">ERROR: %s</p>\n",
		html.EscapeString(heading), errorTextColor, html.EscapeString(message))
	return []byte(doc)
}

// ErrorMarkdown renders a whole-model generation error as a markdown document.
// Used for the md export so a failed run still produces a file where the
// content would normally be.
func ErrorMarkdown(err error) []byte { //nolint:revive // public API name
	return errorMarkdownDoc("Model Generation Failed", errorText(err))
}

// ClassErrorMarkdown renders the page for a single class whose source file
// failed to parse: the rest of the model still renders, only this class page
// shows the error.
func ClassErrorMarkdown(className, message string) []byte { //nolint:revive // public API name
	heading := "Class Failed to Parse"
	if className != "" {
		heading = className + " — Failed to Parse"
	}
	return errorMarkdownDoc(heading, message)
}

// ReloadEventsScript returns inline JavaScript that subscribes to model-wide SSE
// refresh notifications. The pagehide listener closes the stream promptly on
// navigation so the browser frees its per-origin HTTP/1.1 connection slot.
func ReloadEventsScript(model string) string {
	escapedModel := html.EscapeString(model)
	return fmt.Sprintf(
		`<script>const evtSource=new EventSource("/events/%s");`+
			`evtSource.onmessage=()=>location.reload();`+
			`addEventListener("pagehide",()=>evtSource.close());</script>`,
		escapedModel)
}

// ErrorPageHTML renders a generation error as a full HTML page for the web
// display. It keeps the same stylesheet link and Server-Sent-Events reload
// script that a normal page uses, so the page recovers automatically once the
// source is fixed.
func ErrorPageHTML(model, file string, err error) []byte { //nolint:revive // public API name
	escapedModel := html.EscapeString(model)
	escapedErr := html.EscapeString(errorText(err))

	page := fmt.Sprintf(
		"<html><head>"+
			"<link rel=\"stylesheet\" href=\"/%s/style.css\">"+
			"%s"+
			"</head><body>"+
			"<h1>Model Generation Failed</h1>"+
			"<p style=\"color:%s;font-weight:bold;\">ERROR: %s</p>"+
			"</body></html>",
		escapedModel, ReloadEventsScript(model), errorTextColor, escapedErr)
	_ = file // file is part of the URL path; SSE is keyed by model only
	return []byte(page)
}
