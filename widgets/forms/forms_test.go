package forms_test

import (
	"testing"

	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets/forms"
)

func TestFormValidation(t *testing.T) {
	form := forms.NewFormState()
	form.RegisterField("username", forms.Required(), forms.MinLength(3))
	form.RegisterField("email", forms.Required(), forms.Email())

	if form.ValidateAll() {
		t.Fatal("expected empty form to be invalid")
	}

	if len(form.Errors()) != 2 {
		t.Fatalf("expected 2 errors, got %d: %+v", len(form.Errors()), form.Errors())
	}

	// Set invalid username
	form.Set("username", "ab")
	if form.ValidateField("username") {
		t.Fatal("expected username 'ab' to fail min length 3")
	}

	// Set valid username
	form.Set("username", "john_doe")
	if !form.ValidateField("username") {
		t.Fatal("expected 'john_doe' to pass validation")
	}

	// Set invalid email
	form.Set("email", "not-an-email")
	if form.ValidateField("email") {
		t.Fatal("expected 'not-an-email' to fail email validation")
	}

	// Set valid email
	form.Set("email", "john@example.com")
	if !form.ValidateField("email") {
		t.Fatal("expected 'john@example.com' to pass email validation")
	}

	if !form.ValidateAll() || !form.IsValid() {
		t.Fatalf("expected form to be valid now, errors: %+v", form.Errors())
	}

	var submittedVals map[string]any
	form.OnSubmit = func(vals map[string]any) {
		submittedVals = vals
	}

	if !form.Submit() || submittedVals == nil {
		t.Fatal("expected submit to succeed")
	}
	if submittedVals["username"] != "john_doe" || submittedVals["email"] != "john@example.com" {
		t.Fatalf("unexpected submitted values: %+v", submittedVals)
	}
}

func TestFormWidgetsRendering(t *testing.T) {
	emailState := state.String("test@example.com")
	agreeState := state.Bool(true)
	genderState := state.String("female")
	volumeState := state.Float(75)

	formComponent := forms.Form(
		forms.NewFormState(),
		forms.TextField("Enter email").WithLabel("Email").Bind(emailState),
		forms.PasswordField("Enter password").WithLabel("Password"),
		forms.TextArea("Bio...").WithLabel("Bio"),
		forms.NumberInput(10),
		forms.Checkbox("Agree to Terms").Bind(agreeState),
		forms.Radio("Female", "female", genderState),
		forms.Switch("Enable notifications"),
		forms.Slider(0, 100).Bind(volumeState),
		forms.Select(
			forms.SelectOption{Label: "Admin", Value: "admin"},
			forms.SelectOption{Label: "User", Value: "user"},
		),
		forms.DatePicker(),
		forms.FilePicker("Choose CSV file..."),
	)

	node := ui.NewNode(formComponent)
	node.Mount(ui.BuildContext{
		Theme: theme.Dark(),
		Scale: 1.0,
	})

	sz := node.Layout(layout.TightWidth(400))
	if sz.Width != 400 || sz.Height <= 0 {
		t.Fatalf("unexpected form layout size: %s", sz)
	}

	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)
	node.Paint(canvas)

	if buf.Len() == 0 {
		t.Fatal("expected paint commands for form widgets")
	}
}
