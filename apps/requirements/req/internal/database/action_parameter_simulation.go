package database

import (
	"fmt"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"

	"github.com/pkg/errors"
)

func scanActionParameterSimulationRequire(scanner Scanner, actionKeyPtr *identity.Key, parameterKeyPtr *string, logicKeyPtr *identity.Key) error {
	var actionKeyStr, parameterKeyStr, logicKeyStr string
	if err := scanner.Scan(&actionKeyStr, &parameterKeyStr, &logicKeyStr); err != nil {
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
	*logicKeyPtr, err = identity.ParseKey(logicKeyStr)
	return err
}

func scanActionParameterSimulationSpec(scanner Scanner, actionKeyPtr *identity.Key, parameterKeyPtr *string, logicKeyPtr *identity.Key) error {
	return scanActionParameterSimulationRequire(scanner, actionKeyPtr, parameterKeyPtr, logicKeyPtr)
}

// QueryActionParameterSimulationRequires loads simulation require logic keys grouped by parameter key.
func QueryActionParameterSimulationRequires(dbOrTx DbOrTx, modelKey string) (map[identity.Key][]identity.Key, error) {
	result := make(map[identity.Key][]identity.Key)
	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var actionKey identity.Key
		var parameterSubKey string
		var logicKey identity.Key
		if err := scanActionParameterSimulationRequire(scanner, &actionKey, &parameterSubKey, &logicKey); err != nil {
			return errors.WithStack(err)
		}
		paramKey, err := identity.NewParameterKey(actionKey, parameterSubKey)
		if err != nil {
			return errors.WithStack(err)
		}
		result[paramKey] = append(result[paramKey], logicKey)
		return nil
	}, `SELECT
			apsr.action_key,
			apsr.parameter_key,
			apsr.logic_key
		FROM
			action_parameter_simulation_require apsr
		JOIN
			logic l ON l.model_key = apsr.model_key AND l.logic_key = apsr.logic_key
		WHERE
			apsr.model_key = $1
		ORDER BY apsr.action_key, apsr.parameter_key, l.sort_order, apsr.logic_key`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// QueryActionParameterSimulationSpecs loads simulation specification logic keys by parameter key.
func QueryActionParameterSimulationSpecs(dbOrTx DbOrTx, modelKey string) (map[identity.Key]identity.Key, error) {
	result := make(map[identity.Key]identity.Key)
	err := dbQuery(dbOrTx, func(scanner Scanner) error {
		var actionKey identity.Key
		var parameterSubKey string
		var logicKey identity.Key
		if err := scanActionParameterSimulationSpec(scanner, &actionKey, &parameterSubKey, &logicKey); err != nil {
			return errors.WithStack(err)
		}
		paramKey, err := identity.NewParameterKey(actionKey, parameterSubKey)
		if err != nil {
			return errors.WithStack(err)
		}
		result[paramKey] = logicKey
		return nil
	}, `SELECT
			apss.action_key,
			apss.parameter_key,
			apss.logic_key
		FROM
			action_parameter_simulation_spec apss
		WHERE
			apss.model_key = $1
		ORDER BY apss.action_key, apss.parameter_key`, modelKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

// AddActionParameterSimulationRequires inserts simulation require join rows.
func AddActionParameterSimulationRequires(dbOrTx DbOrTx, modelKey string, requires map[identity.Key]map[string][]identity.Key) error {
	totalRows := 0
	for _, paramRequires := range requires {
		for _, logicKeys := range paramRequires {
			totalRows += len(logicKeys)
		}
	}
	if totalRows == 0 {
		return nil
	}
	var qb strings.Builder
	qb.WriteString(`INSERT INTO action_parameter_simulation_require (model_key, action_key, parameter_key, logic_key) VALUES `)
	args := make([]any, 0, totalRows*4)
	i := 0
	for actionKey, paramRequires := range requires {
		for parameterKey, logicKeys := range paramRequires {
			for _, logicKey := range logicKeys {
				if i > 0 {
					qb.WriteString(", ")
				}
				base := i * 4
				fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
				args = append(args, modelKey, actionKey.String(), parameterKey, logicKey.String())
				i++
			}
		}
	}
	return errors.WithStack(dbExec(dbOrTx, qb.String(), args...))
}

// AddActionParameterSimulationSpecs inserts simulation specification join rows.
func AddActionParameterSimulationSpecs(dbOrTx DbOrTx, modelKey string, specs map[identity.Key]map[string]identity.Key) error {
	if len(specs) == 0 {
		return nil
	}
	var qb strings.Builder
	qb.WriteString(`INSERT INTO action_parameter_simulation_spec (model_key, action_key, parameter_key, logic_key) VALUES `)
	args := make([]any, 0, len(specs)*4)
	i := 0
	for actionKey, paramSpecs := range specs {
		for parameterKey, logicKey := range paramSpecs {
			if i > 0 {
				qb.WriteString(", ")
			}
			base := i * 4
			fmt.Fprintf(&qb, "($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4)
			args = append(args, modelKey, actionKey.String(), parameterKey, logicKey.String())
			i++
		}
	}
	return errors.WithStack(dbExec(dbOrTx, qb.String(), args...))
}
