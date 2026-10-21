// Package load reads a requirements JSON tree and its PostgreSQL annotations.
package load

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/annot"
	"github.com/glemzurg/glemzurg/apps/code/postgresql/internal/walk"
	"github.com/glemzurg/glemzurg/apps/requirements/req/pkg/parser_json"
)

// Load reads the model directory and the annotation directory, then walks the model.
// annotDir may be the same directory as modelDir when notes sit beside the model files.
func Load(modelDir, annotDir string) (walk.Result, error) {
	model, err := parser_json.ReadModel(modelDir)
	if err != nil {
		return walk.Result{}, err
	}
	notes, err := annot.Read(annotDir)
	if err != nil {
		return walk.Result{}, err
	}
	if err := requireModelFiles(modelDir, notes); err != nil {
		return walk.Result{}, err
	}
	return walk.Walk(model, notes)
}

func requireModelFiles(modelDir string, notes annot.Set) error {
	var missing []string
	for _, rel := range notes.Paths() {
		info, err := os.Stat(filepath.Join(modelDir, filepath.FromSlash(rel)))
		if err != nil || info.IsDir() {
			missing = append(missing, rel)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("annotation model file missing: %s", strings.Join(missing, ", "))
}
