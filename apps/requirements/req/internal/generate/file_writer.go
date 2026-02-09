package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// FileWriter implements ContentWriter for writing to the filesystem.
type FileWriter struct {
	outputPath string
}

// NewFileWriter creates a new FileWriter that writes to the given output path.
func NewFileWriter(outputPath string) *FileWriter {
	return &FileWriter{outputPath: outputPath}
}

// WriteMarkdown writes markdown content to a file.
func (fw *FileWriter) WriteMarkdown(filename string, content []byte) error {
	return fw.writeFile(filename, content)
}

// WriteSVG writes SVG content to a file.
func (fw *FileWriter) WriteSVG(filename string, content []byte) error {
	return fw.writeFile(filename, content)
}

// WriteCSS writes CSS content to style.css.
func (fw *FileWriter) WriteCSS(content []byte) error {
	return fw.writeFile("style.css", content)
}

// writeFile writes content to a file in the output path.
func (fw *FileWriter) writeFile(filename string, content []byte) error {
	fullPath := filepath.Join(fw.outputPath, filename)
	fmt.Println("WRITING:", fullPath)

	file, err := os.Create(fullPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
