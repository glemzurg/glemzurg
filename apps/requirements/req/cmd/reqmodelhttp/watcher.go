// watcher.go
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func InitWatcher(rootMdPath string) { // Initializes the file watcher and sets up event handling.
	watcher, err := fsnotify.NewWatcher() // Creates a file watcher to monitor changes in files.
	if err != nil {                       // Checks for error in watcher creation.
		log.Fatal(err) // Logs and exits if watcher can't be created.
	}
	// Note: No defer watcher.Close() here because it needs to live for the duration of the program.
	// It will be closed when the program exits.

	go func() { // Starts a goroutine to handle watcher events asynchronously.
		for { // Infinite loop to process events.
			select { // Multiplexes watcher channels.
			case event, ok := <-watcher.Events: // Receives file events.
				if !ok { // Checks if channel is closed.
					return // Exits goroutine if channel closed.
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 { // Checks if the event is a file write or create (to handle atomic replacements).
					rel, err := filepath.Rel(rootMdPath, event.Name) // Gets relative path from root.
					if err == nil {                                  // Proceeds if no error in relative path.
						rel = filepath.ToSlash(rel)      // Normalizes path separators to slashes.
						parts := strings.Split(rel, "/") // Splits path into components (model/file.ext).
						if len(parts) == 2 {             // Ensures it's model/file.
							model := parts[0]                             // Extracts model name.
							file := parts[1]                              // Extracts file name.
							suffix := strings.ToLower(filepath.Ext(file)) // Gets file extension in lowercase.
							if suffix == ".md" || suffix == ".svg" {      // Handles MD or SVG changes.
								base := strings.TrimSuffix(file, suffix) // Removes extension to get base name.
								mdKey := model + "/" + base + ".md"      // Constructs key for the corresponding MD file.
								if b, ok := brokers.Load(mdKey); ok {    // Loads broker if exists.
									b.(*broker).notifier <- []byte("refresh") // Sends refresh notification to specific page broker.
								}
							} else if suffix == ".css" && file == "style.css" { // Handles CSS changes, which affect all pages in model.
								entries, err := os.ReadDir(filepath.Join(rootMdPath, model)) // Reads all files in model directory.
								if err == nil {                                              // Proceeds if no read error.
									for _, e := range entries { // Iterates over entries.
										if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") { // Finds MD files.
											key := model + "/" + e.Name()       // Constructs key for each MD.
											if b, ok := brokers.Load(key); ok { // Loads broker if exists.
												b.(*broker).notifier <- []byte("refresh") // Sends refresh to each page broker.
											}
										}
									}
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

	err = filepath.Walk(rootMdPath, func(path string, info os.FileInfo, err error) error { // Walks the directory tree.
		if err != nil { // Handles walk errors.
			return err // Propagates error.
		}
		if info.IsDir() { // Checks if current path is a directory.
			return watcher.Add(path) // Adds directory to watcher for monitoring.
		}
		return nil // Continues walking.
	})
	if err != nil { // Checks for walk errors.
		log.Fatal(err) // Logs and exits if error.
	}
}
