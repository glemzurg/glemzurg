package annot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	root := t.TempDir()
	writeAnnot(t, root, "model.annot.json", `{"schema":"shop"}`)
	writeAnnot(t, root, filepath.Join("domains", "shop", "subdomains", "default", "classes", "account", "class.annot.json"),
		`{"include":true,"attributes":{"balance":{"type":"numeric(12,2)","default":"0"}}}`)
	writeAnnot(t, root, filepath.Join("domains", "shop", "subdomains", "default", "class_associations", "account--entry--holds.assoc.annot.json"),
		`{"on_delete":"CASCADE"}`)
	writeAnnot(t, root, filepath.Join("actors", "customer.actor.annot.json"), `{}`)
	require.NoError(t, os.WriteFile(filepath.Join(root, "model.json"), []byte(`{"name":"Shop"}`), 0o600))

	set, err := Read(root)
	require.NoError(t, err)
	require.Equal(t, []string{
		"actors/customer.actor.json",
		"domains/shop/subdomains/default/class_associations/account--entry--holds.assoc.json",
		"domains/shop/subdomains/default/classes/account/class.json",
		"model.json",
	}, set.Paths())

	modelNote, ok := set.Note("model.json")
	require.True(t, ok)
	require.Equal(t, "shop", modelNote.Schema)

	classNote, ok := set.Note("domains/shop/subdomains/default/classes/account/class.json")
	require.True(t, ok)
	require.NotNil(t, classNote.Include)
	require.True(t, *classNote.Include)
	require.Equal(t, "numeric(12,2)", classNote.Attributes["balance"].Type)
	require.Equal(t, "0", classNote.Attributes["balance"].Default)

	_, ok = set.Note("model.json.bak")
	require.False(t, ok)
}

func TestReadRejects(t *testing.T) {
	tests := []struct {
		name    string
		rel     string
		body    string
		wantErr string
	}{
		{name: "unknown field", rel: "model.annot.json", body: `{"table":"shop"}`, wantErr: "unknown field"},
		{name: "schema on class", rel: "domains/shop/subdomains/default/classes/account/class.annot.json", body: `{"schema":"shop"}`, wantErr: "schema is only valid"},
		{name: "bad schema ident", rel: "model.annot.json", body: `{"schema":"Shop"}`, wantErr: "not a PostgreSQL identifier"},
		{name: "on delete on class", rel: "domains/shop/subdomains/default/classes/account/class.annot.json", body: `{"on_delete":"CASCADE"}`, wantErr: "on_delete is only valid"},
		{name: "bad on delete", rel: "class_associations/a--b--c.assoc.annot.json", body: `{"on_delete":"DROP"}`, wantErr: "not a PostgreSQL delete action"},
		{name: "attributes on model", rel: "model.annot.json", body: `{"attributes":{"balance":{}}}`, wantErr: "attributes are only valid"},
		{name: "bad attribute key", rel: "domains/shop/classes/account/class.annot.json", body: `{"attributes":{"Bad":{}}}`, wantErr: "not an identifier"},
		{name: "blank type", rel: "domains/shop/classes/account/class.annot.json", body: `{"attributes":{"balance":{"type":" "}}}`, wantErr: "not a PostgreSQL type"},
		{name: "not an object", rel: "model.annot.json", body: `[]`, wantErr: "cannot unmarshal"},
		{name: "empty file", rel: "model.annot.json", body: ``, wantErr: "EOF"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeAnnot(t, root, tc.rel, tc.body)
			_, err := Read(root)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestReadMissingDirectory(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
}

func writeAnnot(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
}
