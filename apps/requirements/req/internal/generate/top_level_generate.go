package generate

import (
	"fmt"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"

	"github.com/pkg/errors"
)

// GenerateMdFromModel generates markdown documentation from an already-parsed model.
func GenerateMdFromModel(debug bool, outputPath string, parsedModel req_model.Model) (err error) {

	// Create necessary output paths if we don't have them.
	if err = createMissingPaths([]string{outputPath}); err != nil {
		return err
	}

	fmt.Println()

	// Use FileWriter to write to filesystem via ContentWriter interface.
	writer := NewFileWriter(outputPath)
	err = GenerateMdToWriter(debug, parsedModel, writer)
	if err != nil {
		return err
	}

	fmt.Println()

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
