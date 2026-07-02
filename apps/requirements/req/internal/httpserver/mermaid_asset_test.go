package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeMermaidJS(t *testing.T) {
	server := NewServer(NewModelStore())
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, mermaidJSPath, nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/javascript; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, mermaidCacheControl, rec.Header().Get("Cache-Control"))
	require.NotEmpty(t, mermaidJS)
	assert.Equal(t, mermaidJS, rec.Body.Bytes())
}

func TestMarkdownHasMermaid(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want bool
	}{
		{name: "mermaid fence", md: "# Title\n\n```mermaid\nclassDiagram\n```\n", want: true},
		{name: "no mermaid", md: "# Title\n\nPlain text only.\n", want: false},
		{name: "other fence", md: "# Title\n\n```go\npackage main\n```\n", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, markdownHasMermaid([]byte(tc.md)))
		})
	}
}

func TestRenderMDUsesLocalMermaidOnlyWhenNeeded(t *testing.T) {
	model := test_helper.GetTestModel()
	store := NewModelStore()
	require.NoError(t, store.SetModel("test_model", &model, nil))
	server := NewServer(store)

	t.Run("diagram page loads local mermaid", func(t *testing.T) {
		code, body := requestMD(server, "/test_model/model.md")
		require.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `<script src="`+mermaidJSPath+`"></script>`)
		assert.NotContains(t, body, "cdn.jsdelivr.net")
		assert.Contains(t, body, "mermaid.run();")
	})

	t.Run("text-only page skips mermaid", func(t *testing.T) {
		var actorFile string
		for key := range model.Actors {
			actorFile = "actor-" + strings.ReplaceAll(key.String(), "/", ".") + ".md"
			break
		}
		require.NotEmpty(t, actorFile)

		code, body := requestMD(server, "/test_model/"+actorFile)
		require.Equal(t, http.StatusOK, code)
		assert.NotContains(t, body, mermaidJSPath)
		assert.NotContains(t, body, "mermaid.run();")
		assert.NotContains(t, body, "cdn.jsdelivr.net")
	})
}
