package parser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type t_FileContents struct {
	Filename string // The filename.
	Contents string // The contents of the file.
	Json     string // The JSON it should convert to in an object.
}

func (fc *t_FileContents) verify(fileLocation string) (err error) {
	if fc.Filename == "" {
		return errors.WithStack(errors.Errorf(`Missing filename '%s': '%+v'`, fileLocation, fc))
	}
	// We can have blanks for the other files.
	return nil
}

func t_ContentsForAllMdFiles(path string) (allFiles []t_FileContents, err error) {

	// Keep track of the file and expected test results.
	fileLookup := map[string]t_FileContents{}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, file := range files {
		if !file.IsDir() {

			nameNoExtension := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			fileContents := fileLookup[nameNoExtension]

			filename := filepath.Join(path, strings.ToLower(file.Name()))
			content, err := os.ReadFile(filename)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			switch filepath.Ext(filename) {
			case ".md":
				// name after the Markdown file.
				fileContents.Filename = filename
				fileContents.Contents = strings.TrimSpace(string(content))

			case ".json":
				// Add the expected golang object data as json.
				fileContents.Json = strings.TrimSpace(string(content))

			default:
				return nil, errors.WithStack(errors.Errorf(`Non-test file found in test folder: '%s'`, filename))

			}

			fileLookup[nameNoExtension] = fileContents
		}
	}

	// Everything in a simple format.
	for baseFilename, fileContents := range fileLookup {

		// Any parts missing is failure.
		if err := fileContents.verify(path + "/" + baseFilename); err != nil {
			return nil, errors.WithStack(err)
		}

		allFiles = append(allFiles, fileContents)
	}

	// This is for a test, so there must always be files.
	if len(allFiles) == 0 {
		return nil, errors.WithStack(errors.Errorf(`No test files found in: '%s'`, path))
	}

	// Sort the test so that simpler tests will fail first in the test cases.
	// Sort
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Filename < allFiles[j].Filename
	})

	return allFiles, nil
}
