package ui

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// Option is a labelled value offered by Select.
type Option[T any] struct {
	Label string
	Value T
}

// Select shows a type-to-filter list and returns the chosen value.
func Select[T any](message string, options []Option[T]) (T, error) {
	var zero T
	if len(options) == 0 {
		return zero, errors.New("nothing to select from")
	}
	labels := make([]string, len(options))
	for i, o := range options {
		labels[i] = o.Label
	}
	prompt := &survey.Select{
		Message:  message,
		Options:  labels,
		PageSize: 15,
		Filter: func(filter, value string, _ int) bool {
			return FuzzyMatch(filter, value)
		},
	}
	var answer survey.OptionAnswer
	if err := survey.AskOne(prompt, &answer); err != nil {
		return zero, err
	}
	return options[answer.Index].Value, nil
}

// Input asks for a free-text value with a default.
func Input(message, defaultValue string) (string, error) {
	var answer string
	prompt := &survey.Input{Message: message, Default: defaultValue}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}
	return answer, nil
}

// InputInt asks for a non-negative integer with a default.
func InputInt(message string, defaultValue int) (int, error) {
	var answer string
	prompt := &survey.Input{Message: message, Default: strconv.Itoa(defaultValue)}
	validator := func(v any) error {
		s, _ := v.(string)
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 0 {
			return errors.New("please enter a whole number of 0 or more")
		}
		return nil
	}
	if err := survey.AskOne(prompt, &answer, survey.WithValidator(validator)); err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(answer))
}

// Confirm asks a yes/no question.
func Confirm(message string, defaultValue bool) (bool, error) {
	var answer bool
	prompt := &survey.Confirm{Message: message, Default: defaultValue}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return false, err
	}
	return answer, nil
}

// IsInterrupt reports whether err came from the user pressing ctrl-c at a prompt.
func IsInterrupt(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

// FuzzyMatch reports whether every rune of pattern appears in s, in order
// (case-insensitive), mirroring the matching used by the original tool.
func FuzzyMatch(pattern, s string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}
	want := []rune(strings.ToLower(pattern))
	i := 0
	for _, r := range strings.ToLower(s) {
		if i < len(want) && unicode.ToLower(r) == want[i] {
			i++
		}
	}
	return i == len(want)
}
