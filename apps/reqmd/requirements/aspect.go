package requirements

import (
	"fmt"
	"strconv"
	"strings"
)

type Aspect struct {
	name      string
	value     string
	sortValue string // Padded with appropriate zeros to sort leading numbers correctly alphabetically.
}

func newAspect(name, value string) (aspect Aspect, err error) {

	aspect = Aspect{
		name:  name,
		value: value,
	}

	return aspect, nil
}

func (a *Aspect) SetSortValue(padWidth uint) {
	a.sortValue = createSortValue(padWidth, a.value)
}

func createSortValue(padWidth uint, value string) (sortValue string) {

	// Start with our value.
	sortValue = strings.ToLower(value)

	// Strip off whitespace and punctuation.
	sortValue = strings.Trim(sortValue, "\n\t\"\\ !@#$%^&*()-_=+'[{|}];:,<.>/?")

	// How many numbers are there leading the string.
	sortValueNoNum := strings.TrimLeft(sortValue, "0123456789")

	// Was there a number?
	if len(sortValueNoNum) != len(sortValue) {

		// What is the number on its own.
		sortValueNum, found := strings.CutSuffix(sortValue, sortValueNoNum)
		if !found {
			panic(`expected numeric leading`)
		}

		// Convert into a integer.
		num, err := strconv.Atoi(sortValueNum)
		if err != nil {
			panic(err)
		}

		// Recreate with numeric padding.
		sortValue = fmt.Sprintf(`%0`+strconv.Itoa(int(padWidth))+`d%s`, num, sortValueNoNum)
	}

	return sortValue
}

func (a *Aspect) valuePadWidth() (padWidth uint) {

	// This padding is all for sorting.
	// Start with the sort values.
	sortValue := createSortValue(0, a.value)

	// Trim any numbers to find their count.
	sortValueNoNumber := strings.TrimLeft(sortValue, "0123456789")

	// How many were removed.
	padWidth = uint(len(sortValue) - len(sortValueNoNumber))

	return padWidth
}

const (
	_DISTILLED_ASPECT_HEADER = `| aspect | value |`
)

func isAspectHeader(textline string) (isHeader bool) {

	// Examine without case.
	textline = strings.ToLower(textline)

	// Split based on `|` then reassemble.
	fragments := strings.Split(textline, `|`)
	for i, fragment := range fragments {
		// Strip all spaces and punctuation excluding `|`.
		fragment = puncToWhitespace(fragment)
		fragments[i] = fragment
	}
	textline = strings.Join(fragments, `|`)

	// Cleanup.
	textline = normalizeWhitespace(textline)
	textline = strings.TrimSpace(textline)

	isHeader = (textline == _DISTILLED_ASPECT_HEADER)

	return isHeader
}

func isAspectHeaderLine(textline string) (is bool) {
	return _aspectTableHeaderLineRegexp.MatchString(textline)
}

func parseAspectValue(textline string) (aspect, value string, is bool) {

	// Header lines don't count.
	if isAspectHeaderLine(textline) {
		return "", "", false
	}

	matches := _aspectTableValueRegexp.FindAllStringSubmatch(textline, -1)

	// No match, not an aspect line.
	if len(matches) == 0 {
		return "", "", false
	}

	// Get parts.
	aspect = matches[0][1]
	value = matches[0][2]

	// Clean up.
	aspect = strings.TrimSpace(normalizeWhitespace(puncToWhitespace(strings.ToLower(aspect))))
	value = strings.TrimSpace(normalizeWhitespace(value))

	return aspect, value, true
}

func extractAspects(body string) (updated string, aspects []Aspect, err error) {
	body = strings.TrimSpace(body)

	// Examine the requirement body line by line.
	textlines := strings.Split(body, "\n")

	// Keep track of whether we're in an aspect table or not.
	inAspects := false

	for _, textline := range textlines {
		trimmedTextline := strings.TrimSpace(textline)

		// What does this line loook like?
		name, value, isValue := parseAspectValue(trimmedTextline)
		switch {
		case isAspectHeader(trimmedTextline):
			inAspects = true
		case inAspects && isAspectHeaderLine(trimmedTextline):
			// Nothing, consume this line.
		case inAspects && isValue:
			aspect, err := newAspect(name, value)
			if err != nil {
				return "", nil, err
			}
			aspects = append(aspects, aspect)
		default:
			// This is not an aspect line. Just part of the body.
			inAspects = false
			updated += "\n" + textline
		}
	}

	updated = strings.TrimSpace(updated)

	return updated, aspects, nil
}

const (
	_nameTitle  = "Aspect"
	_valueTitle = "Value"
)

func generateAspectBlock(orderAspects []string, aspects []Aspect) (block string, err error) {

	// What is the width of each table column.
	nameWidth := len(_nameTitle)   // Have to be at least as wide as the title.
	valueWidth := len(_valueTitle) // Have to be at least as weide as the title.
	for _, orderAspect := range orderAspects {
		if len(orderAspect) > nameWidth {
			nameWidth = len(orderAspect)
		}
	}
	for _, aspect := range aspects {
		if len(aspect.name) > nameWidth {
			nameWidth = len(aspect.name)
		}
		if len(aspect.value) > valueWidth {
			valueWidth = len(aspect.value)
		}
	}

	// Is there an aspect line? We may not even need the header.
	hasAspectLine := false

	// First put in expected elements.
	for _, orderAspect := range orderAspects {

		// Get name and value.
		name := orderAspect
		value := ""

		// Can we find a value?
		var updatedAspects []Aspect
		for _, aspect := range aspects {
			if strings.TrimSpace(strings.ToLower(name)) == strings.TrimSpace(strings.ToLower(aspect.name)) {
				value = strings.TrimSpace(aspect.value)
			} else {
				// Keep this aspect for the next run.
				updatedAspects = append(updatedAspects, aspect)
			}
		}
		aspects = updatedAspects

		// Generate a line.
		block += fmt.Sprintf(`| %-`+strconv.Itoa(nameWidth)+`s | %-`+strconv.Itoa(valueWidth)+"s |\n", name, value)
		hasAspectLine = true

	}

	// Finish with any aspects that are unplanned.
	for _, aspect := range aspects {

		// Drop out anything that has no value and is not an expected attribute.
		if strings.TrimSpace(aspect.value) != "" {
			// Generate a line.
			block += fmt.Sprintf(`| %-`+strconv.Itoa(nameWidth)+`s | %-`+strconv.Itoa(valueWidth)+"s |\n", aspect.name, aspect.value)
			hasAspectLine = true
		}
	}

	if hasAspectLine {
		// Generate the initial block table.
		header := fmt.Sprintf(`| %-`+strconv.Itoa(nameWidth)+`s | %-`+strconv.Itoa(valueWidth)+"s |\n", _nameTitle, _valueTitle)
		header += fmt.Sprintf(`|-` + strings.Repeat(`-`, nameWidth) + `-|-` + strings.Repeat(`-`, valueWidth) + "-|\n")
		block = header + block
	}

	block = strings.TrimSpace(block)

	return block, nil
}
