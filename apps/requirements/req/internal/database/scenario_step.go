package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

	"github.com/pkg/errors"
)

// stepRow is a flattened database row for a scenario step.
type stepRow struct {
	scenarioKey   identity.Key
	parentStepKey *identity.Key
	step          model_scenario.Step
}

// Populate a golang struct from a database row.
func scanStep(scanner Scanner, scenarioKeyPtr *identity.Key, parentStepKeyPtr **identity.Key, step *model_scenario.Step) (err error) {
	var stepKeyStr string
	var scenarioKeyStr string
	var parentStepKeyStrPtr *string
	var leafTypePtr *string
	var conditionPtr *string
	var descriptionPtr *string
	var fromObjectKeyPtr *string
	var toObjectKeyPtr *string
	var eventKeyPtr *string
	var queryKeyPtr *string
	var scenarioRefKeyPtr *string

	if err = scanner.Scan(
		&stepKeyStr,
		&scenarioKeyStr,
		&parentStepKeyStrPtr,
		&step.SortOrder,
		&step.StepType,
		&leafTypePtr,
		&conditionPtr,
		&descriptionPtr,
		&fromObjectKeyPtr,
		&toObjectKeyPtr,
		&eventKeyPtr,
		&queryKeyPtr,
		&scenarioRefKeyPtr,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the step key string into an identity.Key.
	step.Key, err = identity.ParseKey(stepKeyStr)
	if err != nil {
		return err
	}

	// Parse the scenario key string into an identity.Key.
	*scenarioKeyPtr, err = identity.ParseKey(scenarioKeyStr)
	if err != nil {
		return err
	}

	// Parse optional parent step key.
	if parentStepKeyStrPtr != nil {
		parentKey, err := identity.ParseKey(*parentStepKeyStrPtr)
		if err != nil {
			return err
		}
		*parentStepKeyPtr = &parentKey
	}

	// Parse optional fields.
	if leafTypePtr != nil {
		step.LeafType = leafTypePtr
	}
	if conditionPtr != nil {
		step.Condition = *conditionPtr
	}
	if descriptionPtr != nil {
		step.Description = *descriptionPtr
	}
	if fromObjectKeyPtr != nil {
		fromKey, err := identity.ParseKey(*fromObjectKeyPtr)
		if err != nil {
			return err
		}
		step.FromObjectKey = &fromKey
	}
	if toObjectKeyPtr != nil {
		toKey, err := identity.ParseKey(*toObjectKeyPtr)
		if err != nil {
			return err
		}
		step.ToObjectKey = &toKey
	}
	if eventKeyPtr != nil {
		eventKey, err := identity.ParseKey(*eventKeyPtr)
		if err != nil {
			return err
		}
		step.EventKey = &eventKey
	}
	if queryKeyPtr != nil {
		queryKey, err := identity.ParseKey(*queryKeyPtr)
		if err != nil {
			return err
		}
		step.QueryKey = &queryKey
	}
	if scenarioRefKeyPtr != nil {
		scenarioRefKey, err := identity.ParseKey(*scenarioRefKeyPtr)
		if err != nil {
			return err
		}
		step.ScenarioKey = &scenarioRefKey
	}

	return nil
}

// LoadStep loads a single step from the database.
func LoadStep(dbOrTx DbOrTx, modelKey string, stepKey identity.Key) (scenarioKey identity.Key, parentStepKey *identity.Key, step model_scenario.Step, err error) {

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanStep(scanner, &scenarioKey, &parentStepKey, &step); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			scenario_step_key,
			scenario_key,
			parent_step_key,
			sort_order,
			step_type,
			leaf_type,
			condition,
			description,
			from_object_key,
			to_object_key,
			event_key,
			query_key,
			scenario_ref_key
		FROM
			scenario_step
		WHERE
			scenario_step_key = $2
		AND
			model_key = $1`,
		modelKey,
		stepKey.String())
	if err != nil {
		return identity.Key{}, nil, model_scenario.Step{}, errors.WithStack(err)
	}

	return scenarioKey, parentStepKey, step, nil
}

// AddStep adds a single step row to the database.
func AddStep(dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key, parentStepKey *identity.Key, step model_scenario.Step) (err error) {
	return AddSteps(dbOrTx, modelKey, []stepRow{
		{scenarioKey: scenarioKey, parentStepKey: parentStepKey, step: step},
	})
}

// UpdateStep updates a step in the database.
func UpdateStep(dbOrTx DbOrTx, modelKey string, step model_scenario.Step) (err error) {

	// Handle optional key pointers.
	var leafTypePtr *string
	if step.LeafType != nil {
		leafTypePtr = step.LeafType
	}
	var conditionPtr *string
	if step.Condition != "" {
		conditionPtr = &step.Condition
	}
	var descriptionPtr *string
	if step.Description != "" {
		descriptionPtr = &step.Description
	}
	var fromObjectKeyPtr *string
	if step.FromObjectKey != nil {
		s := step.FromObjectKey.String()
		fromObjectKeyPtr = &s
	}
	var toObjectKeyPtr *string
	if step.ToObjectKey != nil {
		s := step.ToObjectKey.String()
		toObjectKeyPtr = &s
	}
	var eventKeyPtr *string
	if step.EventKey != nil {
		s := step.EventKey.String()
		eventKeyPtr = &s
	}
	var queryKeyPtr *string
	if step.QueryKey != nil {
		s := step.QueryKey.String()
		queryKeyPtr = &s
	}
	var scenarioRefKeyPtr *string
	if step.ScenarioKey != nil {
		s := step.ScenarioKey.String()
		scenarioRefKeyPtr = &s
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			scenario_step
		SET
			sort_order       = $3 ,
			step_type        = $4 ,
			leaf_type        = $5 ,
			condition        = $6 ,
			description      = $7 ,
			from_object_key  = $8 ,
			to_object_key    = $9 ,
			event_key        = $10,
			query_key        = $11,
			scenario_ref_key = $12
		WHERE
			model_key = $1
		AND
			scenario_step_key = $2`,
		modelKey,
		step.Key.String(),
		step.SortOrder,
		step.StepType,
		leafTypePtr,
		conditionPtr,
		descriptionPtr,
		fromObjectKeyPtr,
		toObjectKeyPtr,
		eventKeyPtr,
		queryKeyPtr,
		scenarioRefKeyPtr)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveStep deletes a step from the database. CASCADE will delete children.
func RemoveStep(dbOrTx DbOrTx, modelKey string, stepKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			scenario_step
		WHERE
			model_key = $1
		AND
			scenario_step_key = $2`,
		modelKey,
		stepKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QuerySteps loads all steps for a model and reconstructs them into trees keyed by scenario key.
func QuerySteps(dbOrTx DbOrTx, modelKey string) (steps map[identity.Key]*model_scenario.Step, err error) {

	// Collect all flat rows.
	var rows []stepRow

	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var row stepRow
			if err = scanStep(scanner, &row.scenarioKey, &row.parentStepKey, &row.step); err != nil {
				return errors.WithStack(err)
			}
			rows = append(rows, row)
			return nil
		},
		`SELECT
			scenario_step_key,
			scenario_key,
			parent_step_key,
			sort_order,
			step_type,
			leaf_type,
			condition,
			description,
			from_object_key,
			to_object_key,
			event_key,
			query_key,
			scenario_ref_key
		FROM
			scenario_step
		WHERE
			model_key = $1
		ORDER BY scenario_key, sort_order`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	// Build the trees from flat rows.
	steps = rebuildStepTrees(rows)

	return steps, nil
}

// rebuildStepTrees reconstructs step trees from flat rows.
func rebuildStepTrees(rows []stepRow) map[identity.Key]*model_scenario.Step {
	// Index steps by key for lookups.
	stepByKey := make(map[identity.Key]*model_scenario.Step, len(rows))
	// Track which scenario each step belongs to.
	scenarioByStep := make(map[identity.Key]identity.Key, len(rows))
	// Track parent relationships.
	parentByStep := make(map[identity.Key]*identity.Key, len(rows))
	// Track children by parent key.
	childrenByParent := make(map[identity.Key][]identity.Key)
	// Track roots (no parent).
	var rootKeys []identity.Key

	for i := range rows {
		row := &rows[i]
		key := row.step.Key
		stepByKey[key] = &row.step
		scenarioByStep[key] = row.scenarioKey
		parentByStep[key] = row.parentStepKey

		if row.parentStepKey == nil {
			rootKeys = append(rootKeys, key)
		} else {
			childrenByParent[*row.parentStepKey] = append(childrenByParent[*row.parentStepKey], key)
		}
	}

	// Recursively build tree from a step key.
	var buildTree func(key identity.Key) model_scenario.Step
	buildTree = func(key identity.Key) model_scenario.Step {
		step := *stepByKey[key]
		if children, ok := childrenByParent[key]; ok {
			step.Statements = make([]model_scenario.Step, len(children))
			for i, childKey := range children {
				step.Statements[i] = buildTree(childKey)
			}
		}
		return step
	}

	// Build result map keyed by scenario key.
	result := make(map[identity.Key]*model_scenario.Step)
	for _, rootKey := range rootKeys {
		tree := buildTree(rootKey)
		scenKey := scenarioByStep[rootKey]
		result[scenKey] = &tree
	}

	return result
}

// flattenSteps walks a step tree and produces flat stepRow slices in topological order (root first).
func flattenSteps(scenarioKey identity.Key, root *model_scenario.Step) []stepRow {
	var rows []stepRow
	flattenStepsRecursive(scenarioKey, nil, root, &rows)
	return rows
}

// flattenStepsRecursive recursively flattens a step tree.
func flattenStepsRecursive(scenarioKey identity.Key, parentKey *identity.Key, step *model_scenario.Step, rows *[]stepRow) {
	// Add this step (without Statements, which are children in the DB).
	flatStep := model_scenario.Step{
		Key:           step.Key,
		SortOrder:     step.SortOrder,
		StepType:      step.StepType,
		LeafType:      step.LeafType,
		Condition:     step.Condition,
		Description:   step.Description,
		FromObjectKey: step.FromObjectKey,
		ToObjectKey:   step.ToObjectKey,
		EventKey:      step.EventKey,
		QueryKey:      step.QueryKey,
		ScenarioKey:   step.ScenarioKey,
		// Note: Statements not copied — they're separate rows in the DB.
	}
	*rows = append(*rows, stepRow{
		scenarioKey:   scenarioKey,
		parentStepKey: parentKey,
		step:          flatStep,
	})

	// Recurse into children.
	for i := range step.Statements {
		childParent := step.Key
		flattenStepsRecursive(scenarioKey, &childParent, &step.Statements[i], rows)
	}
}

// AddSteps adds multiple step rows to the database. Rows must be in topological order
// (parent before children) due to non-deferrable self-referential FK.
func AddSteps(dbOrTx DbOrTx, modelKey string, rows []stepRow) (err error) {
	if len(rows) == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO scenario_step (model_key, scenario_step_key, scenario_key, parent_step_key, sort_order, step_type, leaf_type, condition, description, from_object_key, to_object_key, event_key, query_key, scenario_ref_key) VALUES `
	args := make([]interface{}, 0, len(rows)*14)

	for i, row := range rows {
		if i > 0 {
			query += ", "
		}
		base := i * 14
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7,
			base+8, base+9, base+10, base+11, base+12, base+13, base+14)

		// Handle optional key pointers.
		var parentStepKeyPtr *string
		if row.parentStepKey != nil {
			s := row.parentStepKey.String()
			parentStepKeyPtr = &s
		}
		var leafTypePtr *string
		if row.step.LeafType != nil {
			leafTypePtr = row.step.LeafType
		}
		var conditionPtr *string
		if row.step.Condition != "" {
			conditionPtr = &row.step.Condition
		}
		var descriptionPtr *string
		if row.step.Description != "" {
			descriptionPtr = &row.step.Description
		}
		var fromObjectKeyPtr *string
		if row.step.FromObjectKey != nil {
			s := row.step.FromObjectKey.String()
			fromObjectKeyPtr = &s
		}
		var toObjectKeyPtr *string
		if row.step.ToObjectKey != nil {
			s := row.step.ToObjectKey.String()
			toObjectKeyPtr = &s
		}
		var eventKeyPtr *string
		if row.step.EventKey != nil {
			s := row.step.EventKey.String()
			eventKeyPtr = &s
		}
		var queryKeyPtr *string
		if row.step.QueryKey != nil {
			s := row.step.QueryKey.String()
			queryKeyPtr = &s
		}
		var scenarioRefKeyPtr *string
		if row.step.ScenarioKey != nil {
			s := row.step.ScenarioKey.String()
			scenarioRefKeyPtr = &s
		}

		args = append(args,
			modelKey,
			row.step.Key.String(),
			row.scenarioKey.String(),
			parentStepKeyPtr,
			row.step.SortOrder,
			row.step.StepType,
			leafTypePtr,
			conditionPtr,
			descriptionPtr,
			fromObjectKeyPtr,
			toObjectKeyPtr,
			eventKeyPtr,
			queryKeyPtr,
			scenarioRefKeyPtr,
		)
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
