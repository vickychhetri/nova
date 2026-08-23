package ui

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
)

// Node represents a mounted instance in the retained UI tree.
type Node struct {
	Component Component
	Parent    *Node
	Children  []*Node

	Bounds   geom.Rect
	IsDirty  bool
	IsHovered bool
	IsFocused bool

	// Event handlers
	OnClick        func()
	OnPointerDown  func(e *event.PointerEvent)
	OnPointerUp    func(e *event.PointerEvent)
	OnPointerMove  func(e *event.PointerEvent)
	OnPointerEnter func()
	OnPointerLeave func()
	OnScroll       func(e *event.ScrollEvent)
	OnKeyDown      func(e *event.KeyEvent)
	OnKeyUp        func(e *event.KeyEvent)
	// Overlay handlers for floating popups and dropdowns
	PaintOverlay         func(canvas *render.Canvas)
	OnOverlayPointerDown func(p geom.Point) bool
	OnOverlayPointerMove func(p geom.Point) bool

	// Cleanups registered by state/effects
	unsubscribes []func()
}

// NewNode creates a new unmounted Node for a component.
func NewNode(comp Component) *Node {
	return &Node{
		Component: comp,
		IsDirty:   true,
		Children:  make([]*Node, 0, 4),
	}
}

// Mount mounts the node and builds its subtree.
func (n *Node) Mount(ctx BuildContext) {
	n.Reconcile(ctx)
}

// Reconcile builds or updates the component subtree.
func (n *Node) Reconcile(ctx BuildContext) {
	if n.Component == nil {
		return
	}

	ctx.Node = n
	built := n.Component.Build(ctx)
	if built == nil {
		// Leaf node, children may be managed directly (e.g. Flex container)
		return
	}

	if len(n.Children) == 0 {
		childNode := NewNode(built)
		childNode.Parent = n
		n.Children = []*Node{childNode}
		childNode.Mount(ctx)
	} else {
		// Single child update
		n.Children[0].Component = built
		n.Children[0].Reconcile(ctx)
	}
}

// Layout performs the layout pass for this node and its children.
func (n *Node) Layout(constraints layout.BoxConstraints) geom.Size {
	if n.Component == nil {
		n.Bounds = geom.NewRect(0, 0, 0, 0)
		return geom.Sz(0, 0)
	}

	if renderable, ok := n.Component.(RenderableComponent); ok {
		sz := renderable.Layout(n, constraints)
		n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
		return sz
	}

	if len(n.Children) > 0 {
		sz := n.Children[0].Layout(constraints)
		n.Children[0].Bounds = geom.NewRect(0, 0, sz.Width, sz.Height)
		n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
		return sz
	}

	sz := constraints.Constrain(geom.Sz(0, 0))
	n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
	return sz
}

// Paint records rendering commands for this node and its children.
func (n *Node) Paint(canvas *render.Canvas) {
	if n.Component == nil {
		return
	}

	canvas.Save()
	canvas.Translate(n.Bounds.X, n.Bounds.Y)

	if renderable, ok := n.Component.(RenderableComponent); ok {
		renderable.Paint(n, canvas)
	} else {
		for _, child := range n.Children {
			child.Paint(canvas)
		}
	}

	canvas.Restore()
}

// HitTestLocal traverses node hierarchy and returns deepest interactive node and point in local coordinates.
func (n *Node) HitTestLocal(p geom.Point) (*Node, geom.Point) {
	if !n.Bounds.ContainsPoint(p) {
		return nil, geom.Point{}
	}

	localPoint := p.Sub(n.Bounds.Origin())

	// Check children in reverse order (top-most z-index first)
	for i := len(n.Children) - 1; i >= 0; i-- {
		child := n.Children[i]
		child.Parent = n
		hit, lp := child.HitTestLocal(localPoint)
		if hit != nil {
			return hit, lp
		}
	}

	if n.IsInteractive() {
		return n, localPoint
	}

	return nil, geom.Point{}
}

// HitTest traverses node hierarchy and returns deepest interactive node under point p.
func (n *Node) HitTest(p geom.Point) *Node {
	hit, _ := n.HitTestLocal(p)
	return hit
}

// GlobalToLocal transforms a global point into local node coordinate space.
func (n *Node) GlobalToLocal(p geom.Point) geom.Point {
	cur := n
	for cur != nil {
		p = p.Sub(cur.Bounds.Origin())
		cur = cur.Parent
	}
	return p
}

// IsInteractive returns true if node has interactive event handlers.
func (n *Node) IsInteractive() bool {
	return n.OnClick != nil ||
		n.OnPointerDown != nil ||
		n.OnPointerUp != nil ||
		n.OnPointerMove != nil ||
		n.OnScroll != nil ||
		n.OnKeyDown != nil ||
		n.OnKeyUp != nil
}

// PaintOverlays recursively paints overlay layers on top of the base UI tree.
func (n *Node) PaintOverlays(canvas *render.Canvas) {
	if n.PaintOverlay != nil {
		n.PaintOverlay(canvas)
	}
	for _, child := range n.Children {
		child.PaintOverlays(canvas)
	}
}

// DispatchOverlayPointerDown recursively checks if any active overlay handles pointer down.
func (n *Node) DispatchOverlayPointerDown(p geom.Point) bool {
	if n.OnOverlayPointerDown != nil {
		if n.OnOverlayPointerDown(p) {
			return true
		}
	}
	for _, child := range n.Children {
		if child.DispatchOverlayPointerDown(p) {
			return true
		}
	}
	return false
}

// DispatchOverlayPointerMove recursively checks if any active overlay handles pointer move.
func (n *Node) DispatchOverlayPointerMove(p geom.Point) bool {
	if n.OnOverlayPointerMove != nil {
		if n.OnOverlayPointerMove(p) {
			return true
		}
	}
	for _, child := range n.Children {
		if child.DispatchOverlayPointerMove(p) {
			return true
		}
	}
	return false
}

// Unmount disposes resources and children.
func (n *Node) Unmount() {
	for _, unsub := range n.unsubscribes {
		unsub()
	}
	n.unsubscribes = nil

	for _, child := range n.Children {
		child.Unmount()
	}
	n.Children = nil
}
