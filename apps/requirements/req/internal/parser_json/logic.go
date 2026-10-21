package parser_json

import (
	"github.com/pkg/errors"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_logic/logic_spec"
)

// pipeSpec is the authored notation and text of an expression or type spec.
// The parsed tree is a cache and is rebuilt by constructors when a caller asks.
type pipeSpec struct {
	Notation      string `json:"notation,omitempty"`
	Specification string `json:"specification,omitempty"`
}

type pipeLogic struct {
	Key                  string    `json:"key"`
	Type                 string    `json:"type"`
	Description          string    `json:"description,omitempty"`
	Target               string    `json:"target,omitempty"`
	OverAssociationKey   string    `json:"over_association_key,omitempty"`
	Spec                 *pipeSpec `json:"spec,omitempty"`
	DestroyEventSpec     *pipeSpec `json:"destroy_event_spec,omitempty"`
	EndpointSelectorSpec *pipeSpec `json:"endpoint_selector_spec,omitempty"`
	TargetTypeSpec       *pipeSpec `json:"target_type_spec,omitempty"`
}

func specFromCore(spec logic_spec.ExpressionSpec) *pipeSpec {
	if spec.Notation == "" && spec.Specification == "" {
		return nil
	}
	return &pipeSpec{Notation: spec.Notation, Specification: spec.Specification}
}

func typeSpecFromCore(spec *logic_spec.TypeSpec) *pipeSpec {
	if spec == nil || (spec.Notation == "" && spec.Specification == "") {
		return nil
	}
	return &pipeSpec{Notation: spec.Notation, Specification: spec.Specification}
}

func specToCore(spec *pipeSpec) (logic_spec.ExpressionSpec, error) {
	if spec == nil {
		return logic_spec.ExpressionSpec{}, nil
	}
	parsed, err := logic_spec.NewExpressionSpec(spec.Notation, spec.Specification, nil)
	if err != nil {
		return logic_spec.ExpressionSpec{}, errors.WithStack(err)
	}
	return parsed, nil
}

func typeSpecToCore(spec *pipeSpec) (*logic_spec.TypeSpec, error) {
	if spec == nil {
		return nil, nil //nolint:nilnil // nil pointer means no type spec present
	}
	parsed, err := logic_spec.NewTypeSpec(spec.Notation, spec.Specification, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &parsed, nil
}

func logicFromCore(logic model_logic.Logic) pipeLogic {
	over := ""
	if logic.OverAssociationKey != nil {
		over = keyString(*logic.OverAssociationKey)
	}
	return pipeLogic{
		Key:                  keyString(logic.Key),
		Type:                 logic.Type,
		Description:          logic.Description,
		Target:               logic.Target,
		OverAssociationKey:   over,
		Spec:                 specFromCore(logic.Spec),
		DestroyEventSpec:     specFromCore(logic.DestroyEventSpec),
		EndpointSelectorSpec: specFromCore(logic.EndpointSelectorSpec),
		TargetTypeSpec:       typeSpecFromCore(logic.TargetTypeSpec),
	}
}

func logicToCore(item pipeLogic) (model_logic.Logic, error) {
	key, err := parseKey("key", item.Key)
	if err != nil {
		return model_logic.Logic{}, err
	}
	spec, err := specToCore(item.Spec)
	if err != nil {
		return model_logic.Logic{}, errors.Wrap(err, "spec")
	}
	targetType, err := typeSpecToCore(item.TargetTypeSpec)
	if err != nil {
		return model_logic.Logic{}, errors.Wrap(err, "target_type_spec")
	}

	logic := model_logic.NewLogic(key, item.Type, item.Description, item.Target, spec, targetType)

	if item.OverAssociationKey != "" {
		over, err := parseKey("over_association_key", item.OverAssociationKey)
		if err != nil {
			return model_logic.Logic{}, err
		}
		logic.SetOverAssociationKey(&over)
	}
	if item.DestroyEventSpec != nil {
		destroySpec, err := specToCore(item.DestroyEventSpec)
		if err != nil {
			return model_logic.Logic{}, errors.Wrap(err, "destroy_event_spec")
		}
		logic.SetDestroyEventSpec(destroySpec)
	}
	if item.EndpointSelectorSpec != nil {
		selector, err := specToCore(item.EndpointSelectorSpec)
		if err != nil {
			return model_logic.Logic{}, errors.Wrap(err, "endpoint_selector_spec")
		}
		logic.SetEndpointSelectorSpec(selector)
	}
	return logic, nil
}

func logicsFromCore(items []model_logic.Logic) []pipeLogic {
	if len(items) == 0 {
		return nil
	}
	out := make([]pipeLogic, len(items))
	for i, item := range items {
		out[i] = logicFromCore(item)
	}
	return out
}

func logicsToCore(items []pipeLogic) ([]model_logic.Logic, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]model_logic.Logic, 0, len(items))
	for i, item := range items {
		logic, err := logicToCore(item)
		if err != nil {
			return nil, errors.Wrapf(err, "logic[%d]", i)
		}
		out = append(out, logic)
	}
	return out, nil
}
