// source_watcher.go
package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type SourceChangeHandler func(db *sql.DB, plantUmlBinaryPath, rootSourcePath, rootOutputPath, model string) (err error) // Function type for handling source changes.

const debounceDuration = 50 * time.Millisecond // Duration to wait before triggering the handler after the last change event.

func InitSourceWatcher(rootSourcePath, rootMdPath string, db *sql.DB, plantUmlBinaryPath string, handler SourceChangeHandler) { // Initializes the file watcher for the source directory and sets up event handling.
	watcher, err := fsnotify.NewWatcher() // Creates a file watcher to monitor changes in source files.
	if err != nil {                       // Checks for error in watcher creation.
		log.Fatal(err) // Logs and exits if watcher can't be created.
	}
	// Note: No defer watcher.Close() here because it needs to live for the duration of the program.
	// It will be closed when the program exits.

	go func() { // Starts a goroutine to handle watcher events asynchronously.
		timers := make(map[string]*time.Timer) // Map to hold debounce timers per model.
		for {                                  // Infinite loop to process events.
			select { // Multiplexes watcher channels.
			case event, ok := <-watcher.Events: // Receives file events.
				if !ok { // Checks if channel is closed.
					return // Exits goroutine if channel closed.
				}
				if event.Op&fsnotify.Write == fsnotify.Write { // Checks if the event is a file write (change).
					rel, err := filepath.Rel(rootSourcePath, event.Name) // Gets relative path from source root.
					if err == nil {                                      // Proceeds if no error in relative path.
						rel = filepath.ToSlash(rel)      // Normalizes path separators to slashes.
						parts := strings.Split(rel, "/") // Splits path into components (model/subdirs/file.ext).
						if len(parts) > 1 {              // Ensures at least model and some file.
							model := parts[0]                             // Extracts model name (first subdir).
							file := parts[len(parts)-1]                   // Extracts the file name.
							suffix := strings.ToLower(filepath.Ext(file)) // Gets file extension in lowercase.
							for _, ext := range sourceExtensions {        // Checks if the suffix matches any source extension.
								if suffix == ext { // If matches, debounce the handler call.
									if t, ok := timers[model]; ok { // If timer exists, reset it.
										t.Reset(debounceDuration) // Resets the timer to wait again.
									} else { // If no timer, create one.
										timers[model] = time.AfterFunc(debounceDuration, func() { // Schedules the handler after debounce.
											err := handler(db, plantUmlBinaryPath, rootSourcePath, rootMdPath, model) // Calls the provided handler for the model.
											if err != nil {                                                           // Checks for error in handling.
												log.Println("Error handling source change:", err) // Logs the error.
											}
											delete(timers, model) // Removes the timer entry after handling.
										})
									}
									break // Breaks after handling to avoid multiple calls if extensions overlap.
								}
							}
						}
					}
				}
			case err, ok := <-watcher.Errors: // Receives errors from watcher.
				if !ok { // Checks if error channel closed.
					return // Exits if closed.
				}
				log.Println("error:", err) // Logs the error.
			}
		}
	}()

	err = filepath.Walk(rootSourcePath, func(path string, info os.FileInfo, err error) error { // Walks the source directory tree.
		if err != nil { // Handles walk errors.
			return err // Propagates error.
		}
		if info.IsDir() { // Checks if current path is a directory.
			return watcher.Add(path) // Adds directory to watcher for monitoring (handles deep nesting).
		}
		return nil // Continues walking.
	})
	if err != nil { // Checks for walk errors.
		log.Fatal(err) // Logs and exits if error.
	}
}
