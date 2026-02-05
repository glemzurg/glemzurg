package generate

import (
	"fmt"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"

	"github.com/pkg/errors"
)

// GenerateMdFromModel generates markdown documentation from an already-parsed model.
func GenerateMdFromModel(debug bool, outputPath string, parsedModel req_model.Model) (err error) {

	// Create necessary output paths if we don't have them.
	if err = createMissingPaths([]string{outputPath}); err != nil {
		return err
	}

	// Create the flattened requirements from the model.
	reqs := req_flat.NewRequirements(parsedModel)

	// Prepare the convenience structures inside.
	reqs.PrepLookups()

	// Output the files.
	err = generateFiles(debug, outputPath, reqs)
	if err != nil {
		return err
	}

	return nil
}

func generateFiles(debug bool, outputPath string, reqs *req_flat.Requirements) (err error) {

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

	// Generate subdomain files (only for domains with multiple subdomains).
	err = generateSubdomainFiles(debug, outputPath, reqs)
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
