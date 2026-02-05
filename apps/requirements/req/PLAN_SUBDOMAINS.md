# Plan: Add Subdomain Folders Under Domain Directories

## Overview

Currently, when parsing the model directory structure, all classes, use cases, and generalizations within a domain are automatically placed into a "default" subdomain. This plan describes how to add support for explicit subdomain folders under domain directories, allowing users to organize their domain contents into logical subdomains.

## Current State

### Directory Structure (Current)
```
model_root/
├── this.model
├── actors/
│   └── *.actor
└── domain_name/
    ├── this.domain
    ├── classes/
    │   ├── *.class
    │   └── *.generalization
    └── use_cases/
        └── *.uc
```

### Current Behavior
- All entities in a domain go into an auto-created "default" subdomain
- The `fileToParse` struct extracts domain from the first path component
- No subdomain information is extracted from the file path
- The write function merges all subdomain contents back into domain-level directories

### Key Files
- [file_to_parse.go](internal/parser/file_to_parse.go) - File discovery and path analysis
- [top_level_parse.go](internal/parser/top_level_parse.go) - Main parsing orchestration
- [top_level_write.go](internal/parser/top_level_write.go) - Writing model to filesystem
- [subdomain.go](internal/parser/subdomain.go) - Subdomain parsing (currently unused for file-based parsing)

## Proposed Directory Structure

```
model_root/
├── this.model
├── actors/
│   └── *.actor
└── domain_name/
    ├── this.domain
    ├── classes/                    # Default subdomain classes (backward compatible)
    │   ├── *.class
    │   └── *.generalization
    ├── use_cases/                  # Default subdomain use cases (backward compatible)
    │   └── *.uc
    └── subdomain_name/             # Explicit subdomain folder
        ├── this.subdomain          # Subdomain definition file
        ├── classes/
        │   ├── *.class
        │   └── *.generalization
        └── use_cases/
            └── *.uc
```

## Implementation Plan

### Phase 1: Update File Discovery (`file_to_parse.go`)

#### 1.1 Add Subdomain Extension Constant
```go
const (
    _EXT_SUBDOMAIN = ".subdomain"
)
```

#### 1.2 Add Subdomain to Sort Order
```go
var _extSortValue = map[string]int{
    _EXT_MODEL:          10,
    _EXT_ACTOR:          9,
    _EXT_DOMAIN:         8,
    _EXT_SUBDOMAIN:      7,  // Parse subdomains before their contents
    _EXT_GENERALIZATION: 6,
    _EXT_CLASS:          5,
    _EXT_USE_CASE:       3,
}
```

#### 1.3 Extend `fileToParse` Struct
```go
type fileToParse struct {
    ModelPath      string
    PathRel        string
    PathAbs        string
    FileType       string
    Generalization string
    Actor          string
    Domain         string
    Subdomain      string  // NEW: The subdomain (if entity is in a subdomain folder)
    Class          string
    UseCase        string
}
```

#### 1.4 Update `newFileToParse` Logic

The path analysis needs to detect subdomain folders. The logic should be:

1. First path component (except "actors/") = domain
2. If second path component is NOT "classes/" or "use_cases/", and contains a `this.subdomain` file, then it's a subdomain folder
3. Classes/use cases in subdomain folders get their subdomain field set

```go
// Detect subdomain from path structure:
// domain/subdomain_name/classes/foo.class -> subdomain = "subdomain_name"
// domain/classes/foo.class -> subdomain = "" (default)
subdomain := ""
if fileType != _EXT_MODEL && fileType != _EXT_ACTOR && fileType != _EXT_DOMAIN {
    pathRelParts := strings.Split(pathRel, string(filepath.Separator))
    if len(pathRelParts) >= 3 && pathRelParts[0] != _PATH_ACTORS {
        // Check if second component is a subdomain (not "classes" or "use_cases")
        secondPart := pathRelParts[1]
        if secondPart != "classes" && secondPart != "use_cases" {
            subdomain = secondPart
        }
    }
}

// Handle .subdomain files
if fileType == _EXT_SUBDOMAIN {
    pathRelParts := strings.Split(pathRel, string(filepath.Separator))
    if len(pathRelParts) >= 2 {
        subdomain = pathRelParts[1]  // domain/subdomain_name/this.subdomain
    }
}
```

#### 1.5 Update Validation
```go
validation.Field(&toParse.FileType, validation.Required, validation.In(
    _EXT_MODEL, _EXT_GENERALIZATION, _EXT_ACTOR, _EXT_DOMAIN,
    _EXT_SUBDOMAIN,  // NEW
    _EXT_CLASS, _EXT_USE_CASE,
)),
```

### Phase 2: Update Parsing Logic (`top_level_parse.go`)

#### 2.1 Add Subdomain Tracking
```go
// Track subdomains by their composite key (domain + subdomain name)
subdomainKeysByPath := map[string]identity.Key{}  // "domain/subdomain" -> subdomain key
```

#### 2.2 Add Case for Parsing `.subdomain` Files
```go
case _EXT_SUBDOMAIN:
    // Find the domain for this subdomain
    domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
    if !ok {
        return req_model.Model{}, errors.Errorf("domain '%s' not found for subdomain '%s'",
            toParseFile.Domain, toParseFile.Subdomain)
    }
    domain := model.Domains[domainKey]

    subdomain, err := parseSubdomain(domainKey, toParseFile.Subdomain, toParseFile.PathRel, contents)
    if err != nil {
        return req_model.Model{}, err
    }

    domain.Subdomains[subdomain.Key] = subdomain
    model.Domains[domainKey] = domain
    subdomainKeysByPath[toParseFile.Domain + "/" + toParseFile.Subdomain] = subdomain.Key
```

#### 2.3 Update Class/UseCase/Generalization Parsing

Modify the existing cases to use explicit subdomain when specified:

```go
case _EXT_CLASS:
    domainKey, ok := domainKeysBySubKey[toParseFile.Domain]
    if !ok {
        return req_model.Model{}, errors.Errorf("domain '%s' not found for class '%s'",
            toParseFile.Domain, toParseFile.Class)
    }
    domain := model.Domains[domainKey]

    // Determine which subdomain to use
    var subdomainKey identity.Key
    if toParseFile.Subdomain != "" {
        // Use explicit subdomain
        pathKey := toParseFile.Domain + "/" + toParseFile.Subdomain
        var ok bool
        subdomainKey, ok = subdomainKeysByPath[pathKey]
        if !ok {
            return req_model.Model{}, errors.Errorf("subdomain '%s' not found in domain '%s' for class '%s'",
                toParseFile.Subdomain, toParseFile.Domain, toParseFile.Class)
        }
    } else {
        // Use default subdomain
        subdomainKey, err = identity.NewSubdomainKey(domainKey, "default")
        if err != nil {
            return req_model.Model{}, errors.WithStack(err)
        }
    }

    subdomain, ok := domain.Subdomains[subdomainKey]
    if !ok {
        return req_model.Model{}, errors.Errorf("subdomain not found for class '%s'", toParseFile.Class)
    }

    // ... rest of class parsing ...
```

Apply similar changes to `_EXT_USE_CASE` and `_EXT_GENERALIZATION` cases.

### Phase 3: Update Write Logic (`top_level_write.go`)

#### 3.1 Update `writeDomain` Function

Modify to write subdomains as separate directories when they are not the default:

```go
func writeDomain(outputPath string, domain model_domain.Domain, ...) error {
    // Create domain directory
    domainDir := filepath.Join(outputPath, domain.Key.SubKey())
    if err := os.MkdirAll(domainDir, 0755); err != nil {
        return errors.Wrapf(err, "failed to create domain directory: %s", domain.Key.SubKey())
    }

    // Write this.domain file
    // ...

    // Process subdomains
    for _, subdomain := range domain.Subdomains {
        if subdomain.Key.SubKey() == "default" {
            // Default subdomain: write directly under domain directory (backward compatible)
            if err := writeSubdomainContents(domainDir, subdomain, classAssociations); err != nil {
                return err
            }
        } else {
            // Explicit subdomain: create subdomain directory
            if err := writeExplicitSubdomain(domainDir, subdomain, classAssociations); err != nil {
                return err
            }
        }
    }

    return nil
}
```

#### 3.2 Add `writeExplicitSubdomain` Function

```go
func writeExplicitSubdomain(domainDir string, subdomain model_domain.Subdomain, classAssociations map[identity.Key]model_class.Association) error {
    // Create subdomain directory
    subdomainDir := filepath.Join(domainDir, subdomain.Key.SubKey())
    if err := os.MkdirAll(subdomainDir, 0755); err != nil {
        return errors.Wrapf(err, "failed to create subdomain directory: %s", subdomain.Key.SubKey())
    }

    // Write this.subdomain file
    subdomainContent := generateSubdomainContent(subdomain)
    subdomainPath := filepath.Join(subdomainDir, "this"+_EXT_SUBDOMAIN)
    if err := os.WriteFile(subdomainPath, []byte(subdomainContent), 0644); err != nil {
        return errors.Wrapf(err, "failed to write subdomain file: %s", subdomain.Key.SubKey())
    }

    // Write contents under subdomain directory
    if err := writeSubdomainContents(subdomainDir, subdomain, classAssociations); err != nil {
        return err
    }

    return nil
}
```

### Phase 4: Testing

#### 4.1 Unit Tests

Create test cases in `file_to_parse_test.go`:
- Test subdomain detection from various path structures
- Test that "classes" and "use_cases" are not mistaken for subdomains
- Test backward compatibility with no subdomains

Create test cases in `top_level_parse_test.go`:
- Test parsing a model with explicit subdomains
- Test parsing a model with only default subdomain (backward compatible)
- Test error when subdomain file missing but subdomain folder has contents

Create test cases in `top_level_write_test.go`:
- Test writing a model with explicit subdomains
- Test writing a model with only default subdomain
- Test round-trip: parse -> write -> parse produces same model

#### 4.2 Integration Tests

Add example model with subdomains to test fixtures:
```
test_models/with_subdomains/
├── this.model
└── order_fulfillment/
    ├── this.domain
    ├── classes/                    # Default subdomain
    │   └── customer.class
    ├── order_management/           # Explicit subdomain
    │   ├── this.subdomain
    │   ├── classes/
    │   │   └── book_order.class
    │   └── use_cases/
    │       └── place_order.uc
    └── inventory/                  # Another explicit subdomain
        ├── this.subdomain
        ├── classes/
        │   └── stock_item.class
        └── use_cases/
            └── check_inventory.uc
```

### Phase 5: Update Example Models

Update `/data/examples/requirements/req/models/web_books` to demonstrate subdomain usage (optional, could be a separate task).

## Backward Compatibility

This implementation maintains full backward compatibility:

1. **Existing models without subdomain folders**: Continue to work exactly as before. All entities go into the "default" subdomain.

2. **Mixed models**: Domains can have both:
   - Entities directly in `classes/` and `use_cases/` (default subdomain)
   - Entities in explicit subdomain folders

3. **Write output**: Models with only default subdomains are written in the original flat structure under the domain directory.

## Edge Cases to Handle

1. **Empty subdomain folder**: A folder exists but has no `this.subdomain` file
   - Decision: Treat as regular folder, not a subdomain (error if it contains .class/.uc files)

2. **Subdomain folder with reserved name**: `classes` or `use_cases` as subdomain name
   - Decision: Disallow - these are reserved folder names

3. **Nested subdomains**: `domain/sub1/sub2/classes/`
   - Decision: Not supported - only one level of subdomain nesting allowed

4. **Class name collisions**: Same class name in different subdomains
   - Already handled: Classes are keyed by full path including subdomain

## File Changes Summary

| File | Changes |
|------|---------|
| `file_to_parse.go` | Add `_EXT_SUBDOMAIN`, `Subdomain` field, path analysis logic |
| `top_level_parse.go` | Add `_EXT_SUBDOMAIN` case, subdomain routing for classes/use cases |
| `top_level_write.go` | Add `writeExplicitSubdomain`, modify `writeDomain` |
| `subdomain.go` | No changes needed (already has `parseSubdomain` and `generateSubdomainContent`) |

## Implementation Order

1. `file_to_parse.go` - Add subdomain detection
2. `top_level_parse.go` - Add subdomain parsing and routing
3. `top_level_write.go` - Add subdomain directory writing
4. Tests for each component
5. Integration tests with example models
