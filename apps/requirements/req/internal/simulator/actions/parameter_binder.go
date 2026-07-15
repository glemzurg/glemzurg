package actions

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strings"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_data_type"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/core/model_state"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/evaluator"
	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/simulator/object"
)

// nullableNullSampleDenom is the denominator for sampling NULL on nullable parameters (1/denom chance).
const nullableNullSampleDenom = 10

const (
	spanBoundUnconstrained = "unconstrained"
	spanBoundOpen          = "open"
	// spanDefaultHalfWidth scales unconstrained span ends: ±(halfWidth × precision).
	spanDefaultHalfWidth = 100
)

// ParameterBinder validates and generates parameter values for actions and queries.
type ParameterBinder struct{}

// NewParameterBinder creates a new parameter binder.
func NewParameterBinder() *ParameterBinder {
	return &ParameterBinder{}
}

// BindParameters validates that all required parameters are provided and returns
// a map suitable for use as evaluator bindings.
func (b *ParameterBinder) BindParameters(
	paramDefs []model_state.Parameter,
	paramValues map[string]object.Object,
) (map[string]object.Object, error) {
	result := make(map[string]object.Object)

	for _, paramDef := range paramDefs {
		value, exists := paramValues[paramDef.Name]
		if !exists {
			return nil, fmt.Errorf("missing required parameter: %s", paramDef.Name)
		}
		result[paramDef.Name] = CoerceValueForDataType(paramDef.DataType, value)
	}

	return result, nil
}

// GenerateRandomParameters creates random parameter values that satisfy data type constraints.
// Used by the simulation engine to drive random exploration.
func (b *ParameterBinder) GenerateRandomParameters(
	paramDefs []model_state.Parameter,
	rng *rand.Rand,
) map[string]object.Object {
	result := make(map[string]object.Object)

	for _, paramDef := range paramDefs {
		result[paramDef.Name] = sampleParameterValue(paramDef, rng)
	}

	coerceSampledParameters(paramDefs, result)
	return result
}

// sampleParameterValue generates a random value for one action/query parameter.
// Nullable parameters may be NULL; non-nullable parameters never are.
func sampleParameterValue(param model_state.Parameter, rng *rand.Rand) object.Object {
	if param.Nullable && rng.Intn(nullableNullSampleDenom) == 0 {
		return evaluator.EMPTY_SET
	}
	return generateRandomValue(param.DataType, rng)
}

// generateRandomValue creates a random non-null value based on data type constraints.
func generateRandomValue(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	if dataType == nil {
		return randomDefaultNumber(rng)
	}
	if coll := generateRandomCollection(dataType, rng); coll != nil {
		return coll
	}
	if values := model_data_type.EnumerationValues(dataType); len(values) > 0 {
		return randomEnumerationValue(dataType, values, rng)
	}
	if val, ok := generateFromTypeSpec(dataType, rng); ok {
		return val
	}
	if dataType.Atomic == nil {
		return randomDefaultNumber(rng)
	}
	return generateAtomicRandomValue(dataType, dataType.Atomic, rng)
}

// generateRandomCollection samples set/tuple/record collections; nil when dataType is atomic.
func generateRandomCollection(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	switch dataType.CollectionType {
	case model_data_type.COLLECTION_TYPE_UNORDERED:
		return randomUnorderedCollection(dataType, rng)
	case model_data_type.COLLECTION_TYPE_ORDERED,
		model_data_type.COLLECTION_TYPE_QUEUE,
		model_data_type.COLLECTION_TYPE_STACK:
		return randomOrderedCollection(dataType, rng)
	case model_data_type.COLLECTION_TYPE_RECORD:
		return randomRecordValue(dataType, rng)
	default:
		return nil
	}
}

func generateFromTypeSpec(dataType *model_data_type.DataType, rng *rand.Rand) (object.Object, bool) {
	if dataType.TypeSpec == nil {
		return nil, false
	}
	switch strings.ToUpper(strings.TrimSpace(dataType.TypeSpec.Specification)) {
	case "STRING":
		return randomString(rng), true
	case "BOOLEAN":
		return object.NewBoolean(rng.Intn(2) != 0), true
	default:
		return nil, false
	}
}

func generateAtomicRandomValue(
	dataType *model_data_type.DataType,
	atomic *model_data_type.Atomic,
	rng *rand.Rand,
) object.Object {
	switch atomic.ConstraintType {
	case model_data_type.CONSTRAINT_TYPE_SPAN:
		return randomNumberInSpan(atomic.Span, rng)
	case model_data_type.CONSTRAINT_TYPE_ENUMERATION:
		if len(atomic.Enums) == 0 {
			return evaluator.EMPTY_SET
		}
		values := make([]string, len(atomic.Enums))
		for i, enum := range atomic.Enums {
			values[i] = enum.Value
		}
		return randomEnumerationValue(dataType, values, rng)
	case model_data_type.CONSTRAINT_TYPE_UNCONSTRAINED:
		return randomString(rng)
	case model_data_type.CONSTRAINT_TYPE_DATETIME:
		return randomDateTimeValue(rng)
	case model_data_type.CONSTRAINT_TYPE_REFERENCE, model_data_type.CONSTRAINT_TYPE_OBJECT:
		return randomString(rng)
	default:
		return randomDefaultNumber(rng)
	}
}

// collectionElementType is the type of one collection member.
// Simple "unordered of enum/span" forms keep the element Atomic on the collection itself.
func collectionElementType(dataType *model_data_type.DataType) *model_data_type.DataType {
	if dataType == nil {
		return nil
	}
	if dataType.ElementDataType != nil {
		return dataType.ElementDataType
	}
	if dataType.Atomic != nil {
		return &model_data_type.DataType{
			CollectionType: model_data_type.COLLECTION_TYPE_ATOMIC,
			Atomic:         dataType.Atomic,
			TypeSpec:       dataType.TypeSpec,
		}
	}
	return nil
}

func collectionSampleSize(dataType *model_data_type.DataType, uniqueCap int, rng *rand.Rand) int {
	minN := 1
	if dataType.CollectionMin != nil {
		minN = *dataType.CollectionMin
	}
	maxN := minN + 2
	if dataType.CollectionMax != nil {
		maxN = *dataType.CollectionMax
	}
	if uniqueCap > 0 && maxN > uniqueCap {
		maxN = uniqueCap
	}
	if maxN < minN {
		maxN = minN
	}
	if maxN == minN {
		return minN
	}
	return minN + rng.Intn(maxN-minN+1)
}

func randomUnorderedCollection(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	elemType := collectionElementType(dataType)
	unique := dataType.CollectionUnique != nil && *dataType.CollectionUnique

	// Unique finite enums: sample a non-empty subset (model-agnostic default).
	if unique && elemType != nil {
		if values := model_data_type.EnumerationValues(elemType); len(values) > 0 {
			return randomUniqueEnumSubset(elemType, values, dataType, rng)
		}
	}

	n := collectionSampleSize(dataType, 0, rng)
	set := object.NewSet()
	// Bound attempts so unique sampling cannot spin forever on a tiny domain.
	for attempts := 0; set.Size() < n && attempts < n*8+8; attempts++ {
		elem := generateRandomValue(elemType, rng)
		if unique && set.Contains(elem) {
			continue
		}
		set.Add(elem)
	}
	return set
}

func randomUniqueEnumSubset(
	elemType *model_data_type.DataType,
	values []string,
	collType *model_data_type.DataType,
	rng *rand.Rand,
) object.Object {
	n := min(max(collectionSampleSize(collType, len(values), rng), 1), len(values))
	// Fisher–Yates partial shuffle for a uniform n-subset.
	order := append([]string(nil), values...)
	for i := range n {
		j := i + rng.Intn(len(order)-i)
		order[i], order[j] = order[j], order[i]
	}
	elems := make([]object.Object, 0, n)
	for i := range n {
		elems = append(elems, valueForEnumerationLiteral(elemType, order[i]))
	}
	return object.NewSetFromElements(elems)
}

func randomOrderedCollection(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	elemType := collectionElementType(dataType)
	n := collectionSampleSize(dataType, 0, rng)
	elems := make([]object.Object, 0, n)
	seen := object.NewSet()
	unique := dataType.CollectionUnique != nil && *dataType.CollectionUnique
	for attempts := 0; len(elems) < n && attempts < n*8+8; attempts++ {
		elem := generateRandomValue(elemType, rng)
		if unique && seen.Contains(elem) {
			continue
		}
		seen.Add(elem)
		elems = append(elems, elem)
	}
	return object.NewTupleFromElements(elems)
}

func randomRecordValue(dataType *model_data_type.DataType, rng *rand.Rand) object.Object {
	fields := make(map[string]object.Object, len(dataType.RecordFields))
	for _, field := range dataType.RecordFields {
		fields[field.Name] = generateRandomValue(field.FieldDataType, rng)
	}
	return object.NewRecordFromFields(fields)
}

func randomDefaultNumber(rng *rand.Rand) object.Object {
	return object.NewNatural(rng.Int63n(100))
}

func randomDateTimeValue(rng *rand.Rand) object.Object {
	span := model_data_type.DateTimeValueMax - model_data_type.DateTimeValueMin + 1
	return object.NewNatural(model_data_type.DateTimeValueMin + rng.Int63n(span))
}

func randomString(rng *rand.Rand) object.Object {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := 1 + rng.Intn(8)
	var b strings.Builder
	for range length {
		b.WriteByte(letters[rng.Intn(len(letters))])
	}
	return object.NewString(b.String())
}

// randomNumberInSpan generates a random Real on a span's precision lattice.
func randomNumberInSpan(span *model_data_type.AtomicSpan, rng *rand.Rand) object.Object {
	step := spanPrecision(span)
	if span == nil {
		extent := spanDefaultHalfWidth * step
		return randomRealOnPrecisionGrid(-extent, extent, step, rng)
	}

	lower, upper := spanSamplingInterval(span)
	if upper <= lower {
		return object.NewFloat(lower)
	}

	return randomRealOnPrecisionGrid(lower, upper, step, rng)
}

func spanPrecision(span *model_data_type.AtomicSpan) float64 {
	if span == nil || span.Precision <= 0 {
		return 1
	}
	return span.Precision
}

func randomRealOnPrecisionGrid(lower, upper, step float64, rng *rand.Rand) object.Object {
	if step <= 0 {
		return object.NewFloat(lower)
	}

	lowerSteps := int(math.Round(lower / step))
	upperSteps := int(math.Round(upper / step))
	if upperSteps < lowerSteps {
		return object.NewFloat(lower)
	}

	stepIndex := lowerSteps + rng.Intn(upperSteps-lowerSteps+1)
	return object.NewFloat(float64(stepIndex) * step)
}

func spanSamplingInterval(span *model_data_type.AtomicSpan) (lower, upper float64) {
	step := spanPrecision(span)
	defaultExtent := spanDefaultHalfWidth * step

	lower = -defaultExtent
	if span.LowerType != spanBoundUnconstrained && span.LowerValue != nil {
		lower, _ = spanValueToRat(span.LowerValue, span.LowerDenominator).Float64()
		if span.LowerType == spanBoundOpen {
			lower += step
		}
	}

	upper = defaultExtent
	if span.HigherType != spanBoundUnconstrained && span.HigherValue != nil {
		upper, _ = spanValueToRat(span.HigherValue, span.HigherDenominator).Float64()
		if span.HigherType == spanBoundOpen {
			upper -= step
		}
	}

	return lower, upper
}

func spanValueToRat(value *int, denom *int) *big.Rat {
	if value == nil {
		return big.NewRat(0, 1)
	}
	denomVal := 1
	if denom != nil {
		denomVal = *denom
	}
	return big.NewRat(int64(*value), int64(denomVal))
}
