package parser_ai_input

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// t_TestFile represents a test file pair: input JSON and expected output JSON.
type t_TestFile struct {
	Filename     string // The input JSON filename.
	InputJSON    string // The contents of the input JSON file.
	ExpectedJSON string // The expected parsed output as JSON.
}

// t_ContentsForAllJSONFiles reads all JSON test file pairs from a directory.
// Files are expected to be named like:
// - 01_basic.input.json (the input file)
// - 01_basic.expected.json (the expected parsed output)
func t_ContentsForAllJSONFiles(path string) (allFiles []t_TestFile, err error) {

	// Keep track of the file and expected test results.
	fileLookup := map[string]t_TestFile{}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(path, file.Name())
		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		name := file.Name()

		// Check for .input.json suffix
		if strings.HasSuffix(name, ".input.json") {
			baseName := strings.TrimSuffix(name, ".input.json")
			testFile := fileLookup[baseName]
			testFile.Filename = filename
			testFile.InputJSON = strings.TrimSpace(string(content))
			fileLookup[baseName] = testFile
			continue
		}

		// Check for .expected.json suffix
		if strings.HasSuffix(name, ".expected.json") {
			baseName := strings.TrimSuffix(name, ".expected.json")
			testFile := fileLookup[baseName]
			testFile.ExpectedJSON = strings.TrimSpace(string(content))
			fileLookup[baseName] = testFile
			continue
		}

		// Unknown file type
		return nil, errors.Errorf("unexpected file in test folder: '%s' (expected .input.json or .expected.json)", filename)
	}

	// Verify all files have both input and expected
	for baseName, testFile := range fileLookup {
		if testFile.Filename == "" {
			return nil, errors.Errorf("missing .input.json file for test '%s' in '%s'", baseName, path)
		}
		if testFile.ExpectedJSON == "" {
			return nil, errors.Errorf("missing .expected.json file for test '%s' in '%s'", baseName, path)
		}
		allFiles = append(allFiles, testFile)
	}

	// This is for a test, so there must always be files.
	if len(allFiles) == 0 {
		return nil, errors.Errorf("no test files found in: '%s'", path)
	}

	// Sort the tests so simpler tests fail first.
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Filename < allFiles[j].Filename
	})

	return allFiles, nil
}

// t_TestFileError represents a test file for error cases.
type t_TestFileError struct {
	Filename   string // The input JSON filename.
	InputJSON  string // The contents of the input JSON file.
	ErrorCode  int    // The expected error code.
	ErrorField string // The expected error field (optional).
}

// t_ContentsForAllErrorJSONFiles reads all JSON error test files from a directory.
// Files are expected to be named like:
// - 01_missing_name.err.json (the input file that should cause an error)
// The file should contain a special _expected_error field with the expected error code.
func t_ContentsForAllErrorJSONFiles(path string) (allFiles []t_TestFileError, err error) {
	files, err := os.ReadDir(path)
	if err != nil {
		// If directory doesn't exist, return empty (no error tests)
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(path, file.Name())
		name := file.Name()

		// Check for .err.json suffix
		if !strings.HasSuffix(name, ".err.json") {
			return nil, errors.Errorf("unexpected file in error test folder: '%s' (expected .err.json)", filename)
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		testFile := t_TestFileError{
			Filename:  filename,
			InputJSON: strings.TrimSpace(string(content)),
		}

		allFiles = append(allFiles, testFile)
	}

	// Sort the tests so simpler tests fail first.
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Filename < allFiles[j].Filename
	})

	return allFiles, nil
}
