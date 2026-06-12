package generate

import (
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
			want:  "\n```\ndraft note\n```\n",
		},
		{
			name:  "multiline preserved",
			notes: "line one\nline two",
			want:  "\n```\nline one\nline two\n```\n",
		},
		{
			name:  "trims outer whitespace",
			notes: "  padded  ",
			want:  "\n```\npadded\n```\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, unfinishedNotesBlock(tc.notes))
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
			name:  "red parenthesized asterism",
			notes: "note",
			want:  ` (<span class="unfinished-notes-marker">` + _unfinishedNotesGlyph + `</span>)`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := unfinishedNotesMarker(tc.notes)
			assert.Equal(t, tc.want, got)
			if tc.want != "" {
				assert.Contains(t, got, "unfinished-notes-marker")
				assert.Contains(t, got, _unfinishedNotesGlyph)
			}
		})
	}
}
