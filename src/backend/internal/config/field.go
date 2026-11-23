package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// UnitValue is a value together with the unit it was written with.
type UnitValue[T comparable] struct {
	Value T
	Unit  string
}

// SplitUnit separates a value from its trailing unit.
//
// A unit is the run of letters at the end of the value, so the separator is whatever sits between the digits
// and that run: nothing, a space, or any amount of whitespace. The split starts from the right, which keeps the
// number whole even for a value written without a separator.
func SplitUnit(text string) (number string, unit string) {
	trimmed := strings.TrimSpace(text)

	index := len(trimmed)
	for index > 0 {
		letter := rune(trimmed[index-1])
		// Micro sign and the greek mu are both used for microseconds, so both count as a unit letter.
		if unicode.IsLetter(letter) || letter == 'µ' {
			index--
			continue
		}
		break
	}

	return strings.TrimSpace(trimmed[:index]), trimmed[index:]
}

// durationUnits maps every accepted suffix onto its length in nanoseconds. The empty suffix means seconds,
// which is what a bare number reads as in a document written for humans.
var durationUnits = map[string]time.Duration{
	"":   time.Second,
	"ns": time.Nanosecond,
	"us": time.Microsecond,
	"µs": time.Microsecond,
	"ms": time.Millisecond,
	"s":  time.Second,
	"m":  time.Minute,
	"h":  time.Hour,
	"d":  24 * time.Hour,
}

// ParseDuration reads a value such as "7200 s", "30m" or "1.5h".
//
// Parsing is tolerant on the separator and strict on the unit: an unknown suffix is an error rather than a
// silent fallback, because a misread timeout is far harder to notice than a rejected document.
func ParseDuration(text string) (time.Duration, error) {
	number, unit := SplitUnit(text)
	if number == "" {
		return 0, fmt.Errorf("config: %q carries no number", text)
	}

	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %q is not a duration: %w", text, err)
	}

	scale, found := durationUnits[strings.ToLower(unit)]
	if !found {
		return 0, fmt.Errorf("config: %q carries an unknown duration unit", text)
	}
	if value < 0 {
		return 0, fmt.Errorf("config: %q is negative", text)
	}
	return time.Duration(value * float64(scale)), nil
}

// byteUnits maps every accepted suffix onto its size in bytes. A bare number reads as bytes.
var byteUnits = map[string]int64{
	"":    1,
	"b":   1,
	"kb":  1000,
	"mb":  1000 * 1000,
	"gb":  1000 * 1000 * 1000,
	"tb":  1000 * 1000 * 1000 * 1000,
	"kib": 1024,
	"mib": 1024 * 1024,
	"gib": 1024 * 1024 * 1024,
	"tib": 1024 * 1024 * 1024 * 1024,
}

// ParseBytes reads a value such as "4 mb", "512KiB" or "1048576".
func ParseBytes(text string) (int64, error) {
	number, unit := SplitUnit(text)
	if number == "" {
		return 0, fmt.Errorf("config: %q carries no number", text)
	}

	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, fmt.Errorf("config: %q is not a size: %w", text, err)
	}

	scale, found := byteUnits[strings.ToLower(unit)]
	if !found {
		return 0, fmt.Errorf("config: %q carries an unknown size unit", text)
	}
	if value < 0 {
		return 0, fmt.Errorf("config: %q is negative", text)
	}
	return int64(value * float64(scale)), nil
}

// ParseCount reads a plain count, rejecting anything that carries a unit. It exists so a caller does not have
// to import strconv just to read one number.
func ParseCount(text string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0, fmt.Errorf("config: %q is not a count: %w", text, err)
	}
	return value, nil
}
