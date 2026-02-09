package model_class

import (
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

const (
	MULTIPLICITY_0_1    = "0..1"    // An optional zero to 1 relationship.
	MULTIPLICITY_ANY    = "any"     // An optional many to many relationship.
	MULTIPLICITY_1      = "1"       // A required 1 to 1 relationship.
	MULTIPLICITY_1_MANY = "1..many" // A required 1 to many relationship.
)

// Multiplicity is how two classes relate to each other.
type Multiplicity struct {
	LowerBound  uint // Zero is "any".
	HigherBound uint // Zero is "any".
}

func NewMultiplicity(value string) (multiplicity Multiplicity, err error) {

	lowerBound, higherBound, err := parseMultiplicity(value)
	if err != nil {
		return Multiplicity{}, err
	}

	multiplicity = Multiplicity{
		LowerBound:  lowerBound,
		HigherBound: higherBound,
	}

	if err = multiplicity.Validate(); err != nil {
		return Multiplicity{}, err
	}

	return multiplicity, nil
}

// Validate validates the Multiplicity struct.
func (m *Multiplicity) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.HigherBound, validation.By(func(value interface{}) error {
			// If HigherBound is 0, it means "any" (unlimited), so no constraint.
			// If LowerBound is 0, it also means "any", so no constraint.
			// Only validate when both are non-zero: HigherBound must be >= LowerBound.
			if m.HigherBound != 0 && m.LowerBound != 0 && m.HigherBound < m.LowerBound {
				return errors.Errorf("higher bound (%d) must be >= lower bound (%d)", m.HigherBound, m.LowerBound)
			}
			return nil
		})),
	)
}

func (m *Multiplicity) String() string {

	// No bounds?
	if m.LowerBound == 0 && m.HigherBound == 0 {
		return "*"
	}

	// No upper bound?
	if m.HigherBound == 0 {
		return strconv.Itoa(int(m.LowerBound)) + "..*" // format of "2..*"
	}

	// Same number?
	if m.LowerBound == m.HigherBound {
		return strconv.Itoa(int(m.LowerBound)) // format of "2"
	}

	// Two numbers.
	return strconv.Itoa(int(m.LowerBound)) + ".." + strconv.Itoa(int(m.HigherBound)) // format of "2..3"
}

// The string that came from the parsed files, a bit different than what shows up in diagrams.
func (m *Multiplicity) ParsedString() (value string) {
	value = m.String()
	if value == "*" {
		value = MULTIPLICITY_ANY
	}
	return value
}

func parseMultiplicity(multiplicity string) (lowerBound, higherBound uint, err error) {

	// If any then there are no bounds.
	if multiplicity == MULTIPLICITY_ANY {
		return 0, 0, nil // Zeros are "any".
	}

	// If the entire string is a number, that is both the lower and higher bound.
	singleUint64, err := strconv.ParseUint(multiplicity, 10, 64)
	if err == nil {
		// No error!
		lowerBound = uint(singleUint64)
		higherBound = uint(singleUint64)
		return lowerBound, higherBound, nil
	}
	// There is an error here.

	// The other values are in the format of "1..2" or "any..many" or "any..any"
	multiplicityParts := strings.SplitN(multiplicity, "..", 2)
	if len(multiplicityParts) != 2 {
		return 0, 0, errors.WithStack(errors.Errorf(`invalid multiplicity: '%s'`, multiplicity))
	}
	lowerPart := multiplicityParts[0]
	higherPart := multiplicityParts[1]

	if lowerPart == "many" || lowerPart == "any" {
		lowerBound = 0 // Zeros are "any".
	} else {
		lowerUint64, err := strconv.ParseUint(lowerPart, 10, 64)
		if err != nil {
			return 0, 0, errors.WithStack(errors.Errorf(`invalid multiplicity: '%s'`, multiplicity))
		}
		lowerBound = uint(lowerUint64)
	}

	if higherPart == "many" || higherPart == "any" {
		higherBound = 0 // Zeros are "any".
	} else {
		higherUint64, err := strconv.ParseUint(higherPart, 10, 64)
		if err != nil {
			return 0, 0, errors.WithStack(errors.Errorf(`invalid multiplicity: '%s'`, multiplicity))
		}
		higherBound = uint(higherUint64)
	}

	return lowerBound, higherBound, nil
}
