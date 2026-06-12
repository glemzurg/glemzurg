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
			want:  "\n```\n" + _unfinishedNotesGlyph + "\n\ndraft note\n```\n",
		},
		{
			name:  "multiline preserved",
			notes: "line one\nline two",
			want:  "\n```\n" + _unfinishedNotesGlyph + "\n\nline one\nline two\n```\n",
		},
		{
			name:  "trims outer whitespace",
			notes: "  padded  ",
			want:  "\n```\n" + _unfinishedNotesGlyph + "\n\npadded\n```\n",
		},
		{
			name:  "literal content in fence",
			notes: "<script>alert(1)</script>",
			want:  "\n```\n" + _unfinishedNotesGlyph + "\n\n<script>alert(1)</script>\n```\n",
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
			name:  "parenthesized asterism",
			notes: "note",
			want:  " (" + _unfinishedNotesGlyph + ")",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, unfinishedNotesMarker(tc.notes))
		})
	}
}
