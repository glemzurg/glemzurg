package httpserver

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/parser_ai"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model"
)

// Input format constants
const (
	InputFormatDataYAML = "data/yaml"
	InputFormatAIJSON   = "ai/json"
)

// SourceExtensionsYAML defines file extensions for YAML format.
var SourceExtensionsYAML = []string{".class", ".domain", ".model"}

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

	for _, ext := range extensions {
		if suffix == ext {
			sw.debounceUpdate()
			break
		}
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
func (sw *SourceWatcher) updateModel() error {
	if sw.inputFormat == InputFormatAIJSON {
		return sw.updateModelFromJSON()
	}
	return sw.updateModelFromYAML()
}

// updateModelFromJSON parses a model from parser_ai JSON format.
func (sw *SourceWatcher) updateModelFromJSON() error {
	// Read the model tree from the directory
	inputModel, err := parser_ai.readModelTree(sw.modelPath)
	if err != nil {
		return err
	}

	// Convert to req_model.Model
	parsedModel, err := parser_ai.ConvertToModel(inputModel, sw.modelName)
	if err != nil {
		return err
	}

	err = sw.store.SetModel(sw.modelName, parsedModel)
	if err != nil {
		return err
	}

	sw.server.NotifyModel(sw.modelName)
	return nil
}

// updateModelFromYAML parses a model from YAML source files.
func (sw *SourceWatcher) updateModelFromYAML() error {
	parsedModel, err := parser.Parse(sw.modelPath)
	if err != nil {
		return err
	}

	err = sw.store.SetModel(sw.modelName, &parsedModel)
	if err != nil {
		return err
	}

	sw.server.NotifyModel(sw.modelName)
	return nil
}

// LoadModel loads the model from the source path into the store.
func (sw *SourceWatcher) LoadModel() error {
	return sw.updateModel()
}

// LoadModelFromData loads a model from a pre-parsed req_model.Model.
func LoadModelFromData(store *ModelStore, server *Server, name string, model *req_model.Model) error {
	err := store.SetModel(name, model)
	if err != nil {
		return err
	}
	if server != nil {
		server.NotifyModel(name)
	}
	return nil
}
