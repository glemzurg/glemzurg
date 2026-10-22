package sqlgen

import (
	"bytes"
	"embed"
	"sync"
	"text/template"
)

//go:embed templates/*.sql.template
var templateFS embed.FS

var (
	parseOnce sync.Once
	parsed    *template.Template
	parseErr  error
)

func templates() (*template.Template, error) {
	parseOnce.Do(func() {
		parsed = template.New("sql")
		for _, name := range []string{"create_table.sql.template", "foreign_key.sql.template"} {
			body, err := templateFS.ReadFile("templates/" + name)
			if err != nil {
				parseErr = err
				return
			}
			if _, err := parsed.New(name).Parse(string(body)); err != nil {
				parseErr = err
				return
			}
		}
	})
	return parsed, parseErr
}

// create_table.sql.template always emits a synthetic primary key. The requirements
// model has no such column; foreign keys reference it.
func render(tables []tableView, keys []foreignKeyView) (string, error) {
	tmpl, err := templates()
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	for i, table := range tables {
		if i > 0 {
			b.WriteByte('\n')
		}
		if err := tmpl.ExecuteTemplate(&b, "create_table.sql.template", table); err != nil {
			return "", err
		}
	}
	for _, key := range keys {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		if err := tmpl.ExecuteTemplate(&b, "foreign_key.sql.template", key); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}
