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
	parserDocs "github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/docs"
	parserErrors "github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/errors"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai/json_schemas"
)

const helpText = `req_check - validate an AI-generated requirements model

Usage:
  req_check <model_path>              validate model, minimal error output
  req_check --json <model_path>       validate model, JSON error output
  req_check --explain <error_code>    show full remediation for an error (e.g. E5003)
  req_check --format-docs             show the JSON model format documentation
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
│       ├── associations/
│       │   └── <from>__<to>__<name>.assoc.json         domain-level association
│       └── subdomains/
│           └── <subdomain_key>/
│               ├── subdomain.json                      subdomain definition
│               ├── associations/
│               │   └── <from>__<to>__<name>.assoc.json subdomain-level association
│               ├── generalizations/
│               │   └── <key>.gen.json                  class generalization
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
├── associations/
│   └── <from>__<to>__<name>.assoc.json                 model-level association
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
		jsonOutput bool
		explainArg string
		formatDocs bool
		schemaArg  string
		showTree   bool
		showHelp   bool
		modelPath  string
	)

	flag.BoolVar(&jsonOutput, "json", false, "output errors as JSON array")
	flag.StringVar(&explainArg, "explain", "", "show full remediation for error code (e.g. E5003 or 5003)")
	flag.BoolVar(&formatDocs, "format-docs", false, "show the JSON model format documentation")
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
	if formatDocs {
		runFormatDocs()
		return
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

	// Output errors.
	if jsonOutput {
		outputJSON(errs)
	} else {
		outputText(errs)
	}
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

// outputText writes errors to stdout in concise text format.
func outputText(errs []error) {
	for i, err := range errs {
		if i > 0 {
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintln(os.Stdout, formatError(err))
	}
	fmt.Fprintf(os.Stdout, "\n%d error(s) found. Use --explain E{code} for detailed remediation.\n", len(errs))
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
			items = append(items, jsonError{
				Type:    "parse",
				Code:    fmt.Sprintf("E%d", pe.Code),
				Message: pe.Message,
				File:    pe.File,
				Field:   pe.Field,
				Hint:    pe.Hint,
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

// formatError formats an error for text output.
func formatError(err error) string {
	var pe *parser_ai.ParseError
	if errors.As(err, &pe) {
		return pe.Error()
	}
	var ve *coreerr.ValidationError
	if errors.As(err, &ve) {
		return ve.Error()
	}
	return err.Error()
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

// runFormatDocs shows the JSON model format documentation.
func runFormatDocs() {
	data, err := fs.ReadFile(parserDocs.Docs, "JSON_AI_MODEL_FORMAT.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load format documentation: %v\n", err)
		os.Exit(2)
	}
	fmt.Fprint(os.Stdout, string(data))
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
