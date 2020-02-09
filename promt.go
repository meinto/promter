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
	Select(label string, options []string) (index int, selection string, err error)
	Text(label string, options ...PromterOptions) (input string, err error)
	OptionalText(label string, options ...PromterOptions) (input string, err error)
	URLWithDefault(label, defaultValue string, options ...PromterOptions) (url string, err error)
	URL(label string, options ...PromterOptions) (url string, err error)
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

func (p *promter) YesNo(label string) (index int, selection string, err error) {
	return p.Select(label, []string{"Yes", "No"})
}

func (p *promter) Select(label string, options []string) (index int, selection string, err error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}
	index, _, err = prompt.Run()
	return index, options[index], err
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

	notEmpty, err := notEmptyText.Run()
	if err != nil {
		return "", err
	}

	return strings.ToLower(notEmpty), nil
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

func (p *promter) URLWithDefault(label, defaultValue string, options ...PromterOptions) (url string, err error) {
	defer func() {
		p.HandleRetries(err, func() {
			p.URLWithDefault(label, defaultValue, options...)
		}, options...)
	}()

	url, err = p.URL(fmt.Sprintf(label, defaultValue), PromterOptions{false})
	if err != nil {
		return url, err
	}

	if url == "" {
		url = defaultValue
	}

	return strings.ToLower(url), nil
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
