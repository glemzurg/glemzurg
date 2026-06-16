package generate

import (
	"fmt"
	"html"
	"maps"
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/notation/tla_plus/convert"
)

// ParseIssueIndex groups every parse or expression error for one model generation pass.
type ParseIssueIndex struct {
	FileErrors  map[string]string
	ExprErrors  map[string][]string
	ModelErrors []string
	failedSpecs map[string]bool
}

// activeParseIssues is set for the duration of GenerateMdToWriter so template helpers
// can render markers without threading state through every template data struct.
var activeParseIssues *ParseIssueIndex

// BuildParseIssueIndex merges per-class file failures with expression issues found in the model.
func BuildParseIssueIndex(model *core.Model, fileErrors map[string]string) *ParseIssueIndex {
	idx := &ParseIssueIndex{
		FileErrors:  copyStringMap(fileErrors),
		ExprErrors:  make(map[string][]string),
		failedSpecs: make(map[string]bool),
	}

	for _, issue := range convert.CollectUnparsedExpressionIssues(model) {
		line := issue.Location + ": " + issue.Message
		if issue.SpecText != "" {
			idx.failedSpecs[issue.SpecText] = true
		}
		if issue.ClassKey.KeyType == "" {
			idx.ModelErrors = append(idx.ModelErrors, line)
			continue
		}
		key := issue.ClassKey.String()
		idx.ExprErrors[key] = append(idx.ExprErrors[key], line)
	}

	sort.Strings(idx.ModelErrors)
	for key := range idx.ExprErrors {
		sort.Strings(idx.ExprErrors[key])
	}

	return idx
}

// HasIssues reports whether the model has any recorded parse or expression errors.
func (idx *ParseIssueIndex) HasIssues() bool {
	if idx == nil {
		return false
	}
	if len(idx.FileErrors) > 0 || len(idx.ModelErrors) > 0 {
		return true
	}
	for _, msgs := range idx.ExprErrors {
		if len(msgs) > 0 {
			return true
		}
	}
	return false
}

// ClassHasIssues reports whether a class has a file-level or expression-level error.
func (idx *ParseIssueIndex) ClassHasIssues(classKey identity.Key) bool {
	if idx == nil {
		return false
	}
	key := classKey.String()
	if _, ok := idx.FileErrors[key]; ok {
		return true
	}
	return len(idx.ExprErrors[key]) > 0
}

// ClassMarker returns a visible warning glyph for class index pages.
func (idx *ParseIssueIndex) ClassMarker(classKey identity.Key) string {
	if !idx.ClassHasIssues(classKey) {
		return ""
	}
	return ` <span class="parse-error-marker" title="Parse error">&#9888;</span>`
}

// ClassExpressionBanner prepends a red summary when a class page has expression errors
// but its source file parsed successfully.
func (idx *ParseIssueIndex) ClassExpressionBanner(classKey identity.Key) string {
	if idx == nil {
		return ""
	}
	key := classKey.String()
	if _, fileFailed := idx.FileErrors[key]; fileFailed {
		return ""
	}
	msgs := idx.ExprErrors[key]
	if len(msgs) == 0 {
		return ""
	}
	return parseIssuesBannerHTML("Expression Parse Errors", msgs)
}

// ModelSummaryBanner returns a model-page summary of every parse issue in scope.
func (idx *ParseIssueIndex) ModelSummaryBanner(model *core.Model) string {
	if idx == nil || !idx.HasIssues() {
		return ""
	}

	var lines []string
	lines = append(lines, idx.ModelErrors...)

	type classLine struct {
		name string
		key  string
	}
	var classes []classLine
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for classKey, class := range subdomain.Classes {
				if !idx.ClassHasIssues(classKey) {
					continue
				}
				classes = append(classes, classLine{name: class.Name, key: classKey.String()})
			}
		}
	}
	sort.Slice(classes, func(i, j int) bool { return classes[i].name < classes[j].name })

	for _, cl := range classes {
		if msg, ok := idx.FileErrors[cl.key]; ok {
			lines = append(lines, fmt.Sprintf("class %s: %s", cl.name, msg))
			continue
		}
		for _, msg := range idx.ExprErrors[cl.key] {
			lines = append(lines, fmt.Sprintf("class %s: %s", cl.name, msg))
		}
	}

	return parseIssuesBannerHTML("Model Parse Errors", lines)
}

// IssueCount returns the total number of recorded parse and expression errors.
func (idx *ParseIssueIndex) IssueCount() int {
	if idx == nil {
		return 0
	}
	n := len(idx.ModelErrors)
	for _, msg := range idx.FileErrors {
		if msg != "" {
			n++
		}
	}
	for _, msgs := range idx.ExprErrors {
		n += len(msgs)
	}
	return n
}

// GlobalPageBanner summarizes issues for injection at the top of every rendered page.
func (idx *ParseIssueIndex) GlobalPageBanner() string {
	if idx == nil || !idx.HasIssues() {
		return ""
	}
	return fmt.Sprintf(
		`<div class="parse-error-banner"><p style="color:%s;font-weight:bold;">`+
			`&#9888; This model has parse errors (%d). See <a href="model.md">model.md</a> for details.</p></div>`,
		errorTextColor, idx.IssueCount())
}

func parseIssuesBannerHTML(heading string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="parse-error-banner"><h2 style="color:`)
	b.WriteString(errorTextColor)
	b.WriteString(`;">`)
	b.WriteString(html.EscapeString(heading))
	b.WriteString(`</h2><ul>`)
	for _, line := range lines {
		b.WriteString(`<li style="color:`)
		b.WriteString(errorTextColor)
		b.WriteString(`;font-weight:bold;">`)
		b.WriteString(html.EscapeString(line))
		b.WriteString(`</li>`)
	}
	b.WriteString(`</ul></div>`)
	return b.String()
}

func expressionSpecDisplay(spec logic_spec.ExpressionSpec) string {
	if spec.Specification == "" {
		return ""
	}
	escaped := html.EscapeString(spec.Specification)
	if activeParseIssues == nil || !activeParseIssues.failedSpecs[spec.Specification] {
		return escaped
	}
	return `<span class="parse-error-spec">` + escaped + `</span>`
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	maps.Copy(out, in)
	return out
}
