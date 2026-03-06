package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestCallExpressionSuite(t *testing.T) {
	suite.Run(t, new(CallExpressionSuite))
}

type CallExpressionSuite struct {
	suite.Suite
}

func (suite *CallExpressionSuite) TestString() {
	tests := []struct {
		testName     string
		modelScope   bool
		domain       *Identifier
		subdomain    *Identifier
		class        *Identifier
		functionName *Identifier
		parameter    Expression
		expected     string
	}{
		{
			testName:     `function only (class scope implied)`,
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
				},
			},
			expected: `Foo([a ↦ 1])`,
		},
		{
			testName:     `model scope with leading _`,
			modelScope:   true,
			functionName: &Identifier{Value: `GlobalFunc`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `x`}, Expression: NewIntLiteral(42)},
				},
			},
			expected: `_GlobalFunc([x ↦ 42])`,
		},
		{
			testName:     `class scoped`,
			class:        &Identifier{Value: `MyClass`},
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
				},
			},
			expected: `MyClass!Foo([a ↦ 1])`,
		},
		{
			testName:     `subdomain and class scoped`,
			subdomain:    &Identifier{Value: `Subdomain`},
			class:        &Identifier{Value: `Class`},
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `x`}, Expression: NewIntLiteral(42)},
				},
			},
			expected: `Subdomain!Class!Foo([x ↦ 42])`,
		},
		{
			testName:     `fully scoped`,
			domain:       &Identifier{Value: `Domain`},
			subdomain:    &Identifier{Value: `Subdomain`},
			class:        &Identifier{Value: `Class`},
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `x`}, Expression: NewIntLiteral(42)},
				},
			},
			expected: `Domain!Subdomain!Class!Foo([x ↦ 42])`,
		},
		{
			testName:     `fully scoped with complex record`,
			domain:       &Identifier{Value: `Domain`},
			subdomain:    &Identifier{Value: `Subdomain`},
			class:        &Identifier{Value: `Class`},
			functionName: &Identifier{Value: `FunctionName`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `p1`}, Expression: &Identifier{Value: `val1`}},
					{Field: &Identifier{Value: `p2`}, Expression: &StringLiteral{Value: `hello`}},
					{Field: &Identifier{Value: `p3`}, Expression: NewIntLiteral(100)},
				},
			},
			expected: `Domain!Subdomain!Class!FunctionName([p1 ↦ val1, p2 ↦ "hello", p3 ↦ 100])`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			c := &CallExpression{
				ModelScope:   tt.modelScope,
				Domain:       tt.domain,
				Subdomain:    tt.subdomain,
				Class:        tt.class,
				FunctionName: tt.functionName,
				Parameter:    tt.parameter,
			}
			assert.Equal(t, tt.expected, c.String())
		})
	}
}

func (suite *CallExpressionSuite) TestAscii() {
	tests := []struct {
		testName     string
		modelScope   bool
		domain       *Identifier
		subdomain    *Identifier
		class        *Identifier
		functionName *Identifier
		parameter    Expression
		expected     string
	}{
		{
			testName:     `function only`,
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
				},
			},
			expected: `Foo([a |-> 1])`,
		},
		{
			testName:     `model scope`,
			modelScope:   true,
			functionName: &Identifier{Value: `GlobalFunc`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `x`}, Expression: NewIntLiteral(42)},
				},
			},
			expected: `_GlobalFunc([x |-> 42])`,
		},
		{
			testName:     `class scoped`,
			class:        &Identifier{Value: `Class`},
			functionName: &Identifier{Value: `Foo`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
				},
			},
			expected: `Class!Foo([a |-> 1])`,
		},
		{
			testName:     `fully scoped with record`,
			domain:       &Identifier{Value: `Domain`},
			subdomain:    &Identifier{Value: `Subdomain`},
			class:        &Identifier{Value: `Class`},
			functionName: &Identifier{Value: `FunctionName`},
			parameter: &RecordInstance{
				Bindings: []*FieldBinding{
					{Field: &Identifier{Value: `p1`}, Expression: &Identifier{Value: `val1`}},
					{Field: &Identifier{Value: `p2`}, Expression: NewIntLiteral(2)},
				},
			},
			expected: `Domain!Subdomain!Class!FunctionName([p1 |-> val1, p2 |-> 2])`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			c := &CallExpression{
				ModelScope:   tt.modelScope,
				Domain:       tt.domain,
				Subdomain:    tt.subdomain,
				Class:        tt.class,
				FunctionName: tt.functionName,
				Parameter:    tt.parameter,
			}
			assert.Equal(t, tt.expected, c.Ascii())
		})
	}
}

func (suite *CallExpressionSuite) TestValidate() {
	tests := []struct {
		testName string
		c        *CallExpression
		errstr   string
	}{
		// OK.
		{
			testName: `valid function only`,
			c: &CallExpression{
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
		},
		{
			testName: `valid model scope`,
			c: &CallExpression{
				ModelScope:   true,
				FunctionName: &Identifier{Value: `GlobalFunc`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
		},
		{
			testName: `valid class scoped`,
			c: &CallExpression{
				Class:        &Identifier{Value: `Class`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
		},
		{
			testName: `valid subdomain and class scoped`,
			c: &CallExpression{
				Subdomain:    &Identifier{Value: `Subdomain`},
				Class:        &Identifier{Value: `Class`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
		},
		{
			testName: `valid fully scoped`,
			c: &CallExpression{
				Domain:       &Identifier{Value: `Domain`},
				Subdomain:    &Identifier{Value: `Subdomain`},
				Class:        &Identifier{Value: `Class`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
		},

		// Errors.
		{
			testName: `error missing function name`,
			c: &CallExpression{
				Domain: &Identifier{Value: `Domain`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `FunctionName`,
		},
		{
			testName: `error empty function name`,
			c: &CallExpression{
				FunctionName: &Identifier{Value: ``},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Value`,
		},
		{
			testName: `error missing parameter`,
			c: &CallExpression{
				FunctionName: &Identifier{Value: `Foo`},
			},
			errstr: `Parameter`,
		},
		{
			testName: `error model scope with domain`,
			c: &CallExpression{
				ModelScope:   true,
				Domain:       &Identifier{Value: `Domain`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Domain`,
		},
		{
			testName: `error model scope with subdomain`,
			c: &CallExpression{
				ModelScope:   true,
				Subdomain:    &Identifier{Value: `Subdomain`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Subdomain`,
		},
		{
			testName: `error model scope with class`,
			c: &CallExpression{
				ModelScope:   true,
				Class:        &Identifier{Value: `Class`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Class`,
		},
		{
			testName: `error domain without subdomain`,
			c: &CallExpression{
				Domain:       &Identifier{Value: `Domain`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Subdomain`,
		},
		{
			testName: `error domain without class`,
			c: &CallExpression{
				Domain:       &Identifier{Value: `Domain`},
				Subdomain:    &Identifier{Value: `Subdomain`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Class`,
		},
		{
			testName: `error subdomain without class`,
			c: &CallExpression{
				Subdomain:    &Identifier{Value: `Subdomain`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Class`,
		},
		{
			testName: `error empty domain`,
			c: &CallExpression{
				Domain:       &Identifier{Value: ``},
				Subdomain:    &Identifier{Value: `Subdomain`},
				Class:        &Identifier{Value: `Class`},
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Value`,
		},
		{
			testName: `error invalid parameter`,
			c: &CallExpression{
				FunctionName: &Identifier{Value: `Foo`},
				Parameter: &RecordInstance{
					Bindings: []*FieldBinding{
						{Field: &Identifier{Value: ``}, Expression: NewIntLiteral(1)},
					},
				},
			},
			errstr: `Value`,
		},
	}
	for _, tt := range tests {
		_ = suite.T().Run(tt.testName, func(t *testing.T) {
			err := tt.c.Validate()
			if tt.errstr == `` {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errstr)
			}
		})
	}
}

func (suite *CallExpressionSuite) TestExpressionNode() {
	// Verify that CallExpression implements the expressionNode interface method.
	c := &CallExpression{
		FunctionName: &Identifier{Value: `Foo`},
		Parameter: &RecordInstance{
			Bindings: []*FieldBinding{
				{Field: &Identifier{Value: `a`}, Expression: NewIntLiteral(1)},
			},
		},
	}
	// This should compile and not panic.
	c.expressionNode()
}
