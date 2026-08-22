package ui

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/theme"
)

// BuildContext provides contextual information during component rendering and tree building.
type BuildContext struct {
	Theme *theme.Theme
	Scale float64
	Node  *Node
}

// Component is the core declarative building block in Nova.
type Component interface {
	Build(ctx BuildContext) Component
	Key() string
}

// RenderableComponent is implemented by primitive leaf components that directly perform layout and drawing.
type RenderableComponent interface {
	Component
	Layout(node *Node, constraints layout.BoxConstraints) geom.Size
	Paint(node *Node, canvas *render.Canvas)
}

// BaseComponent provides default implementations for Component.
type BaseComponent struct {
	CompKey string
}

func (b BaseComponent) Key() string {
	return b.CompKey
}

func (b BaseComponent) Build(ctx BuildContext) Component {
	return nil
}
