package object

import (
	"fmt"
	"math"
	"math/big"
)

// NumberKind indicates the current representation of a Number.
type NumberKind int

const (
	KindNatural  NumberKind = iota // Non-negative integer (0, 1, 2, ...)
	KindInteger                    // Any integer (..., -2, -1, 0, 1, 2, ...)
	KindRational                   // Exact rational number with fractional part (1/2, 3/4, ...)
	KindReal                       // Real number including irrationals, stored as float64
)

// Number is a unified numeric type that can represent Natural, Integer, Rational, or Real values.
// For exact arithmetic (Natural, Integer, Rational), it uses *big.Rat for arbitrary precision.
// For irrational results (Real), it falls back to float64. Once a computation produces an
// irrational result, all subsequent operations with that number will produce Real results
// (contamination).
type Number struct {
	rat   *big.Rat // Used for exact arithmetic (Natural, Integer, Rational)
	float float64  // Used for irrational numbers (Real)
	real  bool     // True if this is a Real (irrational) number using float storage
}

// NewNatural creates a Number from a non-negative integer value.
// Panics if value is negative.
func NewNatural(value int64) *Number {
	if value < 0 {
		panic("NewNatural requires a non-negative value")
	}
	return &Number{rat: new(big.Rat).SetInt64(value)}
}

// NewInteger creates a Number from an integer value.
func NewInteger(value int64) *Number {
	return &Number{rat: new(big.Rat).SetInt64(value)}
}

// NewRational creates a Number from a numerator and denominator (a rational number).
func NewRational(num, denom int64) *Number {
	return &Number{rat: big.NewRat(num, denom)}
}

// NewReal creates a Number from a numerator and denominator (a rational number).
// Deprecated: Use NewRational for exact rationals. This is kept for backwards compatibility.
func NewReal(num, denom int64) *Number {
	return NewRational(num, denom)
}

// NewFloat creates a Number from a float64 value.
// This creates a Real (potentially irrational) number.
func NewFloat(value float64) *Number {
	return &Number{float: value, real: true}
}

// newRealFromFloat creates a Real number from a float64 (internal use).
func newRealFromFloat(value float64) *Number {
	return &Number{float: value, real: true}
}

func (n *Number) Type() ObjectType { return TypeNumber }

func (n *Number) Inspect() string {
	if n.real {
		// Format float without unnecessary trailing zeros
		return fmt.Sprintf("%g", n.float)
	}
	return n.rat.RatString()
}

// Kind returns the current kind of the number (Natural, Integer, Rational, or Real).
func (n *Number) Kind() NumberKind {
	if n.real {
		return KindReal
	}
	if !n.rat.IsInt() {
		return KindRational
	}
	if n.rat.Sign() >= 0 {
		return KindNatural
	}
	return KindInteger
}

// IsReal returns true if this is a Real (irrational) number.
func (n *Number) IsReal() bool {
	return n.real
}

// Rat returns a copy of the underlying big.Rat value.
// For Real numbers, this converts the float to a rational approximation.
func (n *Number) Rat() *big.Rat {
	if n.real {
		r := new(big.Rat)
		r.SetFloat64(n.float)
		return r
	}
	return new(big.Rat).Set(n.rat)
}

// Float64 returns the number as a float64.
// For exact rationals, this may lose precision.
func (n *Number) Float64() float64 {
	if n.real {
		return n.float
	}
	f, _ := n.rat.Float64()
	return f
}

func (n *Number) SetValue(source Object) error {
	src, ok := source.(*Number)
	if !ok {
		return fmt.Errorf("cannot assign %T to Number", source)
	}
	if src.real {
		n.real = true
		n.float = src.float
		n.rat = nil
	} else {
		n.real = false
		n.rat = new(big.Rat).Set(src.rat)
		n.float = 0
	}
	return nil
}

func (n *Number) Clone() Object {
	if n.real {
		return &Number{float: n.float, real: true}
	}
	return &Number{rat: new(big.Rat).Set(n.rat)}
}

// Add returns a new Number that is the sum of n and other.
// If either operand is Real, the result is Real.
func (n *Number) Add(other *Number) *Number {
	if n.real || other.real {
		return newRealFromFloat(n.Float64() + other.Float64())
	}
	result := new(big.Rat).Add(n.rat, other.rat)
	return &Number{rat: result}
}

// Sub returns a new Number that is the difference of n and other.
// If either operand is Real, the result is Real.
func (n *Number) Sub(other *Number) *Number {
	if n.real || other.real {
		return newRealFromFloat(n.Float64() - other.Float64())
	}
	result := new(big.Rat).Sub(n.rat, other.rat)
	return &Number{rat: result}
}

// Mul returns a new Number that is the product of n and other.
// If either operand is Real, the result is Real.
func (n *Number) Mul(other *Number) *Number {
	if n.real || other.real {
		return newRealFromFloat(n.Float64() * other.Float64())
	}
	result := new(big.Rat).Mul(n.rat, other.rat)
	return &Number{rat: result}
}

// Div returns a new Number that is n divided by other.
// If either operand is Real, the result is Real.
func (n *Number) Div(other *Number) *Number {
	if n.real || other.real {
		return newRealFromFloat(n.Float64() / other.Float64())
	}
	result := new(big.Rat).Quo(n.rat, other.rat)
	return &Number{rat: result}
}

// IntDiv returns a new Number that is the integer division of n by other.
// Both operands must be integers (not Rational or Real).
func (n *Number) IntDiv(other *Number) (*Number, error) {
	if n.real || other.real {
		return nil, fmt.Errorf("integer division requires integer operands, got Real")
	}
	if n.Kind() == KindRational || other.Kind() == KindRational {
		return nil, fmt.Errorf("integer division requires integer operands")
	}
	if other.rat.Sign() == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	// Perform integer division: floor(n/other) for positive, toward zero for mixed signs
	quo := new(big.Int).Quo(n.rat.Num(), other.rat.Num())
	return &Number{rat: new(big.Rat).SetInt(quo)}, nil
}

// Mod returns a new Number that is the remainder of n divided by other.
// Both operands must be integers (not Rational or Real).
func (n *Number) Mod(other *Number) (*Number, error) {
	if n.real || other.real {
		return nil, fmt.Errorf("modulo requires integer operands, got Real")
	}
	if n.Kind() == KindRational || other.Kind() == KindRational {
		return nil, fmt.Errorf("modulo requires integer operands")
	}
	if other.rat.Sign() == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	rem := new(big.Int).Rem(n.rat.Num(), other.rat.Num())
	return &Number{rat: new(big.Rat).SetInt(rem)}, nil
}

// Neg returns a new Number that is the negation of n.
func (n *Number) Neg() *Number {
	if n.real {
		return newRealFromFloat(-n.float)
	}
	result := new(big.Rat).Neg(n.rat)
	return &Number{rat: result}
}

// Abs returns a new Number that is the absolute value of n.
func (n *Number) Abs() *Number {
	if n.real {
		return newRealFromFloat(math.Abs(n.float))
	}
	result := new(big.Rat).Abs(n.rat)
	return &Number{rat: result}
}

// Cmp compares n with other and returns:
//
//	-1 if n < other
//	 0 if n == other
//	+1 if n > other
func (n *Number) Cmp(other *Number) int {
	if n.real || other.real {
		nf := n.Float64()
		of := other.Float64()
		if nf < of {
			return -1
		}
		if nf > of {
			return 1
		}
		return 0
	}
	return n.rat.Cmp(other.rat)
}

// Equals returns true if n and other represent the same numeric value.
func (n *Number) Equals(other *Number) bool {
	if n.real || other.real {
		return n.Float64() == other.Float64()
	}
	return n.rat.Cmp(other.rat) == 0
}

// Sign returns -1, 0, or +1 depending on whether n is negative, zero, or positive.
func (n *Number) Sign() int {
	if n.real {
		if n.float < 0 {
			return -1
		}
		if n.float > 0 {
			return 1
		}
		return 0
	}
	return n.rat.Sign()
}

// IsZero returns true if n is zero.
func (n *Number) IsZero() bool {
	if n.real {
		return n.float == 0
	}
	return n.rat.Sign() == 0
}

// Pow returns a new Number that is n raised to the power of exp.
// Supports both integer exponents and fractional exponents (roots).
// For fractional exponent p/q, computes (n^p)^(1/q) = q-th root of n^p.
// If the result is irrational, returns a Real number (float64).
func (n *Number) Pow(exp *Number) (*Number, error) {
	// If either operand is Real, use float arithmetic
	if n.real || exp.real {
		base := n.Float64()
		exponent := exp.Float64()
		if exponent < 0 {
			return nil, fmt.Errorf("exponent must be non-negative")
		}
		if base == 0 && exponent == 0 {
			return nil, fmt.Errorf("0^0 is undefined")
		}
		return newRealFromFloat(math.Pow(base, exponent)), nil
	}

	if exp.Sign() < 0 {
		return nil, fmt.Errorf("exponent must be non-negative")
	}

	// Get numerator and denominator of exponent
	expNum := exp.rat.Num().Int64()     // p in p/q
	expDenom := exp.rat.Denom().Int64() // q in p/q

	// Special cases
	if expNum == 0 {
		// n^0 = 1 (for all n != 0)
		if n.IsZero() {
			return nil, fmt.Errorf("0^0 is undefined")
		}
		return NewNatural(1), nil
	}

	// First compute n^p (numerator of exponent)
	// For rational base a/b, (a/b)^p = a^p / b^p
	baseNum := n.rat.Num()
	baseDenom := n.rat.Denom()

	powNum := new(big.Int).Exp(baseNum, big.NewInt(expNum), nil)
	powDenom := new(big.Int).Exp(baseDenom, big.NewInt(expNum), nil)

	// If exponent is an integer (denom = 1), we're done
	if expDenom == 1 {
		result := new(big.Rat).SetFrac(powNum, powDenom)
		return &Number{rat: result}, nil
	}

	// For fractional exponent, compute q-th root
	// (a/b)^(1/q) = a^(1/q) / b^(1/q)
	// This is only exact if both a and b are perfect q-th powers

	rootNum, ok := intRoot(powNum, expDenom)
	if !ok {
		// Result is irrational, fall back to float
		base := n.Float64()
		exponent := exp.Float64()
		return newRealFromFloat(math.Pow(base, exponent)), nil
	}

	rootDenom, ok := intRoot(powDenom, expDenom)
	if !ok {
		// Result is irrational, fall back to float
		base := n.Float64()
		exponent := exp.Float64()
		return newRealFromFloat(math.Pow(base, exponent)), nil
	}

	result := new(big.Rat).SetFrac(rootNum, rootDenom)
	return &Number{rat: result}, nil
}

// intRoot computes the n-th root of x if it's exact, returning (root, true).
// If x is not a perfect n-th power, returns (nil, false).
func intRoot(x *big.Int, n int64) (*big.Int, bool) {
	if x.Sign() == 0 {
		return big.NewInt(0), true
	}
	if x.Sign() < 0 {
		// Negative numbers don't have real even roots
		if n%2 == 0 {
			return nil, false
		}
		// For odd roots of negative numbers, compute root of absolute value and negate
		absX := new(big.Int).Abs(x)
		root, ok := intRoot(absX, n)
		if !ok {
			return nil, false
		}
		return root.Neg(root), true
	}

	// Use Newton's method to find integer n-th root
	// Start with an initial guess
	root := new(big.Int).Set(x)
	nBig := big.NewInt(n)
	nMinus1 := big.NewInt(n - 1)

	// Newton's method: root = ((n-1)*root + x/root^(n-1)) / n
	for {
		// Compute root^(n-1)
		rootPowNMinus1 := new(big.Int).Exp(root, nMinus1, nil)
		if rootPowNMinus1.Sign() == 0 {
			break
		}

		// Compute x / root^(n-1)
		quotient := new(big.Int).Div(x, rootPowNMinus1)

		// Compute (n-1)*root + quotient
		newRoot := new(big.Int).Mul(nMinus1, root)
		newRoot.Add(newRoot, quotient)
		newRoot.Div(newRoot, nBig)

		// Check for convergence
		if newRoot.Cmp(root) >= 0 {
			break
		}
		root = newRoot
	}

	// Verify: root^n should equal x
	check := new(big.Int).Exp(root, nBig, nil)
	if check.Cmp(x) == 0 {
		return root, true
	}

	// Try root+1 in case we undershot
	rootPlus1 := new(big.Int).Add(root, big.NewInt(1))
	check = new(big.Int).Exp(rootPlus1, nBig, nil)
	if check.Cmp(x) == 0 {
		return rootPlus1, true
	}

	return nil, false
}
