package forms

// Package forms provides form state, validation rules, and reusable input
// controls.

import (
	"regexp"
	"strings"

	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/ui"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidationRule is a function that checks a field value and returns an error message if invalid.
type ValidationRule func(value any) string

// Common validation rule helpers
func Required(message ...string) ValidationRule {
	msg := "This field is required"
	if len(message) > 0 {
		msg = message[0]
	}
	return func(val any) string {
		if val == nil {
			return msg
		}
		if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
			return msg
		}
		return ""
	}
}

func MinLength(min int, message ...string) ValidationRule {
	return func(val any) string {
		if s, ok := val.(string); ok && len(s) < min {
			if len(message) > 0 {
				return message[0]
			}
			return "Minimum length is not met"
		}
		return ""
	}
}

func MaxLength(max int, message ...string) ValidationRule {
	return func(val any) string {
		if s, ok := val.(string); ok && len(s) > max {
			if len(message) > 0 {
				return message[0]
			}
			return "Maximum length exceeded"
		}
		return ""
	}
}

func Email(message ...string) ValidationRule {
	msg := "Please enter a valid email address"
	if len(message) > 0 {
		msg = message[0]
	}
	return func(val any) string {
		if s, ok := val.(string); ok && s != "" {
			if !emailRegex.MatchString(s) {
				return msg
			}
		}
		return ""
	}
}

// FormState manages validation, values, dirty state, and submission.
type FormState struct {
	values   map[string]any
	errors   map[string]string
	rules    map[string][]ValidationRule
	isDirty  bool
	OnSubmit func(values map[string]any)
}

// NewFormState creates a FormState.
func NewFormState() *FormState {
	return &FormState{
		values: make(map[string]any),
		errors: make(map[string]string),
		rules:  make(map[string][]ValidationRule),
	}
}

// RegisterField registers validation rules for a named field.
func (f *FormState) RegisterField(name string, rules ...ValidationRule) {
	f.rules[name] = rules
}

// Set sets a field's value and validates it.
func (f *FormState) Set(name string, value any) {
	f.values[name] = value
	f.isDirty = true
	f.ValidateField(name)
}

// Get returns current value for a field.
func (f *FormState) Get(name string) any {
	return f.values[name]
}

// ValidateField runs validation rules for a single field.
func (f *FormState) ValidateField(name string) bool {
	rules := f.rules[name]
	val := f.values[name]
	for _, rule := range rules {
		if errMsg := rule(val); errMsg != "" {
			f.errors[name] = errMsg
			return false
		}
	}
	delete(f.errors, name)
	return true
}

// ValidateAll validates all registered fields and returns true if form is valid.
func (f *FormState) ValidateAll() bool {
	valid := true
	for name := range f.rules {
		if !f.ValidateField(name) {
			valid = false
		}
	}
	return valid
}

// IsValid returns true if there are no validation errors.
func (f *FormState) IsValid() bool {
	return len(f.errors) == 0
}

// IsDirty returns true if any field has been modified.
func (f *FormState) IsDirty() bool {
	return f.isDirty
}

// Errors returns map of field errors.
func (f *FormState) Errors() map[string]string {
	return f.errors
}

// Submit validates the form and invokes OnSubmit callback if valid.
func (f *FormState) Submit() bool {
	if f.ValidateAll() {
		if f.OnSubmit != nil {
			f.OnSubmit(f.values)
		}
		return true
	}
	return false
}

// Reset clears all values, errors, and dirty state.
func (f *FormState) Reset() {
	f.values = make(map[string]any)
	f.errors = make(map[string]string)
	f.isDirty = false
}

// FormComponent wraps children in a form layout.
type FormComponent struct {
	ui.BaseComponent
	State    *FormState
	Children []ui.Component
}

// Form creates a Form component container.
func Form(state *FormState, children ...ui.Component) *FormComponent {
	return &FormComponent{
		State:    state,
		Children: children,
	}
}

func (fc *FormComponent) Layout(node *ui.Node, constraints layout.BoxConstraints) geom.Size {
	flex := ui.Column(fc.Children...).GapSpacing(12)
	return flex.Layout(node, constraints)
}

func (fc *FormComponent) Paint(node *ui.Node, canvas *render.Canvas) {
	for _, child := range node.Children {
		child.Paint(canvas)
	}
}

// State helpers for forms
type FormFieldBinding[T any] struct {
	Name  string
	State *state.Value[T]
	Rules []ValidationRule
}
