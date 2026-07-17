package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

// SimulationRequireRow is one require join for a parameter simulation rule.
type SimulationRequireRow struct {
	ActionKey    identity.Key
	ParameterKey string
	RuleIndex    int
	LogicKey     identity.Key
}

// SimulationSpecRow is one specification join for a parameter simulation rule.
type SimulationSpecRow struct {
	ActionKey    identity.Key
	ParameterKey string
	RuleIndex    int
	LogicKey     identity.Key
}

func scanActionParameterSimulationRequire(
	scanner Scanner,
	actionKeyPtr *identity.Key,
	parameterKeyPtr *string,
	ruleIndexPtr *int,
	logicKeyPtr *identity.Key,
) error {
	var actionKeyStr, parameterKeyStr, logicKeyStr string
	var ruleIndex int
	if err := scanner.Scan(&actionKeyStr, &parameterKeyStr, &ruleIndex, &logicKeyStr); err != nil {
		if err.Error() == _POSTGRES_NOT_FOUND {
			return ErrNotFound
		}
		return err
	}
	var err error
	*actionKeyPtr, err = identity.ParseKey(actionKeyStr)
	if err != nil {
		return err
	}
	*parameterKeyPtr = parameterKeyStr
	*ruleIndexPtr = ruleIndex
	*logicKeyPtr, err = identity.ParseKey(logicKeyStr)
	return err
}

func scanActionParameterSimulationSpec(
	scanner Scanner,
	actionKeyPtr *identity.Key,
	parameterKeyPtr *string,
	ruleIndexPtr *int,
	logicKeyPtr *identity.Key,
) error {
	return scanActionParameterSimulationRequire(scanner, actionKeyPtr, parameterKeyPtr, ruleIndexPtr, logicKeyPtr)
}

// QueryActionParameterSimulationRequires loads require rows ordered by rule then logic sort.
func QueryActionParameterSimulationRequires(dbOrTx DbOrTx, modelKey string) ([]SimulationRequireRow, error) {
	var rows []SimulationRequireRow
	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var actionKey identity.Key
		var parameterSubKey string
		var ruleIndex int
		var logicKey identity.Key
		if err := scanActionParameterSimulationRequire(scanner, &actionKey, &parameterSubKey, &ruleIndex, &logicKey); err != nil {
			return errors.WithStack(err)
		}
		rows = append(rows, SimulationRequireRow{
			ActionKey:    actionKey,
			ParameterKey: parameterSubKey,
			RuleIndex:    ruleIndex,
			LogicKey:     logicKey,
		})
		return nil
	}, `SELECT
			apsr.action_key,
			apsr.parameter_key,
			apsr.rule_index,
			apsr.logic_key
		FROM
			action_parameter_simulation_require apsr
		JOIN
			logic l ON l.model_key = apsr.model_key AND l.logic_key = apsr.logic_key
		WHERE
			apsr.model_key = $1
		ORDER BY apsr.action_key, apsr.parameter_key, apsr.rule_index, l.sort_order, apsr.logic_key`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rows, nil
}

// QueryActionParameterSimulationSpecs loads specification rows ordered by rule index.
func QueryActionParameterSimulationSpecs(dbOrTx DbOrTx, modelKey string) ([]SimulationSpecRow, error) {
	var rows []SimulationSpecRow
	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var actionKey identity.Key
		var parameterSubKey string
		var ruleIndex int
		var logicKey identity.Key
		if err := scanActionParameterSimulationSpec(scanner, &actionKey, &parameterSubKey, &ruleIndex, &logicKey); err != nil {
			return errors.WithStack(err)
		}
		rows = append(rows, SimulationSpecRow{
			ActionKey:    actionKey,
			ParameterKey: parameterSubKey,
			RuleIndex:    ruleIndex,
			LogicKey:     logicKey,
		})
		return nil
	}, `SELECT
			apss.action_key,
			apss.parameter_key,
			apss.rule_index,
			apss.logic_key
		FROM
			action_parameter_simulation_spec apss
		WHERE
			apss.model_key = $1
		ORDER BY apss.action_key, apss.parameter_key, apss.rule_index`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rows, nil
}

// AddActionParameterSimulationRequires inserts simulation require join rows.
func AddActionParameterSimulationRequires(dbOrTx DbOrTx, modelKey string, rows []SimulationRequireRow) error {
	if len(rows) == 0 {
		return nil
	}
	var qb strings.Builder
	qb.WriteString(`INSERT INTO action_parameter_simulation_require (model_key, action_key, parameter_key, rule_index, logic_key) VALUES `)
	args := make([]any, 0, len(rows)*5)
	for i, row := range rows {
		if i > 0 {
			qb.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args, modelKey, row.ActionKey.String(), row.ParameterKey, row.RuleIndex, row.LogicKey.String())
	}
	return errors.WithStack(dbExec(dbOrTx, qb.String(), args...))
}

// AddActionParameterSimulationSpecs inserts simulation specification join rows.
func AddActionParameterSimulationSpecs(dbOrTx DbOrTx, modelKey string, rows []SimulationSpecRow) error {
	if len(rows) == 0 {
		return nil
	}
	var qb strings.Builder
	qb.WriteString(`INSERT INTO action_parameter_simulation_spec (model_key, action_key, parameter_key, rule_index, logic_key) VALUES `)
	args := make([]any, 0, len(rows)*5)
	for i, row := range rows {
		if i > 0 {
			qb.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5)
		args = append(args, modelKey, row.ActionKey.String(), row.ParameterKey, row.RuleIndex, row.LogicKey.String())
	}
	return errors.WithStack(dbExec(dbOrTx, qb.String(), args...))
}
