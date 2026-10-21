package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_json"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/test_helper"
)

func TestWriteErrorMarkdown(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "sample_model")

	err := writeErrorMarkdown(outputPath, errors.New("parse failed: line 8"))
	if err != nil {
		t.Fatalf("writeErrorMarkdown failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputPath, "model.md"))
	if err != nil {
		t.Fatalf("expected model.md to be written: %v", err)
	}
	body := string(content)
	if !strings.Contains(body, "ERROR: parse failed: line 8") {
		t.Errorf("expected error message in model.md, got: %s", body)
	}
	if !strings.Contains(body, "color:#cc0000") || !strings.Contains(body, "font-weight:bold") {
		t.Errorf("expected red bold styling in model.md, got: %s", body)
	}
}

// A failed markdown conversion writes the error to <output>/model.md and still
// returns the error.
func TestProcessConversionFailureWritesErrorMarkdown(t *testing.T) {
	rootOutput := t.TempDir()

	// Point at a source root that has no such model: parsing fails.
	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: t.TempDir(), rootOutputPath: rootOutput}, "sample_model", conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatMD})
	if err == nil {
		t.Fatal("expected processConversion to fail for a missing model")
	}

	content, readErr := os.ReadFile(filepath.Join(rootOutput, "sample_model", "model.md"))
	if readErr != nil {
		t.Fatalf("expected error model.md to be written: %v", readErr)
	}
	if !strings.Contains(string(content), "Model Generation Failed") {
		t.Errorf("expected error document, got: %s", content)
	}
}

// A failed non-markdown conversion does not write a model.md error file.
func TestProcessConversionFailureNonMarkdownNoErrorFile(t *testing.T) {
	rootOutput := t.TempDir()

	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: t.TempDir(), rootOutputPath: rootOutput}, "sample_model", conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatAIJSON})
	if err == nil {
		t.Fatal("expected processConversion to fail for a missing model")
	}

	if _, statErr := os.Stat(filepath.Join(rootOutput, "sample_model", "model.md")); statErr == nil {
		t.Error("did not expect a model.md error file for ai/json output")
	}
}

func TestProcessConversionJSONWritesModelFile(t *testing.T) {
	rootSource, modelName := writeHumanFixture(t)
	rootOutput := t.TempDir()

	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: rootSource, rootOutputPath: rootOutput}, modelName, conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatJSON})
	if err != nil {
		t.Fatalf("processConversion json: %v", err)
	}

	path := filepath.Join(rootOutput, modelName, "model.json")
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected model.json: %v", readErr)
	}
	if !strings.Contains(string(data), `"format":"core-pipe"`) {
		t.Fatalf("expected core-pipe document, got: %s", data)
	}
	got, readModelErr := parser_json.Read(bytes.NewReader(data))
	if readModelErr != nil {
		t.Fatalf("read model.json: %v", readModelErr)
	}
	// Human parse keys the model by its source directory. The pipe keeps that
	// key; the authored name and domain tree are what survive the round trip.
	if got.Name != "Test Model" {
		t.Fatalf("model name %q", got.Name)
	}
	if len(got.Domains) == 0 {
		t.Fatal("expected domains in the core-pipe document")
	}
	if !modelHasClass(got, "Order") {
		t.Fatal("expected class Order in the core-pipe document")
	}
}

func TestProcessConversionJSONWritesStdout(t *testing.T) {
	rootSource, modelName := writeHumanFixture(t)

	oldStdout := os.Stdout
	reader, writer, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	os.Stdout = writer
	var captured bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&captured, reader)
		close(done)
	}()

	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: rootSource, rootOutputPath: "-"}, modelName, conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatJSON})
	_ = writer.Close()
	os.Stdout = oldStdout
	<-done
	if err != nil {
		t.Fatalf("processConversion json stdout: %v", err)
	}
	if !strings.HasPrefix(captured.String(), `{"format":"core-pipe"`) {
		t.Fatalf("expected core-pipe on stdout, got: %s", captured.String())
	}
}

func TestProcessConversionJSONRefusesClassParseFailure(t *testing.T) {
	rootSource, modelName := writeHumanFixture(t)
	var found string
	walkErr := filepath.WalkDir(filepath.Join(rootSource, modelName), func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && found == "" && strings.HasSuffix(path, "warehouse.class") {
			found = path
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk: %v", walkErr)
	}
	if found == "" {
		t.Fatal("expected warehouse.class in the human fixture")
	}
	orig, readErr := os.ReadFile(found)
	if readErr != nil {
		t.Fatalf("read class: %v", readErr)
	}
	if err := os.WriteFile(found, append(orig, []byte("\nbroken_key: \"unterminated\n")...), 0o644); err != nil { //nolint:gosec // test fixture corruption
		t.Fatalf("corrupt class: %v", err)
	}

	rootOutput := t.TempDir()
	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: rootSource, rootOutputPath: rootOutput}, modelName, conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatJSON})
	if err == nil {
		t.Fatal("expected json output to be refused when a class file failed to parse")
	}
	if !strings.Contains(err.Error(), "refusing core-pipe JSON output") {
		t.Fatalf("expected refusal, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(rootOutput, modelName, "model.json")); statErr == nil {
		t.Fatal("did not expect model.json when class parse failed")
	}
	if _, statErr := os.Stat(filepath.Join(rootOutput, modelName, "model.md")); statErr == nil {
		t.Fatal("did not expect model.md for json output")
	}
}

func modelHasClass(model core.Model, name string) bool {
	for _, domain := range model.Domains {
		for _, subdomain := range domain.Subdomains {
			for _, class := range subdomain.Classes {
				if class.Name == name {
					return true
				}
			}
		}
	}
	return false
}

func writeHumanFixture(t *testing.T) (rootSource, modelName string) {
	t.Helper()
	rootSource = t.TempDir()
	modelName = "sample"
	if err := parser_human.Write(test_helper.GetTestModel(), filepath.Join(rootSource, modelName)); err != nil {
		t.Fatalf("write human fixture: %v", err)
	}
	return rootSource, modelName
}
