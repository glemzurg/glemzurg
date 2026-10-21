package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/parser_json"
	"github.com/stretchr/testify/require"
)

func TestWalkBindsPostgreSQLNotes(t *testing.T) {
	modelDir := shopModel(t)
	notes := shopNotes(t)

	model, err := parser_json.ReadModel(modelDir)
	require.NoError(t, err)
	set, err := annot.Read(notes)
	require.NoError(t, err)

	result, err := Walk(model, set)
	require.NoError(t, err)
	require.Equal(t, shopReport, Format(result))
}

func TestWalkRejectsUnknownAttribute(t *testing.T) {
	modelDir := shopModel(t)
	root := t.TempDir()
	writeFile(t, root, "domains/shop/subdomains/default/classes/account/class.annot.json", `{"attributes":{"missing":{}}}`)

	model, err := parser_json.ReadModel(modelDir)
	require.NoError(t, err)
	set, err := annot.Read(root)
	require.NoError(t, err)
	_, err = Walk(model, set)
	require.ErrorContains(t, err, "unknown attribute missing")
}

func TestWalkRejectsUnappliedNote(t *testing.T) {
	modelDir := shopModel(t)
	root := t.TempDir()
	writeFile(t, root, "do.annot.json", `{}`)

	model, err := parser_json.ReadModel(modelDir)
	require.NoError(t, err)
	set, err := annot.Read(root)
	require.NoError(t, err)
	_, err = Walk(model, set)
	require.ErrorContains(t, err, "annotations not applied: do.json")
}

const shopReport = `model: Shop Books
key: sample
schema: shop
domain: shop name=Shop
subdomain: shop/default name=Default
class: shop/default/account name=Account attributes=1 include=true
attribute: shop/default/account.balance type=numeric(12,2) default=0
class: shop/default/entry name=Entry attributes=0
association: shop/default/account--entry--holds on_delete=CASCADE
`

func shopModel(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "sample")
	writeFile(t, root, "model.json", `{"name":"Shop Books"}`)
	writeFile(t, root, "domains/shop/domain.json", `{"name":"Shop"}`)
	writeFile(t, root, "domains/shop/subdomains/default/subdomain.json", `{"name":"Default"}`)
	writeFile(t, root, "domains/shop/subdomains/default/classes/account/class.json",
		`{"name":"Account","attributes":[{"key":"balance","name":"Balance"}]}`)
	writeFile(t, root, "domains/shop/subdomains/default/classes/entry/class.json", `{"name":"Entry"}`)
	writeFile(t, root, "domains/shop/subdomains/default/class_associations/account--entry--holds.assoc.json",
		`{"name":"Holds","from_class_key":"account","from_multiplicity":"1","to_class_key":"entry","to_multiplicity":"0..*"}`)
	return root
}

func shopNotes(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "notes")
	writeFile(t, root, "model.annot.json", `{"schema":"shop"}`)
	writeFile(t, root, "domains/shop/subdomains/default/classes/account/class.annot.json",
		`{"include":true,"attributes":{"balance":{"type":"numeric(12,2)","default":"0"}}}`)
	writeFile(t, root, "domains/shop/subdomains/default/class_associations/account--entry--holds.assoc.annot.json",
		`{"on_delete":"CASCADE"}`)
	return root
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
}
