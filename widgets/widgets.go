package widgets

// Package widgets provides the public, ergonomic widget facade for Nova.
//
// It re-exports common UI builders and specialized widgets from subpackages so
// applications can compose controls through one import path.

import (
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets/basic"
	"github.com/vickychhetri/nova/widgets/data"
	"github.com/vickychhetri/nova/widgets/editor"
	"github.com/vickychhetri/nova/widgets/feedback"
	"github.com/vickychhetri/nova/widgets/forms"
	"github.com/vickychhetri/nova/widgets/nav"
)

// --- Basic UI Primitives ---
// These aliases expose geometry and core UI builders alongside common widgets.
var (
	All         = ui.All
	Symmetric   = ui.Symmetric
	TRBL        = ui.TRBL
	Pt          = ui.Pt
	Sz          = ui.Sz
	Button      = ui.Button
	Text        = ui.Text
	Column      = ui.Column
	Row         = ui.Row
	Stack       = ui.Stack
	Container   = ui.Container
	Center      = ui.Center
	Padding     = ui.Padding
	Spacer      = ui.Spacer
	Badge       = basic.Badge
	Avatar      = basic.Avatar
	Spinner     = basic.Spinner
	Progress    = basic.Progress
	Card        = basic.Card
	GroupBox    = basic.GroupBox
	Skeleton    = basic.Skeleton
	FormatCount = basic.FormatCount
)

// --- Form Controls & Inputs ---
// These aliases expose forms, validation helpers, and input controls.
var (
	Form          = forms.Form
	NewFormState  = forms.NewFormState
	Required      = forms.Required
	MinLength     = forms.MinLength
	MaxLength     = forms.MaxLength
	Email         = forms.Email
	TextField     = forms.TextField
	PasswordField = forms.PasswordField
	TextArea      = forms.TextArea
	NumberInput   = forms.NumberInput
	Checkbox      = forms.Checkbox
	Radio         = forms.Radio
	Switch        = forms.Switch
	Slider        = forms.Slider
	Select        = forms.Select
	DatePicker    = forms.DatePicker
	ColorPicker   = forms.ColorPicker
	FilePicker    = forms.FilePicker
)

// SelectOption re-exports the form select option type.
type SelectOption = forms.SelectOption

// --- Navigation & Structure ---
// These aliases expose navigation containers and menu model types.
var (
	Tabs          = nav.Tabs
	Sidebar       = nav.Sidebar
	SplitPane     = nav.SplitPane
	Breadcrumb    = nav.Breadcrumb
	MenuBar       = nav.MenuBar
	MenuBarSimple = nav.MenuBarFromItems
	NewMenu       = nav.NewMenu
	SimpleMenu    = nav.SimpleMenu
	ActionItem    = nav.ActionItem
	ShortcutItem  = nav.ShortcutItem
	DividerItem   = nav.DividerItem
	Toolbar       = nav.Toolbar
	StatusBar     = nav.StatusBar
)

type TabItem = nav.TabItem
type SidebarItem = nav.SidebarItem
type BreadcrumbItem = nav.BreadcrumbItem
type MenuBarItem = nav.MenuBarItem
type Menu = nav.Menu
type MenuItem = nav.MenuItem
type StatusSegment = nav.StatusSegment

const (
	SplitHorizontal = nav.SplitHorizontal
	SplitVertical   = nav.SplitVertical
)

// --- Feedback & Overlays ---
// These aliases expose status, dialog, toast, and command-palette widgets.
var (
	Alert           = feedback.Alert
	Dialog          = feedback.Dialog
	CommandPalette  = feedback.CommandPalette
	NewToastManager = feedback.NewToastManager
)

const (
	AlertInfo    = feedback.AlertInfo
	AlertSuccess = feedback.AlertSuccess
	AlertWarning = feedback.AlertWarning
	AlertError   = feedback.AlertError
)

type CommandItem = feedback.CommandItem

// --- Data Displays ---
// These aliases expose virtualized and hierarchical data widgets.
var (
	VirtualList = data.VirtualList
	Table       = data.Table
	Tree        = data.Tree
)

type TableColumn = data.TableColumn
type TreeNode = data.TreeNode

// --- Specialized Editors ---
// These aliases expose editor-oriented components.
var (
	CodeEditor = editor.CodeEditor
	Canvas     = editor.Canvas
)

type CustomDrawFunc = editor.CustomDrawFunc

// Unused import suppression
var (
	_ = color.White
	_ = state.Int
	_ = ui.Text
)
