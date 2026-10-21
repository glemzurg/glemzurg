package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

type pipeClassAssociation struct {
	Key                 string           `json:"key"`
	Name                string           `json:"name"`
	Details             string           `json:"details,omitempty"`
	FromClassKey        string           `json:"from_class_key"`
	FromMultiplicity    string           `json:"from_multiplicity"`
	ToClassKey          string           `json:"to_class_key"`
	ToMultiplicity      string           `json:"to_multiplicity"`
	AssociationClassKey string           `json:"association_class_key,omitempty"`
	Uniqueness          []pipeUniqueness `json:"uniqueness,omitempty"`
	UmlComment          string           `json:"uml_comment,omitempty"`
	Invariants          []pipeLogic      `json:"invariants,omitempty"`
}

type pipeUniqueness struct {
	FromAttributes []string `json:"from_attributes,omitempty"`
	ToAttributes   []string `json:"to_attributes,omitempty"`
}

func classAssociationFromCore(assoc model_class.Association) (pipeClassAssociation, error) {
	fromMult, err := multiplicityFromCore(assoc.FromMultiplicity)
	if err != nil {
		return pipeClassAssociation{}, errors.Wrap(err, "from_multiplicity")
	}
	toMult, err := multiplicityFromCore(assoc.ToMultiplicity)
	if err != nil {
		return pipeClassAssociation{}, errors.Wrap(err, "to_multiplicity")
	}
	item := pipeClassAssociation{
		Key:                 keyString(assoc.Key),
		Name:                assoc.Name,
		Details:             assoc.Details,
		FromClassKey:        keyString(assoc.FromClassKey),
		FromMultiplicity:    fromMult,
		ToClassKey:          keyString(assoc.ToClassKey),
		ToMultiplicity:      toMult,
		AssociationClassKey: optionalKeyString(assoc.AssociationClassKey),
		UmlComment:          assoc.UmlComment,
	}
	for _, uniqueness := range assoc.Uniqueness {
		item.Uniqueness = append(item.Uniqueness, pipeUniqueness{
			FromAttributes: keyStrings(uniqueness.FromAttributeKeys),
			ToAttributes:   keyStrings(uniqueness.ToAttributeKeys),
		})
	}
	for _, inv := range assoc.Invariants {
		item.Invariants = append(item.Invariants, logicFromCore(inv))
	}
	return item, nil
}

func classAssociationsFromCore(assocs map[identity.Key]model_class.Association) ([]pipeClassAssociation, error) {
	if len(assocs) == 0 {
		return nil, nil
	}
	out := make([]pipeClassAssociation, 0, len(assocs))
	for _, assoc := range assocs {
		item, err := classAssociationFromCore(assoc)
		if err != nil {
			return nil, errors.Wrapf(err, "class association %s", assoc.Key.String())
		}
		out = append(out, item)
	}
	sortByKey(out, func(item pipeClassAssociation) string { return item.Key })
	return out, nil
}

func classAssociationsToCore(items []pipeClassAssociation) (map[identity.Key]model_class.Association, error) {
	if len(items) == 0 {
		return nil, nil //nolint:nilnil // omitted collection is the zero value, not an error
	}
	out := make(map[identity.Key]model_class.Association, len(items))
	for i, item := range items {
		assoc, err := classAssociationToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "[%d]", i)
		}
		if err := putUnique(out, assoc.Key, assoc, i); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func classAssociationToCore(item pipeClassAssociation) (model_class.Association, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_class.Association{}, err
	}
	fromClass, err := parseKey("from_class_key", item.FromClassKey)
	if err != nil {
		return model_class.Association{}, err
	}
	toClass, err := parseKey("to_class_key", item.ToClassKey)
	if err != nil {
		return model_class.Association{}, err
	}
	fromMult, err := multiplicityToCore("from_multiplicity", item.FromMultiplicity)
	if err != nil {
		return model_class.Association{}, err
	}
	toMult, err := multiplicityToCore("to_multiplicity", item.ToMultiplicity)
	if err != nil {
		return model_class.Association{}, err
	}
	assocClass, err := optionalKeyToCore("association_class_key", item.AssociationClassKey)
	if err != nil {
		return model_class.Association{}, err
	}
	uniqueness, err := uniquenessToCore(item.Uniqueness)
	if err != nil {
		return model_class.Association{}, err
	}
	invariants, err := logicsToCore(item.Invariants)
	if err != nil {
		return model_class.Association{}, errors.Wrap(err, "invariants")
	}

	assoc := model_class.NewAssociation(key, model_class.AssociationDetails{
		Name:    item.Name,
		Details: item.Details,
	}, model_class.AssociationEnd{
		ClassKey:     fromClass,
		Multiplicity: fromMult,
	}, model_class.AssociationEnd{
		ClassKey:     toClass,
		Multiplicity: toMult,
	}, model_class.AssociationOptions{
		AssociationClassKey: assocClass,
		Uniqueness:          uniqueness,
		UmlComment:          item.UmlComment,
	})
	assoc.SetInvariants(invariants)
	return assoc, nil
}

func uniquenessToCore(items []pipeUniqueness) ([]model_class.AssociationUniqueness, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_class.AssociationUniqueness, 0, len(items))
	for i, item := range items {
		fromKeys, err := keysToCore("from_attributes", item.FromAttributes)
		if err != nil {
			return nil, errors.Wrapf(err, "uniqueness[%d]", i)
		}
		toKeys, err := keysToCore("to_attributes", item.ToAttributes)
		if err != nil {
			return nil, errors.Wrapf(err, "uniqueness[%d]", i)
		}
		out = append(out, model_class.NewAssociationUniqueness(fromKeys, toKeys))
	}
	return out, nil
}
