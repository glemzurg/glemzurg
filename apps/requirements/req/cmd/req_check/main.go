package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/coreerr"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	parserErrors "github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
)

const helpText = `req_check - validate an AI-generated requirements model

Usage:
  req_check <model_path>              validate model (JSON output)
  req_check --explain <error_code>    show full remediation for an error (e.g. E5003)
  req_check --schema <entity>         show JSON schema (model, class, action, ...)
  req_check --tree                    show expected directory tree structure
  req_check --help                    show this help

Exit codes: 0 = valid, 1 = validation errors, 2 = usage error
`

const treeText = `Expected directory structure for a requirements model:

<model_name>/
├── model.json                                          model definition
├── actors/
│   └── <actor_key>.actor.json                          actor definitions
├── actor_generalizations/
│   └── <key>.agen.json                                 actor generalization
├── domains/
│   └── <domain_key>/
│       ├── domain.json                                 domain definition
│       ├── class_associations/
│       │   └── <from>--<to>--<name>.assoc.json         domain-level association
│       └── subdomains/
│           └── <subdomain_key>/
│               ├── subdomain.json                      subdomain definition
│               ├── class_associations/
│               │   └── <from>--<to>--<name>.assoc.json subdomain-level association
│               ├── class_generalizations/
│               │   └── <key>.cgen.json                 class generalization
│               └── classes/
│                   └── <class_key>/
│                       ├── class.json                  class definition (attributes, indexes)
│                       ├── state_machine.json          state machine (states, events, guards, transitions)
│                       ├── actions/
│                       │   └── <action_key>.json       action (requires, guarantees, safety_rules)
│                       ├── queries/
│                       │   └── <query_key>.json        query (requires, guarantees)
│                       ├── invariants/
│                       │   └── <N>.json                class invariant logic
│                       ├── use_cases/
│                       │   └── <use_case_key>/
│                       │       ├── use_case.json       use case definition
│                       │       └── scenarios/
│                       │           └── <scenario_key>.json
│                       └── attributes/
│                           └── <attr_key>/
│                               └── invariants/
│                                   └── <N>.json        attribute invariant logic
├── class_associations/
│   └── <from>--<to>--<name>.assoc.json                 model-level association
├── domain_associations/
│   └── <key>.domain_assoc.json                         domain association
├── global_functions/
│   └── <func_key>.json                                 global function
└── named_sets/
    └── <set_key>.json                                  named set

Keys must be lowercase snake_case: ^[a-z][a-z0-9]*(_[a-z0-9]+)*$
`

func main() {
	var (
		explainArg string
		schemaArg  string
		showTree   bool
		showHelp   bool
		modelPath  string
	)

	flag.StringVar(&explainArg, "explain", "", "show full remediation for error code (e.g. E5003 or 5003)")
	flag.StringVar(&schemaArg, "schema", "", "show JSON schema for entity (model, class, action, ...)")
	flag.BoolVar(&showTree, "tree", false, "show expected directory tree structure")
	flag.BoolVar(&showHelp, "help", false, "show help")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
	}
	flag.Parse()

	// Handle info flags (no model path needed).
	if showHelp {
		fmt.Fprint(os.Stdout, helpText)
		os.Exit(0)
	}
	if showTree {
		fmt.Fprint(os.Stdout, treeText)
		os.Exit(0)
	}
	if schemaArg != "" {
		runSchema(schemaArg)
		return
	}
	if explainArg != "" {
		runExplain(explainArg)
		return
	}

	// Model validation mode — need a path.
	if flag.NArg() > 0 {
		modelPath = flag.Arg(0)
	}
	if modelPath == "" {
		fmt.Fprint(os.Stderr, helpText)
		os.Exit(2)
	}

	// Diagnostics to stderr.
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.Printf("validating: %s", modelPath)

	errs := validateModel(modelPath)
	if len(errs) == 0 {
		fmt.Fprintln(os.Stdout, "OK")
		os.Exit(0)
	}

	// Output errors as JSON.
	outputJSON(errs)
	os.Exit(1)
}

// validateModel reads and validates a model, returning all errors found.
func validateModel(modelPath string) []error {
	var allErrors []error

	// Read and validate the input model.
	m, err := parser_ai.ReadModel(modelPath)
	if err != nil {
		allErrors = append(allErrors, flattenErrors(err)...)
		return allErrors
	}

	// Validate the core model.
	if err := m.Validate(); err != nil {
		allErrors = append(allErrors, flattenErrors(err)...)
	}

	return allErrors
}

// flattenErrors unwraps joined errors into individual errors.
func flattenErrors(err error) []error {
	if err == nil {
		return nil
	}
	// Check if it's a joined error (errors.Join produces this).
	type joinedError interface {
		Unwrap() []error
	}
	if je, ok := err.(joinedError); ok {
		var result []error
		for _, e := range je.Unwrap() {
			result = append(result, flattenErrors(e)...)
		}
		return result
	}
	return []error{err}
}

// outputJSON writes errors to stdout as a JSON array.
func outputJSON(errs []error) {
	outputJSONTo(os.Stdout, errs)
}

// outputJSONTo writes errors as a JSON array to the given writer.
func outputJSONTo(w io.Writer, errs []error) {
	type jsonError struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
		File    string `json:"file,omitempty"`
		Field   string `json:"field,omitempty"`
		Hint    string `json:"hint,omitempty"`
		Path    string `json:"path,omitempty"`
		Got     string `json:"got,omitempty"`
		Want    string `json:"want,omitempty"`
	}

	var items []jsonError
	for _, err := range errs {
		var pe *parser_ai.ParseError
		var ve *coreerr.ValidationError
		switch {
		case errors.As(err, &pe):
			code := fmt.Sprintf("E%d", pe.Code)
			hint := buildParseErrorHint(pe, code)
			items = append(items, jsonError{
				Type:    "parse",
				Code:    code,
				Message: pe.Message,
				File:    pe.File,
				Field:   pe.Field,
				Hint:    hint,
			})
		case errors.As(err, &ve):
			items = append(items, jsonError{
				Type:    "validation",
				Code:    string(ve.Code()),
				Message: ve.Message(),
				Field:   ve.Field(),
				Path:    coreerr.FormatPath(ve.Path()),
				Got:     ve.Got(),
				Want:    ve.Want(),
			})
		default:
			items = append(items, jsonError{
				Type:    "error",
				Message: err.Error(),
			})
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(items)
}

// buildParseErrorHint constructs the hint string for a parse error in JSON output.
// It combines the error's own hint with category-specific follow-up commands:
//   - File/directory structure errors: adds --tree
//   - All parse errors: adds --explain E{code}
func buildParseErrorHint(pe *parser_ai.ParseError, code string) string {
	parts := []string{}
	if pe.Hint != "" {
		parts = append(parts, pe.Hint)
	}
	// File/directory structure errors get --tree guidance.
	if isFileStructureError(pe.Code) {
		parts = append(parts, "run: req_check --tree")
	}
	// All parse errors get --explain with the specific code.
	parts = append(parts, "run: req_check --explain "+code)
	return strings.Join(parts, " | ")
}

// isFileStructureError returns true for errors about missing files/directories
// or malformed filenames — cases where --tree helps the AI understand the
// expected directory layout.
func isFileStructureError(code int) bool {
	switch code {
	case
		parser_ai.ErrTreeModelNoActors,                // need to create actor files
		parser_ai.ErrTreeModelNoDomains,               // need to create domain dirs
		parser_ai.ErrTreeDomainNoSubdomains,           // need to create subdomain dirs
		parser_ai.ErrTreeSubdomainTooFewClasses,       // need to create class dirs
		parser_ai.ErrTreeSubdomainNoAssociations,      // need to create assoc files
		parser_ai.ErrTreeClassNoStateMachine,          // need to create state_machine.json
		parser_ai.ErrKeyInvalidFormat,                 // filename/dir naming
		parser_ai.ErrAssocFilenameInvalidFormat,       // assoc filename format
		parser_ai.ErrAssocFilenameInvalidComponent,    // assoc filename component
		parser_ai.ErrTreeSingleSubdomainNotDefault,    // dir naming
		parser_ai.ErrTreeMultipleSubdomainsHasDefault: // dir naming
		return true
	}
	return false
}

// runExplain shows full remediation for an error code.
func runExplain(arg string) {
	if err := runExplainTo(os.Stdout, arg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
}

// runExplainTo writes the error documentation to the given writer.
func runExplainTo(w io.Writer, arg string) error {
	// Strip leading 'E' if present.
	codeStr := strings.TrimPrefix(arg, "E")
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return fmt.Errorf("invalid error code: %q (expected E1234 or 1234)", arg)
	}

	content, _, loadErr := parserErrors.LoadErrorDoc(code)
	if loadErr != nil {
		return fmt.Errorf("no documentation found for error code E%d", code)
	}

	fmt.Fprint(w, content)
	return nil
}

// runSchema shows the JSON schema for a given entity type.
func runSchema(entity string) {
	if err := runSchemaTo(os.Stdout, entity); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}
}

// runSchemaTo writes the JSON schema for a given entity type to the given writer.
func runSchemaTo(w io.Writer, entity string) error {
	name := strings.ToLower(entity)
	filename := name + ".schema.json"

	data, err := json_schemas.Schemas.ReadFile(filename)
	if err != nil {
		// List available schemas.
		entries, _ := fs.ReadDir(json_schemas.Schemas, ".")
		var available []string
		for _, e := range entries {
			n := e.Name()
			if after, ok := strings.CutSuffix(n, ".schema.json"); ok {
				available = append(available, after)
			}
		}
		return fmt.Errorf("unknown schema: %q\navailable: %s", entity, strings.Join(available, ", "))
	}
	fmt.Fprint(w, string(data))
	return nil
}
