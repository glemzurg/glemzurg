package parser_ai

import (
	"fmt"
	"os"
	"path/filepath"
)

// domainLevelContentDirs lists directories that parser_human may place directly under a
// domain folder; parser_ai must reject them and require subdomains/{subdomain}/ instead.
var domainLevelContentDirs = []string{"classes", "use_cases"}

// rejectDomainLevelModelContent reports when a domain directory uses the human YAML layout
// (classes/ or use_cases/ at the domain root) instead of subdomains/{subdomain}/.
func rejectDomainLevelModelContent(domainDir string) error {
	domainFile := filepath.Join(domainDir, "domain.json")
	for _, dirName := range domainLevelContentDirs {
		childPath := filepath.Join(domainDir, dirName)
		info, err := os.Stat(childPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if !info.IsDir() {
			continue
		}
		return NewParseError(
			ErrDomainDirInvalid,
			fmt.Sprintf("domain directory contains %q/ at the domain root; AI JSON requires subdomains/{subdomain}/%s/", dirName, dirName),
			domainFile,
		).WithHint(fmt.Sprintf("move content under domains/{domain}/subdomains/default/%s/ (or the appropriate subdomain)", dirName))
	}
	return nil
}
