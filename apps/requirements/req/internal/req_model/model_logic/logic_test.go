package model_logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogicTestSuite struct {
	suite.Suite
}

func TestLogicSuite(t *testing.T) {
	suite.Run(t, new(LogicTestSuite))
}

// TestValidate tests all validation rules for Logic.
func (s *LogicTestSuite) TestValidate() {
	tests := []struct {
		testName string
		logic    Logic
		errstr   string
	}{
		{
			testName: "valid minimal",
			logic: Logic{
				Id:          "inv_1",
				Description: "All orders must have at least one item.",
				Notation:    NotationTLAPlus,
			},
		},
		{
			testName: "valid with specification",
			logic: Logic{
				Id:            "inv_2",
				Description:   "Stock is never negative.",
				Notation:      NotationTLAPlus,
				Specification: "\\A p \\in Products : p.stock >= 0",
			},
		},
		{
			testName: "valid with empty specification",
			logic: Logic{
				Id:            "inv_3",
				Description:   "Placeholder invariant.",
				Notation:      NotationTLAPlus,
				Specification: "",
			},
		},
		{
			testName: "error missing id",
			logic: Logic{
				Id:          "",
				Description: "Some description.",
				Notation:    NotationTLAPlus,
			},
			errstr: "Id",
		},
		{
			testName: "error missing description",
			logic: Logic{
				Id:          "inv_1",
				Description: "",
				Notation:    NotationTLAPlus,
			},
			errstr: "Description",
		},
		{
			testName: "error missing notation",
			logic: Logic{
				Id:          "inv_1",
				Description: "Some description.",
				Notation:    "",
			},
			errstr: "Notation",
		},
		{
			testName: "error invalid notation",
			logic: Logic{
				Id:          "inv_1",
				Description: "Some description.",
				Notation:    "Z",
			},
			errstr: "Notation",
		},
		{
			testName: "error missing id and description",
			logic: Logic{
				Id:          "",
				Description: "",
				Notation:    NotationTLAPlus,
			},
			errstr: "Id",
		},
		{
			testName: "error missing id with specification set",
			logic: Logic{
				Id:            "",
				Description:   "Some description.",
				Notation:      NotationTLAPlus,
				Specification: "TRUE",
			},
			errstr: "Id",
		},
		{
			testName: "error invalid notation with specification set",
			logic: Logic{
				Id:            "inv_1",
				Description:   "Some description.",
				Notation:      "Alloy",
				Specification: "some spec",
			},
			errstr: "Notation",
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.testName, func(t *testing.T) {
			err := tt.logic.Validate()
			if tt.errstr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errstr)
			}
		})
	}
}

// TestNew tests that NewLogic maps parameters correctly and calls Validate.
func (s *LogicTestSuite) TestNew() {
	// Test all parameters are mapped correctly.
	logic, err := NewLogic("inv_1", "Stock is never negative.", NotationTLAPlus, "\\A p \\in Products : p.stock >= 0")
	s.NoError(err)
	s.Equal(Logic{
		Id:            "inv_1",
		Description:   "Stock is never negative.",
		Notation:      NotationTLAPlus,
		Specification: "\\A p \\in Products : p.stock >= 0",
	}, logic)

	// Test with empty specification (optional).
	logic, err = NewLogic("inv_2", "Placeholder.", NotationTLAPlus, "")
	s.NoError(err)
	s.Equal(Logic{
		Id:          "inv_2",
		Description: "Placeholder.",
		Notation:    NotationTLAPlus,
	}, logic)

	// Test that Validate is called (invalid data should fail).
	_, err = NewLogic("", "Some description.", NotationTLAPlus, "")
	s.Error(err)
	s.Contains(err.Error(), "Id")

	// Test that invalid notation fails.
	_, err = NewLogic("inv_1", "Some description.", "Z", "")
	s.Error(err)
	s.Contains(err.Error(), "Notation")
}
