package parser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type t_FileContents struct {
	Filename     string // The filename.
	Contents     string // The contents of the file.
	Json         string // The JSON it should convert to in an object.
	JsonChildren string // The JSON for child objects (optional, e.g., associations). Empty if no _children.json file exists.
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

			filename := filepath.Join(path, strings.ToLower(file.Name()))
			content, err := os.ReadFile(filename)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			switch filepath.Ext(filename) {
			case ".md":
				// name after the Markdown file.
				nameNoExtension := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				fileContents := fileLookup[nameNoExtension]
				fileContents.Filename = filename
				fileContents.Contents = strings.TrimSpace(string(content))
				fileLookup[nameNoExtension] = fileContents

			case ".json":
				// Check if this is a children file (ends with _children.json).
				nameNoExtension := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				if strings.HasSuffix(nameNoExtension, "_children") {
					// This is a children file - link to the base test file.
					baseName := strings.TrimSuffix(nameNoExtension, "_children")
					fileContents := fileLookup[baseName]
					fileContents.JsonChildren = strings.TrimSpace(string(content))
					fileLookup[baseName] = fileContents
				} else {
					// Add the expected golang object data as json.
					fileContents := fileLookup[nameNoExtension]
					fileContents.Json = strings.TrimSpace(string(content))
					fileLookup[nameNoExtension] = fileContents
				}

			default:
				return nil, errors.WithStack(errors.Errorf(`Non-test file found in test folder: '%s'`, filename))

			}
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
