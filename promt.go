package promter

import (
	"fmt"
	"strings"

	"github.com/imdario/mergo"
	"github.com/manifoldco/promptui"
)

type PromterOptions struct {
	handleRetries bool
}

var prompterDefaultOptions = PromterOptions{true}

func MergeOptions(options ...PromterOptions) PromterOptions {
	o := prompterDefaultOptions
	for _, option := range options {
		mergo.Merge(&o, option, mergo.WithOverride)
	}
	return o
}

type Promter interface {
	YesNo(label string) (index int, selection string, err error)
	YesNoDefault(label, defaultValue string) (index int, selection string, err error)
	Select(label string, options []string) (index int, selection string, err error)
	SelectDefault(label, defaultValue string, options []string) (index int, selection string, err error)
	Text(label string, options ...PromterOptions) (input string, err error)
	TextDefault(label, defaultValue string, options ...PromterOptions) (input string, err error)
	OptionalText(label string, options ...PromterOptions) (input string, err error)
	OptionalTextDefault(label, defaultValue string, options ...PromterOptions) (input string, err error)
	URL(label string, options ...PromterOptions) (url string, err error)
	URLDefault(label, defaultValue string, options ...PromterOptions) (url string, err error)
}

type promter struct {
	retrys int
}

func NewPromter() Promter {
	return &promter{0}
}

func (p *promter) HandleRetries(err error, cb func(), options ...PromterOptions) {
	o := MergeOptions(options...)
	if o.handleRetries {
		if err == nil {
			p.retrys = 0
		}
		p.retrys++
		if err != nil && p.retrys <= 3 {
			cb()
		}
	}
}

func LabelWithDefault(label, defaultValue string) string {
	str := label
	if strings.TrimSpace(defaultValue) != "" {
		str = fmt.Sprintf("%s (default: %s)", label, defaultValue)
	}
	return str
}

func (p *promter) YesNo(label string) (index int, selection string, err error) {
	return p.Select(label, []string{"Yes", "No"})
}

func (p *promter) YesNoDefault(label, defaultValue string) (index int, selection string, err error) {
	return p.Select(LabelWithDefault(label, defaultValue), []string{"Yes", "No"})
}

func (p *promter) Select(label string, options []string) (index int, selection string, err error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}
	index, _, err = prompt.Run()
	return index, options[index], err
}

func (p *promter) SelectDefault(label, defaultValue string, options []string) (index int, selection string, err error) {
	return p.Select(LabelWithDefault(label, defaultValue), options)
}

func (p *promter) Text(label string, options ...PromterOptions) (input string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.Text(label, options...)
		}, options...)
	}()

	validate := func(input string) error {
		if strings.TrimSpace(input) == "" {
			return fmt.Errorf("please provide a text")
		}
		return nil
	}

	notEmptyText := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	input, err = notEmptyText.Run()
	if err != nil {
		return "", err
	}

	return input, nil
}

func (p *promter) TextDefault(label, defaultValue string, options ...PromterOptions) (input string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.TextDefault(label, defaultValue, options...)
		}, options...)
	}()

	validate := func(input string) error {
		if strings.TrimSpace(input) == "" && strings.TrimSpace(defaultValue) == "" {
			return fmt.Errorf("please provide a text")
		}
		return nil
	}

	notEmptyText := promptui.Prompt{
		Label:    LabelWithDefault(label, defaultValue),
		Validate: validate,
	}

	input, err = notEmptyText.Run()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(input) == "" {
		input = defaultValue
	}

	return input, nil
}

func (p *promter) OptionalText(label string, options ...PromterOptions) (input string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.OptionalText(label, options...)
		}, options...)
	}()

	getValue := promptui.Prompt{
		Label: label,
	}
	value, err := getValue.Run()
	return value, err
}

func (p *promter) OptionalTextDefault(label, defaultValue string, options ...PromterOptions) (input string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.OptionalTextDefault(label, defaultValue, options...)
		}, options...)
	}()

	input, err = p.OptionalText(LabelWithDefault(label, defaultValue), PromterOptions{false})

	if strings.TrimSpace(input) == "" {
		input = defaultValue
	}
	return input, err
}

func (p *promter) URL(label string, options ...PromterOptions) (url string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.URL(label, options...)
		}, options...)
	}()

	validate := func(input string) error {
		if input != "" &&
			!strings.HasPrefix(input, "http://") &&
			!strings.HasPrefix(input, "https://") {
			return fmt.Errorf("please enter a valid url")
		}
		return nil
	}

	getURL := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	url, err = getURL.Run()
	if err != nil {
		return "", err
	}

	return strings.ToLower(url), nil
}

func (p *promter) URLDefault(label, defaultValue string, options ...PromterOptions) (url string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.URLDefault(label, defaultValue, options...)
		}, options...)
	}()

	url, err = p.URL(LabelWithDefault(label, defaultValue), PromterOptions{false})
	if err != nil {
		return url, err
	}

	if url == "" {
		url = defaultValue
	}

	return strings.ToLower(url), nil
}
