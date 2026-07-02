package generate

import (
	"log"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"

	"github.com/pkg/errors"
)

// GenerateMdFromModel generates markdown documentation from an already-parsed model.
//
// classErrors maps a class key string to a parse-error message; those classes'
// pages are rendered as red-bold error blocks. Pass nil when there are none.
func GenerateMdFromModel(outputPath string, parsedModel core.Model, classErrors map[string]string) (err error) { //nolint:revive // public API name
	// Create necessary output paths if we don't have them.
	if err = createMissingPaths([]string{outputPath}); err != nil {
		return err
	}

	log.Println()

	// Use FileWriter to write to filesystem via ContentWriter interface.
	writer := NewFileWriter(outputPath)
	err = GenerateMdToWriter(parsedModel, writer, classErrors)
	if err != nil {
		return err
	}

	log.Println()

	return nil
}

func createMissingPaths(paths []string) (err error) {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				err := os.Mkdir(path, 0755)
				if err != nil {
					return errors.WithStack(err)
				}
			} else {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}
