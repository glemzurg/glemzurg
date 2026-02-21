package database

import (
	"fmt"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/req_model/model_scenario"

	"github.com/pkg/errors"
)

// Populate a golang struct from a database row.
func scanScenario(scanner Scanner, useCaseKeyPtr *identity.Key, scenario *model_scenario.Scenario) (err error) {
	var scenarioKeyStr string
	var useCaseKeyStr string
	var stepsJSON []byte
	if err = scanner.Scan(
		&scenarioKeyStr,
		&scenario.Name,
		&useCaseKeyStr,
		&scenario.Details,
		&stepsJSON,
	); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			err = ErrNotFound
		}
		return err // Do not wrap in stack here. It will be wrapped in the database calls.
	}

	// Parse the scenario key string into an identity.Key.
	scenario.Key, err = identity.ParseKey(scenarioKeyStr)
	if err != nil {
		return err
	}

	// Parse the use case key string into an identity.Key.
	*useCaseKeyPtr, err = identity.ParseKey(useCaseKeyStr)
	if err != nil {
		return err
	}

	// Unmarshal the steps JSON if present
	if len(stepsJSON) > 0 {
		scenario.Steps = &model_scenario.Step{}
		if err = scenario.Steps.FromJSON(string(stepsJSON)); err != nil {
			return err
		}
	}

	return nil
}

// LoadScenario loads a scenario from the database
func LoadScenario(dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key) (useCaseKey identity.Key, scenario model_scenario.Scenario, err error) {

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
		scenarioKey.String())
	if err != nil {
		return identity.Key{}, model_scenario.Scenario{}, errors.WithStack(err)
	}

	return useCaseKey, scenario, nil
}

// AddScenario adds a scenario to the database.
func AddScenario(dbOrTx DbOrTx, modelKey string, useCaseKey identity.Key, scenario model_scenario.Scenario) (err error) {
	return AddScenarios(dbOrTx, modelKey, map[identity.Key][]model_scenario.Scenario{
		useCaseKey: {scenario},
	})
}

// UpdateScenario updates a scenario in the database.
func UpdateScenario(dbOrTx DbOrTx, modelKey string, scenario model_scenario.Scenario) (err error) {

	// Serialize the steps to JSON
	var stepsJSON interface{}
	if scenario.Steps != nil {
		stepsJSON, err = scenario.Steps.ToJSON()
		if err != nil {
			return err
		}
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
		scenario.Key.String(),
		scenario.Name,
		scenario.Details,
		stepsJSON)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// RemoveScenario deletes a scenario from the database.
func RemoveScenario(dbOrTx DbOrTx, modelKey string, scenarioKey identity.Key) (err error) {

	// Delete the data.
	_, err = dbExec(dbOrTx, `
		DELETE FROM
			scenario
		WHERE
			scenario_key = $2
		AND
			model_key = $1`,
		modelKey,
		scenarioKey.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// QueryScenarios queries all scenarios for a model.
func QueryScenarios(dbOrTx DbOrTx, modelKey string) (scenarios map[identity.Key][]model_scenario.Scenario, err error) {

	// Query the database.
	err = dbQuery(
		dbOrTx,
		func(scanner Scanner) (err error) {
			var scenario model_scenario.Scenario
			var useCaseKey identity.Key
			if err = scanScenario(scanner, &useCaseKey, &scenario); err != nil {
				return err
			}
			if scenarios == nil {
				scenarios = map[identity.Key][]model_scenario.Scenario{}
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

// AddScenarios adds multiple scenarios to the database in a single insert.
func AddScenarios(dbOrTx DbOrTx, modelKey string, scenarios map[identity.Key][]model_scenario.Scenario) (err error) {
	// Count total scenarios.
	count := 0
	for _, scens := range scenarios {
		count += len(scens)
	}
	if count == 0 {
		return nil
	}

	// Build the bulk insert query.
	query := `INSERT INTO scenario (model_key, scenario_key, name, use_case_key, details, steps) VALUES `
	args := make([]interface{}, 0, count*6)
	i := 0
	for useCaseKey, scenList := range scenarios {
		for _, scenario := range scenList {
			if i > 0 {
				query += ", "
			}
			base := i * 6
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6)

			// Serialize the steps to JSON.
			var stepsJSON interface{}
			if scenario.Steps != nil {
				stepsJSON, err = scenario.Steps.ToJSON()
				if err != nil {
					return err
				}
			}

			args = append(args, modelKey, scenario.Key.String(), scenario.Name, useCaseKey.String(), scenario.Details, stepsJSON)
			i++
		}
	}

	_, err = dbExec(dbOrTx, query, args...)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
