// main.go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/database"
)

func main() { // Entry point of the program.
	var err error
	var port string               // Local variable for the server port.
	var rootMdPath string         // Local variable for the models directory path.
	var rootSourcePath string     // Local variable for the source directory path.
	var plantUmlBinaryPath string // Local variable for the PlantUML binary path.
	var noDb bool                 // Local variable indicating if there should be no database exercised.

	flag.StringVar(&port, "port", "8080", "port to listen on")                                                        // Defines flag for port with default.
	flag.StringVar(&rootMdPath, "rootmd", "", "models directory")                                                     // Defines flag for models directory with no default.
	flag.StringVar(&rootSourcePath, "rootsource", "", "source directory")                                             // Defines flag for source directory with no default.
	flag.StringVar(&plantUmlBinaryPath, "plantuml", "", "PlantUML binary path")                                       // Defines flag for PlantUML binary path.
	flag.BoolVar(&noDb, "nodb", false, "Whether to pass all the data through a parsed database struture to validate") // Defines flag for whether to exercise struture through a database.
	flag.Parse()                                                                                                      // Parses command-line flags into variables.

	if rootMdPath == "" {
		log.Fatal("models directory is required")
	}
	if rootSourcePath == "" {
		log.Fatal("source directory is required")
	}

	fmt.Printf("Starting server on port :%s with models directory %s\n", port, rootMdPath) // Prints startup configuration for user awareness.

	// Write the requirements to the database to ensure the data is well-formed.
	var db *sql.DB
	if !noDb {
		db, err = database.NewDb()
		if err != nil {
			fmt.Printf("%+v\n\n", err)
			os.Exit(1)
		}
	}

	InitWatcher(rootMdPath)                                                                   // Initializes the file watcher for monitoring changes in the models directory.
	InitSourceWatcher(rootSourcePath, rootMdPath, db, plantUmlBinaryPath, handleSourceChange) // Initializes the file watcher for monitoring changes in the source directory.

	mux := http.NewServeMux()                     // Creates a new HTTP request multiplexer.
	mux.HandleFunc("/", Handler(rootMdPath))      // Registers root handler for all requests.
	mux.HandleFunc("/events/", eventsHandler)     // Registers SSE handler for events.
	log.Fatal(http.ListenAndServe(":"+port, mux)) // Starts the HTTP server on the specified port and logs fatal if fails.
}
