package generate

import (
	"sort"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_class"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_scenario"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/generate/req_flat"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/pkg/errors"
)

// stepContext bundles the lookup maps needed when processing scenario steps.
type stepContext struct {
	eventLookup    map[string]model_state.Event
	scenarioLookup map[string]model_scenario.Scenario
	objectLookup   map[string]model_scenario.Object
	classLookup    map[string]model_class.Class
}

// mermaidSequence accumulates indented Mermaid sequenceDiagram body lines.
type mermaidSequence struct {
	lines  []string
	indent int
}

func generateScenarioMermaidContents(reqs *req_flat.Requirements, scenario model_scenario.Scenario) (contents string, err error) {
	ctx := newStepContext(reqs)
	builder := &mermaidSequence{}

	participantIDs, err := writeParticipants(ctx, builder, scenario)
	if err != nil {
		return "", err
	}
	if len(participantIDs) == 0 {
		builder.writeLine("Note over no_actors: No actors defined")
		return strings.Join(builder.lines, "\n"), nil
	}

	if scenario.Steps == nil || len(scenario.Steps.Statements) == 0 {
		builder.writeMessage(participantIDs[0], participantIDs[0], "No operations defined")
	} else if err = addMermaidSteps(ctx, builder, scenario.Steps.Statements); err != nil {
		return "", err
	}

	return strings.Join(builder.lines, "\n"), nil
}

func (m *mermaidSequence) writeLine(line string) {
	m.lines = append(m.lines, strings.Repeat("    ", m.indent)+line)
}

func (m *mermaidSequence) writeMessage(fromID, toID, text string) {
	m.writeLine(fromID + "->>" + toID + ": " + mermaidSequenceText(text))
}

func newStepContext(reqs *req_flat.Requirements) stepContext {
	classLookup, _ := reqs.ClassLookup()
	return stepContext{
		eventLookup:    reqs.EventLookup(),
		scenarioLookup: reqs.ScenarioLookup(),
		objectLookup:   reqs.ObjectLookup(),
		classLookup:    classLookup,
	}
}

func writeParticipants(ctx stepContext, builder *mermaidSequence, scenario model_scenario.Scenario) ([]string, error) {
	if len(scenario.Objects) == 0 {
		builder.writeLine("participant no_actors as No actors defined")
		return nil, nil
	}

	var participantIDs []string
	for _, obj := range scenarioObjectsInOrder(scenario) {
		object, found := ctx.objectLookup[obj.Key.String()]
		if !found {
			return nil, errors.Errorf("unknown object key: '%s'", obj.Key.String())
		}
		class, found := ctx.classLookup[object.ClassKey.String()]
		if !found {
			return nil, errors.Errorf("unknown class key: '%s'", object.ClassKey.String())
		}

		participantID := scenarioObjectParticipantID(object.Key)
		participantIDs = append(participantIDs, participantID)
		builder.writeLine("participant " + participantID + " as " + mermaidParticipantAlias(object.GetName(class)))
	}
	return participantIDs, nil
}

func scenarioObjectParticipantID(objectKey identity.Key) string {
	return mermaidNodeID("sobject", objectKey)
}

// scenarioObjectsInOrder returns scenario objects sorted by ObjectNumber (YAML objects array order).
func scenarioObjectsInOrder(scenario model_scenario.Scenario) []model_scenario.Object {
	objects := make([]model_scenario.Object, 0, len(scenario.Objects))
	for _, obj := range scenario.Objects {
		objects = append(objects, obj)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ObjectNumber < objects[j].ObjectNumber
	})
	return objects
}

func mermaidNodeID(prefix string, key identity.Key) string {
	keyStr := key.String()
	keyStr = strings.ReplaceAll(keyStr, "/", "_")
	keyStr = strings.ReplaceAll(keyStr, "-", "_")
	keyStr = strings.ReplaceAll(keyStr, ".", "_")
	return prefix + "_" + keyStr
}

// mermaidSequenceText escapes characters that break sequenceDiagram message syntax.
// Colons are safe in message text because only the first colon after the arrow is structural.
func mermaidSequenceText(text string) string {
	return mermaidEscapeSequenceLine(text)
}

// mermaidParticipantAlias renders object labels without quotes; colons become line breaks.
func mermaidParticipantAlias(text string) string {
	if colon := strings.Index(text, ":"); colon >= 0 {
		text = text[:colon] + "<br/>" + text[colon+1:]
	}
	return mermaidEscapeSequenceLine(text)
}

func mermaidCondition(condition string) string {
	return mermaidEscapeSequenceLine(condition)
}

func mermaidEscapeSequenceLine(text string) string {
	text = strings.ReplaceAll(text, "#", "#35;")
	text = strings.ReplaceAll(text, ";", "#59;")
	text = strings.ReplaceAll(text, "\n", "<br/>")
	return text
}

func addMermaidSteps(ctx stepContext, builder *mermaidSequence, statements []model_scenario.Step) error {
	for _, stmt := range statements {
		switch stmt.StepType {
		case model_scenario.STEP_TYPE_LEAF:
			if err := addMermaidLeafStep(ctx, builder, stmt); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_SEQUENCE:
			if err := addMermaidSteps(ctx, builder, stmt.Statements); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_SWITCH:
			if err := addMermaidSwitchStep(ctx, builder, stmt); err != nil {
				return err
			}

		case model_scenario.STEP_TYPE_LOOP:
			if err := addMermaidLoopStep(ctx, builder, stmt); err != nil {
				return err
			}

		default:
			return errors.Errorf("unsupported step type in scenario Mermaid generation: '%s'", stmt.StepType)
		}
	}
	return nil
}

func addMermaidLeafStep(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	if stmt.LeafType == nil {
		return errors.Errorf("leaf step missing leaf_type: '%+v'", stmt)
	}

	switch *stmt.LeafType {
	case model_scenario.LEAF_TYPE_EVENT:
		return addMermaidEventLeaf(ctx, builder, stmt)
	case model_scenario.LEAF_TYPE_QUERY:
		return addMermaidQueryLeaf(ctx, builder, stmt)
	case model_scenario.LEAF_TYPE_SCENARIO:
		return addMermaidScenarioLeaf(ctx, builder, stmt)
	case model_scenario.LEAF_TYPE_DELETE:
		return addMermaidDeleteLeaf(ctx, builder, stmt)
	default:
		return errors.Errorf("unsupported leaf type in scenario Mermaid generation: '%s'", *stmt.LeafType)
	}
}

func addMermaidSwitchStep(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	if len(stmt.Statements) == 0 {
		return nil
	}
	if len(stmt.Statements) == 1 {
		return writeMermaidBlock(ctx, builder, "opt", stmt.Statements[0].Condition, stmt.Statements[0].Statements)
	}

	builder.writeLine("alt " + mermaidCondition(stmt.Statements[0].Condition))
	builder.indent++
	if err := addMermaidSteps(ctx, builder, stmt.Statements[0].Statements); err != nil {
		return err
	}
	for i := 1; i < len(stmt.Statements); i++ {
		builder.indent--
		builder.writeLine("else " + mermaidCondition(stmt.Statements[i].Condition))
		builder.indent++
		if err := addMermaidSteps(ctx, builder, stmt.Statements[i].Statements); err != nil {
			return err
		}
	}
	builder.indent--
	builder.writeLine("end")
	return nil
}

func addMermaidLoopStep(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	return writeMermaidBlock(ctx, builder, "loop", stmt.Condition, stmt.Statements)
}

func writeMermaidBlock(ctx stepContext, builder *mermaidSequence, keyword, condition string, statements []model_scenario.Step) error {
	builder.writeLine(keyword + " " + mermaidCondition(condition))
	builder.indent++
	if err := addMermaidSteps(ctx, builder, statements); err != nil {
		return err
	}
	builder.indent--
	builder.writeLine("end")
	return nil
}

func resolveFromToParticipantIDs(ctx stepContext, stmt model_scenario.Step) (fromID, toID string, err error) {
	fromObject, found := ctx.objectLookup[stmt.FromObjectKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
	}
	toObject, found := ctx.objectLookup[stmt.ToObjectKey.String()]
	if !found {
		return "", "", errors.Errorf("unknown to object key: '%s'", stmt.ToObjectKey.String())
	}

	return scenarioObjectParticipantID(fromObject.Key), scenarioObjectParticipantID(toObject.Key), nil
}

func addMermaidEventLeaf(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	fromID, toID, err := resolveFromToParticipantIDs(ctx, stmt)
	if err != nil {
		return err
	}

	builder.writeMessage(fromID, toID, buildEventText(ctx, stmt))
	return nil
}

func buildEventText(ctx stepContext, stmt model_scenario.Step) string {
	var textBuilder strings.Builder
	textBuilder.WriteString(stmt.Description)

	if stmt.EventKey != nil {
		event, found := ctx.eventLookup[stmt.EventKey.String()]
		if found {
			textBuilder.WriteString(model_state.SystemEventDisplayName(event.Name))
			if len(event.ParameterNames) > 0 {
				textBuilder.WriteString("(")
				for i, name := range event.ParameterNames {
					if i > 0 {
						textBuilder.WriteString(", ")
					}
					textBuilder.WriteString(name)
				}
				textBuilder.WriteString(")")
			}
		}
	}

	return textBuilder.String()
}

func addMermaidQueryLeaf(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	fromID, toID, err := resolveFromToParticipantIDs(ctx, stmt)
	if err != nil {
		return err
	}

	builder.writeMessage(fromID, toID, stmt.Description)
	return nil
}

func addMermaidScenarioLeaf(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	fromID, toID, err := resolveFromToParticipantIDs(ctx, stmt)
	if err != nil {
		return err
	}

	calledScenario, found := ctx.scenarioLookup[stmt.ScenarioKey.String()]
	if !found {
		return errors.Errorf("unknown called scenario key: '%s'", stmt.ScenarioKey.String())
	}
	builder.writeMessage(fromID, toID, "Scenario: "+calledScenario.Name)
	return nil
}

func addMermaidDeleteLeaf(ctx stepContext, builder *mermaidSequence, stmt model_scenario.Step) error {
	fromObject, found := ctx.objectLookup[stmt.FromObjectKey.String()]
	if !found {
		return errors.Errorf("unknown from object key: '%s'", stmt.FromObjectKey.String())
	}

	participantID := scenarioObjectParticipantID(fromObject.Key)
	builder.writeMessage(participantID, participantID, "(delete)")
	return nil
}
