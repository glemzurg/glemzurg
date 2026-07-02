package generate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// mermaidRenderHook renders Mermaid source to SVG bytes. Tests may replace it.
var mermaidRenderHook = renderMermaidToSVGWithMMDC

// SetMermaidRenderHook swaps the Mermaid-to-SVG renderer (tests only).
func SetMermaidRenderHook(hook func(string) ([]byte, error)) {
	if hook == nil {
		mermaidRenderHook = renderMermaidToSVGWithMMDC
		return
	}
	mermaidRenderHook = hook
}

func renderMermaidToSVG(source string) ([]byte, error) {
	return mermaidRenderHook(source)
}

func renderMermaidToSVGWithMMDC(source string) ([]byte, error) {
	mmdcPath, err := resolveMMDCBinary()
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "req-mermaid-*")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	inputPath := filepath.Join(tmpDir, "diagram.mmd")
	outputPath := filepath.Join(tmpDir, "diagram.svg")

	if err := os.WriteFile(inputPath, []byte(source), 0600); err != nil {
		return nil, errors.WithStack(err)
	}

	cmd := exec.CommandContext(context.Background(), mmdcPath,
		"-i", inputPath,
		"-o", outputPath,
		"-b", "transparent",
	)
	if chrome := strings.TrimSpace(os.Getenv("PUPPETEER_EXECUTABLE_PATH")); chrome != "" {
		cmd.Env = append(os.Environ(), "PUPPETEER_EXECUTABLE_PATH="+chrome)
	}
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return nil, errors.Wrapf(runErr, "mmdc failed: %s", msg)
		}
		return nil, errors.Wrap(runErr, "mmdc failed")
	}

	svg, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(svg) == 0 {
		return nil, errors.New("mmdc produced an empty SVG")
	}
	return svg, nil
}

func resolveMMDCBinary() (string, error) {
	if override := strings.TrimSpace(os.Getenv("REQ_MMDC")); override != "" {
		if _, err := os.Stat(override); err != nil { //nolint:gosec // REQ_MMDC is an explicit operator override
			return "", errors.Wrapf(err, "REQ_MMDC points to missing binary: %s", override)
		}
		return override, nil
	}

	path, err := exec.LookPath("mmdc")
	if err != nil {
		return "", errors.New(
			"mmdc not found: install @mermaid-js/mermaid-cli (mermaid-cli) for file markdown output, " +
				"then run `node $(npm root -g)/@mermaid-js/mermaid-cli/node_modules/puppeteer/install.mjs` " +
				"to fetch a headless browser, or set REQ_MMDC to the mmdc binary path",
		)
	}
	return path, nil
}
