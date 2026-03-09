package parser

import (
	"testing"
)

// FuzzParseExpression ensures ParseExpression never panics on arbitrary input.
func FuzzParseExpression(f *testing.F) {
	// Seed corpus with valid TLA+ expressions and edge cases.
	f.Add("TRUE")
	f.Add("FALSE")
	f.Add("x + y")
	f.Add("x' = 42")
	f.Add(`\A x \in S : x > 0`)
	f.Add(`\E x \in S : x = 0`)
	f.Add("IF x THEN y ELSE z")
	f.Add("CASE x -> 1 [] y -> 2")
	f.Add("[a |-> 1, b |-> 2]")
	f.Add("{1, 2, 3}")
	f.Add("<<1, 2, 3>>")
	f.Add(`"hello"`)
	f.Add("")
	f.Add("(((())))")
	f.Add("x.y.z")
	f.Add("f(x, y)")
	f.Add("x \\in S")
	f.Add("x \\notin S")
	f.Add("DOMAIN f")
	f.Add("Len(s)")
	f.Add("~x")
	f.Add("x /\\ y")
	f.Add("x \\/ y")
	f.Add("x => y")
	f.Add("LET x == 1 IN x + 1")
	f.Add("CHOOSE x \\in S : x > 0")

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic — either returns a valid AST or an error.
		_, _ = ParseExpression(input)
	})
}
