package geom

import (
	"fmt"
	"math"
)

// Point represents a 2D coordinate in a Cartesian 2D space.
type Point struct {
	X float64
	Y float64
}

// Pt creates a Point with the supplied x and y coordinates.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Add returns the component-wise sum of p and other.
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub returns the component-wise difference p - other.
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// Scale multiplies both coordinates by s relative to the origin.
func (p Point) Scale(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Distance returns the Euclidean distance between p and other.
func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Hypot(dx, dy)
}

func (p Point) String() string {
	// Keep debug output compact and stable to one decimal place.
	return fmt.Sprintf("Point(%.1f, %.1f)", p.X, p.Y)
}

// Size represents non-negative two-dimensional dimensions.
type Size struct {
	Width  float64
	Height float64
}

// Sz creates a Size and clamps negative dimensions to zero.
func Sz(width, height float64) Size {
	return Size{Width: math.Max(0, width), Height: math.Max(0, height)}
}

// IsZero reports whether both dimensions are non-positive.
func (s Size) IsZero() bool {
	return s.Width <= 0 && s.Height <= 0
}

// IsInfinite reports whether either dimension is positive infinity.
func (s Size) IsInfinite() bool {
	return math.IsInf(s.Width, 1) || math.IsInf(s.Height, 1)
}

// Clamp restricts each dimension independently to the inclusive [min, max]
// range. It assumes the corresponding min dimension is not greater than max.
func (s Size) Clamp(min, max Size) Size {
	return Size{
		Width:  math.Max(min.Width, math.Min(max.Width, s.Width)),
		Height: math.Max(min.Height, math.Min(max.Height, s.Height)),
	}
}

func (s Size) String() string {
	return fmt.Sprintf("Size(%.1f x %.1f)", s.Width, s.Height)
}

// Rect represents an axis-aligned rectangle using its top-left origin and
// non-negative Width and Height.
//
// Its right and bottom edges are derived as X+Width and Y+Height. Geometry
// operations in this package use those derived edges rather than storing a
// separate maximum coordinate.
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewRect creates a rectangle and clamps negative width or height to zero.
func NewRect(x, y, width, height float64) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  math.Max(0, width),
		Height: math.Max(0, height),
	}
}

// RectFromPoints creates the smallest axis-aligned rectangle spanning p1 and
// p2, regardless of the order of the two points.
func RectFromPoints(p1, p2 Point) Rect {
	minX := math.Min(p1.X, p2.X)
	minY := math.Min(p1.Y, p2.Y)
	maxX := math.Max(p1.X, p2.X)
	maxY := math.Max(p1.Y, p2.Y)
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// Origin returns the rectangle's top-left corner.
func (r Rect) Origin() Point {
	return Point{X: r.X, Y: r.Y}
}

// Size returns the rectangle dimensions as a Size value.
func (r Rect) Size() Size {
	return Size{Width: r.Width, Height: r.Height}
}

// MinX returns the left edge coordinate.
func (r Rect) MinX() float64 { return r.X }

// MaxX returns the right edge coordinate, X+Width.
func (r Rect) MaxX() float64 { return r.X + r.Width }

// MinY returns the top edge coordinate.
func (r Rect) MinY() float64 { return r.Y }

// MaxY returns the bottom edge coordinate, Y+Height.
func (r Rect) MaxY() float64 { return r.Y + r.Height }

// Center returns the midpoint between the rectangle's opposite corners.
func (r Rect) Center() Point {
	return Point{
		X: r.X + r.Width/2,
		Y: r.Y + r.Height/2,
	}
}

// ContainsPoint reports whether p lies inside or on the rectangle boundary.
func (r Rect) ContainsPoint(p Point) bool {
	return p.X >= r.MinX() && p.X <= r.MaxX() && p.Y >= r.MinY() && p.Y <= r.MaxY()
}

// ContainsRect reports whether other is fully enclosed, including coincident
// edges, by r.
func (r Rect) ContainsRect(other Rect) bool {
	return other.MinX() >= r.MinX() && other.MaxX() <= r.MaxX() &&
		other.MinY() >= r.MinY() && other.MaxY() <= r.MaxY()
}

// Intersects reports whether r and other overlap with positive area. Merely
// touching at an edge or corner does not count as an intersection.
func (r Rect) Intersects(other Rect) bool {
	return r.MinX() < other.MaxX() && r.MaxX() > other.MinX() &&
		r.MinY() < other.MaxY() && r.MaxY() > other.MinY()
}

// Intersection returns the positive-area overlap of r and other, or a zero
// rectangle when they do not overlap.
func (r Rect) Intersection(other Rect) Rect {
	minX := math.Max(r.MinX(), other.MinX())
	minY := math.Max(r.MinY(), other.MinY())
	maxX := math.Min(r.MaxX(), other.MaxX())
	maxY := math.Min(r.MaxY(), other.MaxY())

	if maxX <= minX || maxY <= minY {
		return Rect{}
	}
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// Union returns the smallest axis-aligned rectangle containing both inputs.
// A zero-width and zero-height rectangle is treated as an empty sentinel and
// does not expand the other rectangle.
func (r Rect) Union(other Rect) Rect {
	if r.Width == 0 && r.Height == 0 {
		return other
	}
	if other.Width == 0 && other.Height == 0 {
		return r
	}
	minX := math.Min(r.MinX(), other.MinX())
	minY := math.Min(r.MinY(), other.MinY())
	maxX := math.Max(r.MaxX(), other.MaxX())
	maxY := math.Max(r.MaxY(), other.MaxY())
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

// Inset moves each edge inward by the corresponding inset. Width and height
// are clamped to zero when the requested insets exceed the rectangle size.
func (r Rect) Inset(insets Insets) Rect {
	return Rect{
		X:      r.X + insets.Left,
		Y:      r.Y + insets.Top,
		Width:  math.Max(0, r.Width-insets.Horizontal()),
		Height: math.Max(0, r.Height-insets.Vertical()),
	}
}

// Offset translates the rectangle origin by dx and dy without changing its
// dimensions.
func (r Rect) Offset(dx, dy float64) Rect {
	return Rect{
		X:      r.X + dx,
		Y:      r.Y + dy,
		Width:  r.Width,
		Height: r.Height,
	}
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect(%.1f, %.1f, %.1f, %.1f)", r.X, r.Y, r.Width, r.Height)
}

// Insets represents independent top, right, bottom, and left distances.
type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// All creates the same inset value on all four sides.
func All(value float64) Insets {
	return Insets{Top: value, Right: value, Bottom: value, Left: value}
}

// Symmetric creates equal top/bottom insets and equal left/right insets.
func Symmetric(vertical, horizontal float64) Insets {
	return Insets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// TRBL creates insets in top-right-bottom-left order.
func TRBL(top, right, bottom, left float64) Insets {
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Horizontal returns the combined left and right inset.
func (i Insets) Horizontal() float64 {
	return i.Left + i.Right
}

// Vertical returns the combined top and bottom inset.
func (i Insets) Vertical() float64 {
	return i.Top + i.Bottom
}

// CornerRadius stores the radius of each rounded-rectangle corner separately.
// Values are not clamped here; renderers or callers decide how oversized or
// negative radii should be handled.
type CornerRadius struct {
	TopLeft     float64
	TopRight    float64
	BottomRight float64
	BottomLeft  float64
}

// RadiusUniform creates one radius value shared by all four corners.
func RadiusUniform(r float64) CornerRadius {
	return CornerRadius{TopLeft: r, TopRight: r, BottomRight: r, BottomLeft: r}
}

// RadiusSeparate creates radii in top-left, top-right, bottom-right,
// bottom-left order.
func RadiusSeparate(tl, tr, br, bl float64) CornerRadius {
	return CornerRadius{TopLeft: tl, TopRight: tr, BottomRight: br, BottomLeft: bl}
}

// Matrix2D represents a 2D affine transformation matrix in column-vector form:
// [ A  C  Tx ]
// [ B  D  Ty ]
// [ 0  0  1  ]
type Matrix2D struct {
	A, B, C, D, Tx, Ty float64
}

// IdentityMatrix returns the matrix that leaves every point unchanged.
func IdentityMatrix() Matrix2D {
	return Matrix2D{A: 1, B: 0, C: 0, D: 1, Tx: 0, Ty: 0}
}

// TranslationMatrix creates a matrix that adds (tx, ty) to every point.
func TranslationMatrix(tx, ty float64) Matrix2D {
	return Matrix2D{A: 1, B: 0, C: 0, D: 1, Tx: tx, Ty: ty}
}

// ScaleMatrix creates a matrix that scales x by sx and y by sy.
func ScaleMatrix(sx, sy float64) Matrix2D {
	return Matrix2D{A: sx, B: 0, C: 0, D: sy, Tx: 0, Ty: 0}
}

// RotationMatrix creates a counterclockwise rotation matrix for angle radians.
func RotationMatrix(angle float64) Matrix2D {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix2D{A: cos, B: sin, C: -sin, D: cos, Tx: 0, Ty: 0}
}

// Transform applies m to p, including its linear and translation components.
func (m Matrix2D) Transform(p Point) Point {
	return Point{
		X: m.A*p.X + m.C*p.Y + m.Tx,
		Y: m.B*p.X + m.D*p.Y + m.Ty,
	}
}

// Multiply returns the composition m * other.
//
// When applied to a point, the returned matrix applies other first and m
// second: (m * other).Transform(p) == m.Transform(other.Transform(p)).
func (m Matrix2D) Multiply(other Matrix2D) Matrix2D {
	return Matrix2D{
		A:  m.A*other.A + m.C*other.B,
		B:  m.B*other.A + m.D*other.B,
		C:  m.A*other.C + m.C*other.D,
		D:  m.B*other.C + m.D*other.D,
		Tx: m.A*other.Tx + m.C*other.Ty + m.Tx,
		Ty: m.B*other.Tx + m.D*other.Ty + m.Ty,
	}
}
