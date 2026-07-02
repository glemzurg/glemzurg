package httpserver

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_human"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/perftrack"
)

// Input format constants.
const (
	InputFormatDataYAML = "data/yaml"
	InputFormatAIJSON   = "ai/json"
)

// SourceExtensionsYAML defines file extensions for YAML format.
var SourceExtensionsYAML = []string{".actor", ".class", ".domain", ".generalization", ".model", ".subdomain", ".uc"}

// SourceExtensionsJSON defines file extensions for JSON format.
var SourceExtensionsJSON = []string{".json"}

const debounceDuration = 50 * time.Millisecond

// SourceWatcher watches a single model's source directory for changes and updates the model store.
type SourceWatcher struct {
	modelPath   string
	modelName   string
	inputFormat string
	store       *ModelStore
	server      *Server
	watcher     *fsnotify.Watcher
	timer       *time.Timer
}

// NewSourceWatcher creates a new source watcher for a single model.
func NewSourceWatcher(modelPath, inputFormat string, store *ModelStore, server *Server) (*SourceWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Extract model name from path
	modelName := filepath.Base(modelPath)

	sw := &SourceWatcher{
		modelPath:   modelPath,
		modelName:   modelName,
		inputFormat: inputFormat,
		store:       store,
		server:      server,
		watcher:     watcher,
	}

	return sw, nil
}

// Start begins watching for source file changes.
func (sw *SourceWatcher) Start() error {
	go sw.eventLoop()

	return filepath.Walk(sw.modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return sw.watcher.Add(path)
		}
		return nil
	})
}

// Close stops the watcher.
func (sw *SourceWatcher) Close() error {
	return sw.watcher.Close()
}

// eventLoop handles file system events.
func (sw *SourceWatcher) eventLoop() {
	for {
		select {
		case event, ok := <-sw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create != 0 {
				sw.handleCreateEvent(event.Name)
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				sw.handleFileChange(event.Name)
			}
		case err, ok := <-sw.watcher.Errors:
			if !ok {
				return
			}
			log.Println("watcher error:", err)
		}
	}
}

// handleCreateEvent adds newly created directories to the watcher.
func (sw *SourceWatcher) handleCreateEvent(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		err = sw.watcher.Add(path)
		if err != nil {
			log.Println("watcher add error:", err)
		}
	}
}

// handleFileChange processes a file change event.
func (sw *SourceWatcher) handleFileChange(path string) {
	file := filepath.Base(path)
	suffix := strings.ToLower(filepath.Ext(file))

	// Select extensions based on input format
	var extensions []string
	if sw.inputFormat == InputFormatAIJSON {
		extensions = SourceExtensionsJSON
	} else {
		extensions = SourceExtensionsYAML
	}

	if slices.Contains(extensions, suffix) {
		sw.debounceUpdate()
	}
}

// debounceUpdate delays the model update to avoid rapid successive updates.
func (sw *SourceWatcher) debounceUpdate() {
	if sw.timer != nil {
		sw.timer.Reset(debounceDuration)
	} else {
		sw.timer = time.AfterFunc(debounceDuration, func() {
			err := sw.updateModel()
			if err != nil {
				log.Println("Error updating model:", err)
			}
			sw.timer = nil
		})
	}
}

// updateModel parses the source files and updates the model store.
//
// On any failure (parse error, or generation error inside SetModel) it records
// the error in the store and notifies connected browsers, so the web display
// reloads onto a red-bold error page instead of silently keeping stale content.
// The error is still returned for logging.
func (sw *SourceWatcher) updateModel() error {
	tracker := perftrack.New("model.reload " + sw.modelName)
	defer tracker.LogIfSlow()

	var err error
	if sw.inputFormat == InputFormatAIJSON {
		err = sw.updateModelFromJSON(tracker)
	} else {
		err = sw.updateModelFromYAML(tracker)
	}
	if err != nil {
		sw.store.SetModelError(sw.modelName, err)
		sw.server.NotifyModel(sw.modelName)
		return err
	}
	return nil
}

// updateModelFromJSON parses a model from parser_ai JSON format.
func (sw *SourceWatcher) updateModelFromJSON(tracker *perftrack.Tracker) error {
	var parsedModel core.Model
	var err error
	perftrack.RunOn(tracker, "parse.json", func() {
		parsedModel, err = parser_ai.ReadModel(sw.modelPath)
	})
	if err != nil {
		return err
	}

	perftrack.RunOn(tracker, "store.setModel", func() {
		err = sw.store.SetModelTracked(sw.modelName, &parsedModel, nil, tracker)
	})
	if err != nil {
		return err
	}

	sw.server.NotifyModel(sw.modelName)
	return nil
}

// updateModelFromYAML parses a model from YAML source files.
//
// A parse failure in a single .class file is not a catastrophic error: Parse
// returns the partial model plus the per-class failures, which generation turns
// into red-bold error blocks on those classes' pages.
func (sw *SourceWatcher) updateModelFromYAML(tracker *perftrack.Tracker) error {
	var parsedModel core.Model
	var failures []parser_human.ParseFailure
	var err error
	perftrack.RunOn(tracker, "parse.yaml", func() {
		parsedModel, failures, err = parser_human.Parse(sw.modelPath)
	})
	if err != nil {
		return err
	}

	perftrack.RunOn(tracker, "store.setModel", func() {
		err = sw.store.SetModelTracked(sw.modelName, &parsedModel, classErrorMap(failures), tracker)
	})
	if err != nil {
		return err
	}

	sw.server.NotifyModel(sw.modelName)
	return nil
}

// classErrorMap converts parser failures into a class-key -> error-message map
// for the generator. Returns nil when there are no failures.
func classErrorMap(failures []parser_human.ParseFailure) map[string]string {
	if len(failures) == 0 {
		return nil
	}
	m := make(map[string]string, len(failures))
	for _, f := range failures {
		log.Printf("parse failure: %s: %s", f.Path, f.Err)
		m[f.ClassKey.String()] = f.Err
	}
	return m
}

// LoadModel loads the model from the source path into the store.
func (sw *SourceWatcher) LoadModel() error {
	return sw.updateModel()
}

// LoadModelFromData loads a model from a pre-parsed core.Model.
func LoadModelFromData(store *ModelStore, server *Server, name string, model *core.Model) error {
	err := store.SetModel(name, model, nil)
	if err != nil {
		return err
	}
	if server != nil {
		server.NotifyModel(name)
	}
	return nil
}
