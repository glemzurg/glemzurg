// Package parser_json dumps a core.Model as one JSON document and reads that
// document back. The document is a short-lived pipe for other tools. It is
// regenerated from the current model and is not an authoring format.
package parser_json

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// formatCorePipe distinguishes this document from persistent AI JSON files.
// It names the current pipe shape and does not promise compatibility.
const formatCorePipe = "core-pipe"

// Write encodes model as one core-pipe document. The model is emitted as given;
// callers pass a model they already trust.
func Write(model core.Model, w io.Writer) error {
	doc, err := fromModel(model)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(doc); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Read decodes one core-pipe document into a model and validates it.
// Parsed expression trees and data-type trees are left for constructors to
// rebuild from the authored text, so a caller that needs those trees parses
// after Read.
func Read(r io.Reader) (core.Model, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var doc pipeModel
	if err := dec.Decode(&doc); err != nil {
		return core.Model{}, errors.Wrap(err, "core-pipe JSON")
	}
	if err := rejectTrailing(dec); err != nil {
		return core.Model{}, err
	}

	model, err := toModel(doc)
	if err != nil {
		return core.Model{}, err
	}
	if err := model.Validate(); err != nil {
		return core.Model{}, errors.WithStack(err)
	}
	return model, nil
}

func rejectTrailing(dec *json.Decoder) error {
	var extra json.RawMessage
	err := dec.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "core-pipe trailing JSON")
	}
	return errors.New("core-pipe document must be a single JSON value")
}

func parseKey(field, raw string) (identity.Key, error) {
	if strings.TrimSpace(raw) == "" {
		return identity.Key{}, errors.Errorf("%s is required", field)
	}
	key, err := identity.ParseKey(raw)
	if err != nil {
		return identity.Key{}, errors.Wrapf(err, "%s", field)
	}
	return key, nil
}

func keyString(key identity.Key) string {
	if key == (identity.Key{}) {
		return ""
	}
	return key.String()
}

func optionalKeyString(key *identity.Key) string {
	if key == nil {
		return ""
	}
	return keyString(*key)
}

func optionalKeyToCore(field, raw string) (*identity.Key, error) {
	if raw == "" {
		return nil, nil //nolint:nilnil // blank key is absence, not an error
	}
	key, err := parseKey(field, raw)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func keyStrings(keys []identity.Key) []string {
	if len(keys) == 0 {
		return nil
	}
	out := make([]string, len(keys))
	for i, key := range keys {
		out[i] = keyString(key)
	}
	return out
}

func keysToCore(field string, raw []string) ([]identity.Key, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]identity.Key, len(raw))
	for i, item := range raw {
		key, err := parseKey(fmt.Sprintf("%s[%d]", field, i), item)
		if err != nil {
			return nil, err
		}
		out[i] = key
	}
	return out, nil
}
