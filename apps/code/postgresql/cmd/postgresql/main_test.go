package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteRequiresModel(t *testing.T) {
	err := Execute(nil, &bytes.Buffer{})
	require.ErrorContains(t, err, "model directory is required")
}

func TestExecuteReadsSiblingAnnotations(t *testing.T) {
	root := filepath.Join(t.TempDir(), "sample")
	write := func(rel, body string) {
		t.Helper()
		path := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write("model.json", `{"name":"Shop Books"}`)
	write("model.annot.json", `{"schema":"shop"}`)
	write("domains/shop/domain.json", `{"name":"Shop"}`)
	write("domains/shop/subdomains/default/subdomain.json", `{"name":"Default"}`)
	write("domains/shop/subdomains/default/classes/account/class.json", `{"name":"Account"}`)

	var stdout bytes.Buffer
	err := Execute([]string{"-model", root}, &stdout)
	require.NoError(t, err)
	require.Equal(t, `model: Shop Books
key: sample
schema: shop
domain: shop name=Shop
subdomain: shop/default name=Default
class: shop/default/account name=Account attributes=0
`, stdout.String())
}
