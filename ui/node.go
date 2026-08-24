package ui

import (
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/event"
	"github.com/vickychhetri/nova/layout"
	"github.com/vickychhetri/nova/render"
)

// Node represents a mounted instance in Nova's retained UI tree.
//
// A Node connects a Component with its runtime relationships, layout bounds,
// interaction handlers, rendering callbacks, and cleanup functions. Component
// values describe what should exist; Nodes hold the persistent instance state
// used between builds and frames.
type Node struct {
	// Component is the component currently represented by this node.
	Component Component
	// Parent and Children define the retained tree relationships.
	Parent   *Node
	Children []*Node

	// Bounds is the node's position and size in its parent's coordinate space.
	Bounds geom.Rect
	// IsDirty indicates that the node or its dependent state needs updating.
	IsDirty bool
	// IsHovered and IsFocused store interaction state for this node.
	IsHovered bool
	IsFocused bool

	// Event handlers receive input routed to this node. Nil handlers are
	// treated as absent by dispatch code.
	OnClick        func()
	OnPointerDown  func(e *event.PointerEvent)
	OnPointerUp    func(e *event.PointerEvent)
	OnPointerMove  func(e *event.PointerEvent)
	OnPointerEnter func()
	OnPointerLeave func()
	OnScroll       func(e *event.ScrollEvent)
	OnKeyDown      func(e *event.KeyEvent)
	OnKeyUp        func(e *event.KeyEvent)
	// Overlay handlers support floating popups and dropdowns that are painted
	// above the ordinary component tree.
	PaintOverlay         func(canvas *render.Canvas)
	OnOverlayPointerDown func(p geom.Point) bool
	OnOverlayPointerMove func(p geom.Point) bool

	// Cleanup functions registered by state or effect integrations. They run
	// during Unmount and are then discarded.
	unsubscribes []func()
}

// NewNode creates an unmounted Node for comp.
//
// New nodes begin dirty so the first lifecycle pass can build, lay out, and
// paint them. Child capacity is reserved for the common case of a small tree.
func NewNode(comp Component) *Node {
	return &Node{
		Component: comp,
		IsDirty:   true,
		Children:  make([]*Node, 0, 4),
	}
}

// Mount mounts the node by reconciling its component and creating or updating
// the subtree represented by that component.
func (n *Node) Mount(ctx BuildContext) {
	n.Reconcile(ctx)
}

// Reconcile builds or updates the component subtree represented by n.
//
// The current reconciliation model supports a component producing either no
// child (a leaf or directly managed container) or one component child. Existing
// child nodes are reused so their identity, handlers, and state survive an
// update; only their Component value is replaced.
func (n *Node) Reconcile(ctx BuildContext) {
	if n.Component == nil {
		return
	}

	ctx.Node = n
	// Build produces the component description for the node's child subtree.
	built := n.Component.Build(ctx)
	if built == nil {
		// Leaf node, or a container that manages children directly (for example,
		// a Flex component), so reconciliation does not create a child here.
		return
	}

	if len(n.Children) == 0 {
		// First build: create the child node and establish the parent link.
		childNode := NewNode(built)
		childNode.Parent = n
		n.Children = []*Node{childNode}
		childNode.Mount(ctx)
	} else {
		// Update the existing single child instead of replacing its identity.
		n.Children[0].Component = built
		n.Children[0].Reconcile(ctx)
	}
}

// Layout performs the layout pass for n and returns its resulting size.
//
// RenderableComponent implementations own their layout calculation. A
// non-renderable wrapper delegates to its first child, while a component with
// no child accepts the smallest size allowed by the supplied constraints.
func (n *Node) Layout(constraints layout.BoxConstraints) geom.Size {
	if n.Component == nil {
		n.Bounds = geom.NewRect(0, 0, 0, 0)
		return geom.Sz(0, 0)
	}

	if renderable, ok := n.Component.(RenderableComponent); ok {
		// Renderable components can measure themselves and update their own
		// bounds using the incoming constraints.
		sz := renderable.Layout(n, constraints)
		n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
		return sz
	}

	if len(n.Children) > 0 {
		// Wrapper nodes inherit the first child's measured size and place that
		// child at the wrapper's local origin.
		sz := n.Children[0].Layout(constraints)
		n.Children[0].Bounds = geom.NewRect(0, 0, sz.Width, sz.Height)
		n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
		return sz
	}

	// A component with no renderable implementation and no child has no
	// intrinsic content, so let constraints determine its minimum result.
	sz := constraints.Constrain(geom.Sz(0, 0))
	n.Bounds = geom.NewRect(n.Bounds.X, n.Bounds.Y, sz.Width, sz.Height)
	return sz
}

// Paint records rendering commands for n and its descendants.
//
// Each node saves the canvas state and translates by its local Bounds origin.
// As a result, a component can paint using local coordinates while nested
// children accumulate their positions through the canvas transform stack.
func (n *Node) Paint(canvas *render.Canvas) {
	if n.Component == nil {
		return
	}

	canvas.Save()
	canvas.Translate(n.Bounds.X, n.Bounds.Y)

	if renderable, ok := n.Component.(RenderableComponent); ok {
		// Renderable nodes paint their own content; their children are not
		// traversed here because the component owns that rendering decision.
		renderable.Paint(n, canvas)
	} else {
		// Non-renderable nodes act as wrappers and paint each child in order.
		for _, child := range n.Children {
			child.Paint(canvas)
		}
	}

	canvas.Restore()
}

// HitTestLocal traverses the node hierarchy and returns the deepest interactive
// node containing p, together with p expressed in that node's local space.
//
// Children are visited in reverse order so the last painted child is treated as
// topmost. This matches painter-order hit testing when later children visually
// cover earlier children.
func (n *Node) HitTestLocal(p geom.Point) (*Node, geom.Point) {
	if !n.Bounds.ContainsPoint(p) {
		return nil, geom.Point{}
	}

	localPoint := p.Sub(n.Bounds.Origin())

	// Check children in reverse order (topmost paint order first).
	for i := len(n.Children) - 1; i >= 0; i-- {
		child := n.Children[i]
		// Keep the relationship available for GlobalToLocal even when a child
		// was attached by a container-specific path.
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

// HitTest returns the deepest interactive node under p, discarding the local
// coordinate returned by HitTestLocal.
func (n *Node) HitTest(p geom.Point) *Node {
	hit, _ := n.HitTestLocal(p)
	return hit
}

// GlobalToLocal converts a point from the current outer coordinate space into
// this node's local coordinate space by subtracting each ancestor origin.
//
// The method follows Parent links until the root. The caller must provide a
// point in the same coordinate space used by the root bounds.
func (n *Node) GlobalToLocal(p geom.Point) geom.Point {
	cur := n
	for cur != nil {
		p = p.Sub(cur.Bounds.Origin())
		cur = cur.Parent
	}
	return p
}

// IsInteractive reports whether n has any direct interaction handler that can
// make it a hit-test target.
//
// Hover enter/leave callbacks are intentionally not included because they are
// state transitions managed by the pointer-dispatch logic rather than direct
// target handlers in this predicate.
func (n *Node) IsInteractive() bool {
	return n.OnClick != nil ||
		n.OnPointerDown != nil ||
		n.OnPointerUp != nil ||
		n.OnPointerMove != nil ||
		n.OnScroll != nil ||
		n.OnKeyDown != nil ||
		n.OnKeyUp != nil
}

// PaintOverlays recursively records overlay layers after the base tree.
//
// Overlay callbacks are used for floating content such as popups and dropdowns;
// they are painted in traversal order on top of ordinary node content.
func (n *Node) PaintOverlays(canvas *render.Canvas) {
	if n.PaintOverlay != nil {
		n.PaintOverlay(canvas)
	}
	for _, child := range n.Children {
		child.PaintOverlays(canvas)
	}
}

// DispatchOverlayPointerDown checks overlays recursively for a pointer-down
// handler and returns true as soon as one reports that it consumed the point.
func (n *Node) DispatchOverlayPointerDown(p geom.Point) bool {
	if n.OnOverlayPointerDown != nil {
		if n.OnOverlayPointerDown(p) {
			return true
		}
	}
	for _, child := range n.Children {
		// Stop at the first consuming overlay so only one popup handles the
		// pointer-down action.
		if child.DispatchOverlayPointerDown(p) {
			return true
		}
	}
	return false
}

// DispatchOverlayPointerMove checks overlays recursively for a pointer-move
// handler and returns true when one consumes the movement.
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

// Unmount disposes resources registered by n and recursively unmounts children.
//
// Local cleanup functions run before child cleanup. After execution, the local
// cleanup list and child slice are cleared so the node no longer retains those
// resources or relationships.
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
