package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteErrorMarkdown(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "evenplay")

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
	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: t.TempDir(), rootOutputPath: rootOutput}, "evenplay", conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatMD})
	if err == nil {
		t.Fatal("expected processConversion to fail for a missing model")
	}

	content, readErr := os.ReadFile(filepath.Join(rootOutput, "evenplay", "model.md"))
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

	err := processConversion(conversionFlags{debug: false, skipDB: true}, conversionPaths{rootSourcePath: t.TempDir(), rootOutputPath: rootOutput}, "evenplay", conversionFormats{inputFormat: InputFormatDataYAML, outputFormat: OutputFormatAIJSON})
	if err == nil {
		t.Fatal("expected processConversion to fail for a missing model")
	}

	if _, statErr := os.Stat(filepath.Join(rootOutput, "evenplay", "model.md")); statErr == nil {
		t.Error("did not expect a model.md error file for ai/json output")
	}
}
