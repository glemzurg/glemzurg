package database

import (
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanScenario(scanner Scanner, useCaseKeyPtr *string, scenario *model_scenario.Scenario) (err error) {
	var stepsJSON []byte
	if err = scanner.Scan(
		&scenario.Key,
		&scenario.Name,
		useCaseKeyPtr,
		&scenario.Details,
		&stepsJSON,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Unmarshal the steps JSON if present
	if len(stepsJSON) > 0 {
		if err = scenario.Steps.FromJSON(string(stepsJSON)); err != nil {
			return err
		}
	}

	return nil
}

// LoadScenario loads a scenario from the database
func LoadScenario(dbOrTx DbOrTx, modelKey, scenarioKey string) (useCaseKey string, scenario model_scenario.Scenario, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return "", model_scenario.Scenario{}, err
	}
	scenarioKey, err = identity.PreenKey(scenarioKey)
	if err != nil {
		return "", model_scenario.Scenario{}, err
	}

	// Query the database.
	err = dbQueryRow(
		dbOrTx,
		func(scanner Scanner) (err error) {
			if err = scanScenario(scanner, &useCaseKey, &scenario); err != nil {
				return err
			}
			return nil
		},
		`SELECT
			scenario_key,
			name,
			use_case_key,
			details,
			steps
		FROM
			scenario
		WHERE
			scenario_key = $2
		AND
			model_key = $1`,
		modelKey,
		scenarioKey)
	if err != nil {
		return "", model_scenario.Scenario{}, errors.WithStack(err)
	}

	return useCaseKey, scenario, nil
}

// AddScenario adds a scenario to the database.
func AddScenario(dbOrTx DbOrTx, modelKey, useCaseKey string, scenario model_scenario.Scenario) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	useCaseKey, err = identity.PreenKey(useCaseKey)
	if err != nil {
		return err
	}
	scenario.Key, err = identity.PreenKey(scenario.Key)
	if err != nil {
		return err
	}

	// Serialize the steps to JSON
	stepsJSON, err := scenario.Steps.ToJSON()
	if err != nil {
		return err
	}

	// Insert the record.
	_, err = dbExec(
		dbOrTx,
		`INSERT INTO scenario (
			model_key,
			scenario_key,
			name,
			use_case_key,
			details,
			steps
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)`,
		modelKey,
		scenario.Key,
		scenario.Name,
		useCaseKey,
		scenario.Details,
		stepsJSON)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// UpdateScenario updates a scenario in the database.
func UpdateScenario(dbOrTx DbOrTx, modelKey string, scenario model_scenario.Scenario) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	scenarioKey, err := identity.PreenKey(scenario.Key)
	if err != nil {
		return err
	}

	// Serialize the steps to JSON
	stepsJSON, err := scenario.Steps.ToJSON()
	if err != nil {
		return err
	}

	// Update the data.
	_, err = dbExec(dbOrTx, `
		UPDATE
			scenario
		SET
			name         = $3 ,
			details      = $4 ,
			steps        = $5
		WHERE
			scenario_key = $2
		AND
			model_key = $1`,
		modelKey,
		scenarioKey,
		scenario.Name,
		scenario.Details,
		stepsJSON)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveScenario deletes a scenario from the database.
func RemoveScenario(dbOrTx DbOrTx, modelKey, scenarioKey string) (err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return err
	}
	scenarioKey, err = identity.PreenKey(scenarioKey)
	if err != nil {
		return err
	}

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			scenario
		WHERE
			scenario_key = $2
		AND
			model_key = $1`,
		modelKey,
		scenarioKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryScenarios queries all scenarios for a model.
func QueryScenarios(dbOrTx DbOrTx, modelKey string) (scenarios map[string][]model_scenario.Scenario, err error) {

	// Keys should be preened so they collide correctly.
	modelKey, err = identity.PreenKey(modelKey)
	if err != nil {
		return nil, err
	}

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var scenario model_scenario.Scenario
			var useCaseKey string
			if err = scanScenario(scanner, &useCaseKey, &scenario); err != nil {
				return err
			}
			if scenarios == nil {
				scenarios = map[string][]model_scenario.Scenario{}
			}
			scenarios[useCaseKey] = append(scenarios[useCaseKey], scenario)
			return nil
		},
		`SELECT
			scenario_key,
			name,
			use_case_key,
			details,
			steps
		FROM
			scenario
		WHERE
			model_key = $1
		ORDER BY
			use_case_key, scenario_key`,
		modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return scenarios, nil
}
