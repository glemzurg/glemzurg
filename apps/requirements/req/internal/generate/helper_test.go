package generate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnfinishedNotesBlock(t *testing.T) {
	tests := []struct {
		name  string
		notes string
		want  string
	}{
		{
			name:  "empty",
			notes: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			notes: "  \n\t  ",
			want:  "",
		},
		{
			name:  "single line",
			notes: "draft note",
			want: "\n<pre class=\"unfinished-notes-block\"><span class=\"unfinished-notes-glyph\">" +
				_unfinishedNotesGlyph + "</span><br />\ndraft note\n</pre>\n",
		},
		{
			name:  "multiline preserved",
			notes: "line one\nline two",
			want: "\n<pre class=\"unfinished-notes-block\"><span class=\"unfinished-notes-glyph\">" +
				_unfinishedNotesGlyph + "</span><br />\nline one\nline two\n</pre>\n",
		},
		{
			name:  "trims outer whitespace",
			notes: "  padded  ",
			want: "\n<pre class=\"unfinished-notes-block\"><span class=\"unfinished-notes-glyph\">" +
				_unfinishedNotesGlyph + "</span><br />\npadded\n</pre>\n",
		},
		{
			name:  "escapes HTML in notes",
			notes: "<script>alert(1)</script>",
			want: "\n<pre class=\"unfinished-notes-block\"><span class=\"unfinished-notes-glyph\">" +
				_unfinishedNotesGlyph + "</span><br />\n&lt;script&gt;alert(1)&lt;/script&gt;\n</pre>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := unfinishedNotesBlock(tc.notes)
			assert.Equal(t, tc.want, got)
			if strings.TrimSpace(tc.notes) != "" {
				assert.Contains(t, got, `class="unfinished-notes-glyph"`)
				assert.Contains(t, got, _unfinishedNotesGlyph)
			}
		})
	}
}

func TestUnfinishedNotesMarker(t *testing.T) {
	tests := []struct {
		name  string
		notes string
		want  string
	}{
		{
			name:  "empty",
			notes: "",
			want:  "",
		},
		{
			name:  "whitespace only",
			notes: "   ",
			want:  "",
		},
		{
			name:  "red warning glyph",
			notes: "note",
			want:  ` <span class="unfinished-notes-glyph">` + _unfinishedNotesGlyph + `</span>`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := unfinishedNotesMarker(tc.notes)
			assert.Equal(t, tc.want, got)
			if tc.want != "" {
				assert.Contains(t, got, "unfinished-notes-glyph")
			}
		})
	}
}
