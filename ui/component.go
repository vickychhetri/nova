package ui

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
)

// BuildContext provides contextual information while a Component builds or
// updates the UI tree.
//
// The context carries theme and scale information from the application and the
// mounted Node currently being reconciled. Components should treat it as
// lifecycle input for the current build rather than retain it indefinitely.
type BuildContext struct {
	// Theme is the active design-system configuration for this build.
	Theme *theme.Theme
	// Scale converts logical UI units to device pixels where needed.
	Scale float64
	// Node is the mounted runtime node associated with the current component.
	Node *Node
}

// Component is Nova's core declarative building block.
//
// Build returns the component description for the node's child subtree. Key
// identifies the component for reconciliation and state preservation. A nil
// result from Build represents a leaf or a component that manages children by
// another mechanism.
type Component interface {
	// Build constructs the component subtree for the current context.
	Build(ctx BuildContext) Component
	// Key returns the stable identity key used by reconciliation.
	Key() string
}

// RenderableComponent is implemented by components that directly calculate
// their layout and record drawing commands.
//
// It extends Component with the two renderer-facing phases. A Node delegates to
// these methods instead of automatically laying out or painting the component's
// children when the component implements this interface.
type RenderableComponent interface {
	Component
	// Layout measures the component under constraints and returns its size.
	Layout(node *Node, constraints layout.BoxConstraints) geom.Size
	// Paint records the component's visual commands into canvas.
	Paint(node *Node, canvas *render.Canvas)
}

// BaseComponent provides default Component implementations for components that
// only need a stable key or that manage their children through other APIs.
type BaseComponent struct {
	// CompKey is the component identity returned by Key.
	CompKey string
}

// Key returns the configured component identity key.
func (b BaseComponent) Key() string {
	return b.CompKey
}

// Build returns nil by default, identifying BaseComponent as a leaf or as a
// component whose child management is supplied by a specialized implementation.
func (b BaseComponent) Build(ctx BuildContext) Component {
	return nil
}
