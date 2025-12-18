package generate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/database"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/requirements"

	"github.com/pkg/errors"
)

func GenerateMd(debug bool, db *sql.DB, rootSourcePath, rootOutputPath, model string) (err error) {

	sourcePath := filepath.Join(rootSourcePath, model)
	outputPath := filepath.Join(rootOutputPath, model)

	// Create necessary output paths if we don't have them.
	if err = createMissingPaths([]string{outputPath}); err != nil {
		return err
	}

	// Create the new requirements.
	reqs, err := parser.Parse(sourcePath)
	if err != nil {
		return err
	}

	// We may not want to exercice through a database.
	if db != nil {
		log.Println("Exercising data model through database.")
		// Write the requirements to the database to ensure the data is well-formed.
		err = database.WriteRequirements(db, reqs)
		if err != nil {
			return err
		}
		// Read the model from the database to ensure we can get it back out correctly.
		reqs, err = database.ReadRequirements(db, reqs.Model.Key)
		if err != nil {
			return err
		}
	}

	// Prepare the convenience structures inside.
	reqs.PrepLookups()

	// Output the files.
	err = generateFiles(debug, outputPath, reqs)
	if err != nil {
		return err
	}

	return nil
}

func generateFiles(debug bool, outputPath string, reqs requirements.Requirements) (err error) {

	fmt.Println()

	// Create the necessary support images for graphs.
	err = generateSupportImages(outputPath)
	if err != nil {
		return err
	}

	// Create the necessary css for md.
	err = generateSupportCss(outputPath)
	if err != nil {
		return err
	}

	// Build up the nested data for easy templates.

	// Generate model files.
	err = generateModelFiles(debug, outputPath, reqs)
	if err != nil {
		return err
	}

	// Generate actor files.
	err = generateActorFiles(outputPath, reqs)
	if err != nil {
		return err
	}

	// Generate domain files.
	err = generateDomainFiles(debug, outputPath, reqs)
	if err != nil {
		return err
	}

	// Generate class files.
	err = generateClassFiles(debug, outputPath, reqs)
	if err != nil {
		return err
	}

	// Generate use case files.
	err = generateUseCaseFiles(outputPath, reqs)
	if err != nil {
		return err
	}

	// Generate scenario files.
	err = generateScenarioFiles(outputPath, reqs)
	if err != nil {
		return err
	}

	fmt.Println()

	return nil
}

func writeFile(filename, contents string) (err error) {

	fmt.Println("WRITING:", filename)

	file, err := os.Create(filename)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	_, err = file.WriteString(contents)
	if err != nil {
		return errors.WithStack(err)
	}
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
