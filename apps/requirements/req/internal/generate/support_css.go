package generate

import "path/filepath"

const _MD_CSS = `
h4, h5, h6 {
    text-decoration: underline;
}

h1 {
  border-bottom: 4px solid #000; /* Adjust color, thickness, and style as needed */
}

h2, h3 {
  border-bottom: 1px solid #000; /* Adjust color, thickness, and style as needed */
}

table {
  border-collapse: collapse;
}

th, td {
  border: 1px solid #ccc;
  padding: 8px;
}
`

func generateSupportCss(outputPath string) (err error) {

	// Generate css.
	if err = writeFile(filepath.Join(outputPath, "style.css"), _MD_CSS); err != nil {
		return err
	}

	return nil
}
