package annot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Set is the PostgreSQL annotations found under one directory tree.
// Notes are keyed by the model file they describe.
type Set struct {
	byPath map[string]Note
}

// Read loads every .annot.json file under root.
// The key of each note is the sibling model path: class.annot.json describes class.json.
func Read(root string) (Set, error) {
	info, err := os.Stat(root)
	if err != nil {
		return Set{}, err
	}
	if !info.IsDir() {
		return Set{}, fmt.Errorf("%s is not a directory", root)
	}

	notes := make(map[string]Note)
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), Suffix) {
			return nil
		}
		return readOne(root, path, notes)
	})
	if err != nil {
		return Set{}, err
	}
	return Set{byPath: notes}, nil
}

func readOne(root, path string, notes map[string]Note) error {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	modelRel := strings.TrimSuffix(rel, Suffix) + ".json"
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	note, err := parseNote(modelRel, content)
	if err != nil {
		return err
	}
	if _, exists := notes[modelRel]; exists {
		return fmt.Errorf("duplicate annotation for %s", modelRel)
	}
	notes[modelRel] = note
	return nil
}

// Note returns the annotation for a model file path, using slash separators.
func (s Set) Note(relPath string) (Note, bool) {
	note, ok := s.byPath[relPath]
	return note, ok
}

// Paths returns the model file paths that have an annotation, in lexical order.
func (s Set) Paths() []string {
	paths := make([]string, 0, len(s.byPath))
	for rel := range s.byPath {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths
}
