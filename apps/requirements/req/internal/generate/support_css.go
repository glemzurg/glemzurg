package generate

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

.unfinished-notes-block {
    color: maroon;               /* text color */
    background-color: lightgrey; /* background */
    border-left: 5px solid red;
    padding: 15px;
    font-family: monospace;
}

.unfinished-notes-glyph {
  color: red;
}

.parse-error-marker {
  color: #cc0000;
  font-weight: bold;
}

.parse-error-spec {
  color: #cc0000;
  font-weight: bold;
  text-decoration: underline wavy #cc0000;
}

.parse-error-banner {
  border: 2px solid #cc0000;
  background-color: #fff5f5;
  padding: 12px 16px;
  margin-bottom: 1.5em;
}
`
