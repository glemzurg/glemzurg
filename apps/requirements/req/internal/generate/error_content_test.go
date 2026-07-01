package generate

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorMarkdown(t *testing.T) {
	out := string(ErrorMarkdown(errors.New("failed to parse: line 8")))

	if !strings.Contains(out, "# Model Generation Failed") {
		t.Errorf("expected a heading, got: %s", out)
	}
	if !strings.Contains(out, "ERROR: failed to parse: line 8") {
		t.Errorf("expected the error message, got: %s", out)
	}
	if !strings.Contains(out, "color:#cc0000") || !strings.Contains(out, "font-weight:bold") {
		t.Errorf("expected red bold styling, got: %s", out)
	}
}

func TestErrorMarkdownEscapesHTML(t *testing.T) {
	out := string(ErrorMarkdown(errors.New("bad <tag> & \"quote\"")))
	if strings.Contains(out, "<tag>") {
		t.Errorf("error message should be HTML-escaped, got: %s", out)
	}
	if !strings.Contains(out, "&lt;tag&gt;") {
		t.Errorf("expected escaped tag, got: %s", out)
	}
}

func TestErrorMarkdownNilError(t *testing.T) {
	out := string(ErrorMarkdown(nil))
	if !strings.Contains(out, "unknown error") {
		t.Errorf("expected fallback text for nil error, got: %s", out)
	}
}

func TestErrorPageHTML(t *testing.T) {
	out := string(ErrorPageHTML("test_model", "model.md", errors.New("boom")))

	if !strings.Contains(out, "ERROR: boom") {
		t.Errorf("expected the error message, got: %s", out)
	}
	if !strings.Contains(out, "color:#cc0000") || !strings.Contains(out, "font-weight:bold") {
		t.Errorf("expected red bold styling, got: %s", out)
	}
	// Recovery: must keep the model-wide SSE reload script.
	if !strings.Contains(out, `EventSource("/events/test_model")`) {
		t.Errorf("expected SSE reload script, got: %s", out)
	}
	if !strings.Contains(out, `pagehide`) {
		t.Errorf("expected pagehide listener to close EventSource, got: %s", out)
	}
	if !strings.Contains(out, `href="/test_model/style.css"`) {
		t.Errorf("expected stylesheet link, got: %s", out)
	}
}

func TestErrorPageHTMLEscapes(t *testing.T) {
	out := string(ErrorPageHTML("m", "f.md", errors.New("bad <script>")))
	if strings.Contains(out, "bad <script>") {
		t.Errorf("error message should be HTML-escaped, got: %s", out)
	}
}
