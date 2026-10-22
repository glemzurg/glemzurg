package sqlgen

import (
	"maps"
	"os"
	"path/filepath"
	"testing"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/parser_json"
	"github.com/stretchr/testify/require"
)

const (
	accountClass = "domains/shop/subdomains/default/classes/account/class.json"
	accountAnnot = "domains/shop/subdomains/default/classes/account/class.annot.json"
	entryClass   = "domains/shop/subdomains/default/classes/entry/class.json"
	entryAnnot   = "domains/shop/subdomains/default/classes/entry/class.annot.json"
	holdsAssoc   = "domains/shop/subdomains/default/class_associations/account--entry--holds.assoc.json"
	holdsAnnot   = "domains/shop/subdomains/default/class_associations/account--entry--holds.assoc.annot.json"
)

func TestCreateTables(t *testing.T) {
	nullableModel, nullableNotes := onlyAccount(
		`{"name":"Account","attributes":[{"key":"note","name":"Note","nullable":true}]}`,
		`{"attributes":{"note":{"type":"text"}}}`,
	)
	derivedModel, derivedNotes := onlyAccount(
		`{"name":"Account","attributes":[{"key":"balance","name":"Balance"},{"key":"total","name":"Total","derivation_policy":{"type":"value","description":"Computed","notation":"tla_plus"}}]}`,
		`{"attributes":{"balance":{"type":"numeric(12,2)","default":"0"}}}`,
	)
	skippedModel, skippedNotes := onlyAccount(
		`{"name":"Account","attributes":[{"key":"balance","name":"Balance"},{"key":"label","name":"Label"}]}`,
		`{"attributes":{"balance":{"include":false},"label":{"type":"text"}}}`,
	)
	excluded := shopNotes()
	excluded[entryAnnot] = `{"include":false}`

	tests := []struct {
		name  string
		model map[string]string
		notes map[string]string
		want  string
		err   string
	}{
		{
			name:  "foreign key on the many class",
			model: shopModel(),
			notes: shopNotes(),
			want:  shopSQL,
		},
		{
			name:  "foreign key on the from class",
			model: withFile(shopModel(), holdsAssoc, `{"name":"Holds","from_class_key":"account","from_multiplicity":"0..*","to_class_key":"entry","to_multiplicity":"1"}`),
			notes: shopNotes(),
			want: `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  balance numeric(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE entry (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);

ALTER TABLE account ADD COLUMN entry_pk bigint NOT NULL;
ALTER TABLE account ADD CONSTRAINT fk_account_entry
  FOREIGN KEY (entry_pk) REFERENCES entry (pk) ON DELETE CASCADE;
`,
		},
		{
			name:  "optional foreign key",
			model: withFile(shopModel(), holdsAssoc, `{"name":"Holds","from_class_key":"account","from_multiplicity":"0..1","to_class_key":"entry","to_multiplicity":"0..*"}`),
			notes: withFile(shopNotes(), holdsAnnot, `{"on_delete":"SET NULL"}`),
			want: `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  balance numeric(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE entry (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);

ALTER TABLE entry ADD COLUMN account_pk bigint;
ALTER TABLE entry ADD CONSTRAINT fk_entry_account
  FOREIGN KEY (account_pk) REFERENCES account (pk) ON DELETE SET NULL;
`,
		},
		{
			name:  "nullable column",
			model: nullableModel,
			notes: nullableNotes,
			want: `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  note text
);
`,
		},
		{
			name:  "derived attribute is not a column",
			model: derivedModel,
			notes: derivedNotes,
			want: `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  balance numeric(12,2) NOT NULL DEFAULT 0
);
`,
		},
		{
			name:  "excluded attribute is not a column",
			model: skippedModel,
			notes: skippedNotes,
			want: `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  label text NOT NULL
);
`,
		},
		{
			name:  "attribute type annotation",
			model: shopModel(),
			notes: withFile(shopNotes(), accountAnnot, `{"attributes":{"balance":{}}}`),
			err:   "attribute balance on account needs a PostgreSQL type",
		},
		{
			name:  "on_delete annotation",
			model: shopModel(),
			notes: without(shopNotes(), holdsAnnot),
			err:   "association holds needs an on_delete annotation",
		},
		{
			name:  "one-to-one needs an annotation",
			model: withFile(shopModel(), holdsAssoc, `{"name":"Holds","from_class_key":"account","from_multiplicity":"1","to_class_key":"entry","to_multiplicity":"1"}`),
			notes: shopNotes(),
			err:   "association holds is one-to-one; an annotation must say which end holds the foreign key",
		},
		{
			name:  "many-to-many is not a column",
			model: withFile(shopModel(), holdsAssoc, `{"name":"Holds","from_class_key":"account","from_multiplicity":"*","to_class_key":"entry","to_multiplicity":"*"}`),
			notes: shopNotes(),
			err:   "association holds is many-to-many; a join table is not generated yet",
		},
		{
			name:  "excluded class has no table",
			model: shopModel(),
			notes: excluded,
			err:   "association holds needs table entry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model, notes := loadModel(t, tc.model, tc.notes)
			got, err := CreateTables(model, notes)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

const shopSQL = `CREATE TABLE account (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  balance numeric(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE entry (
  pk bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);

ALTER TABLE entry ADD COLUMN account_pk bigint NOT NULL;
ALTER TABLE entry ADD CONSTRAINT fk_entry_account
  FOREIGN KEY (account_pk) REFERENCES account (pk) ON DELETE CASCADE;
`

func shopModel() map[string]string {
	return map[string]string{
		"model.json":               `{"name":"Shop Books"}`,
		"domains/shop/domain.json": `{"name":"Shop"}`,
		"domains/shop/subdomains/default/subdomain.json": `{"name":"Default"}`,
		accountClass: `{"name":"Account","attributes":[{"key":"balance","name":"Balance"}]}`,
		entryClass:   `{"name":"Entry"}`,
		holdsAssoc:   `{"name":"Holds","from_class_key":"account","from_multiplicity":"1","to_class_key":"entry","to_multiplicity":"0..*"}`,
	}
}

func shopNotes() map[string]string {
	return map[string]string{
		accountAnnot: `{"attributes":{"balance":{"type":"numeric(12,2)","default":"0"}}}`,
		holdsAnnot:   `{"on_delete":"CASCADE"}`,
	}
}

func onlyAccount(classJSON, annotJSON string) (map[string]string, map[string]string) {
	model := shopModel()
	delete(model, entryClass)
	delete(model, holdsAssoc)
	model[accountClass] = classJSON
	return model, map[string]string{accountAnnot: annotJSON}
}

func withFile(files map[string]string, rel, body string) map[string]string {
	out := maps.Clone(files)
	out[rel] = body
	return out
}

func without(files map[string]string, rel string) map[string]string {
	out := maps.Clone(files)
	delete(out, rel)
	return out
}

func loadModel(t *testing.T, modelFiles, noteFiles map[string]string) (core.Model, annot.Set) {
	t.Helper()
	model, err := parser_json.ReadModel(writeTree(t, modelFiles))
	require.NoError(t, err)
	notes, err := annot.Read(writeTree(t, noteFiles))
	require.NoError(t, err)
	return model, notes
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "sample")
	for rel, body := range files {
		path := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	return root
}
