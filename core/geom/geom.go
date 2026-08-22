package geom

import (
	"fmt"
	"math"
)

// Point represents a 2D coordinate (X, Y).
type Point struct {
	X float64
	Y float64
}

// Pt creates a new Point.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Add adds two points.
func (p Point) Add(other Point) Point {
	return Point{X: p.X + other.X, Y: p.Y + other.Y}
}

// Sub subtracts other from p.
func (p Point) Sub(other Point) Point {
	return Point{X: p.X - other.X, Y: p.Y - other.Y}
}

// Scale scales the point by factor s.
func (p Point) Scale(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Distance calculates Euclidean distance to another point.
func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Hypot(dx, dy)
}

func (p Point) String() string {
	return fmt.Sprintf("Point(%.1f, %.1f)", p.X, p.Y)
}

// Size represents a 2D dimension (Width, Height).
type Size struct {
	Width  float64
	Height float64
}

// Sz creates a new Size.
func Sz(width, height float64) Size {
	return Size{Width: math.Max(0, width), Height: math.Max(0, height)}
}

// IsZero returns true if both dimensions are 0.
func (s Size) IsZero() bool {
	return s.Width <= 0 && s.Height <= 0
}

// IsInfinite returns true if width or height is infinity.
func (s Size) IsInfinite() bool {
	return math.IsInf(s.Width, 1) || math.IsInf(s.Height, 1)
}

// Clamp restricts size within min and max size bounds.
func (s Size) Clamp(min, max Size) Size {
	return Size{
		Width:  math.Max(min.Width, math.Min(max.Width, s.Width)),
		Height: math.Max(min.Height, math.Min(max.Height, s.Height)),
	}
}

func (s Size) String() string {
	return fmt.Sprintf("Size(%.1f x %.1f)", s.Width, s.Height)
}

// Rect represents a 2D rectangle defined by origin (X, Y) and dimensions (Width, Height).
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewRect creates a new rectangle.
func NewRect(x, y, width, height float64) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  math.Max(0, width),
		Height: math.Max(0, height),
	}
}

// RectFromPoints creates a rectangle spanning from p1 to p2.
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

// Origin returns top-left point.
func (r Rect) Origin() Point {
	return Point{X: r.X, Y: r.Y}
}

// Size returns the size of the rectangle.
func (r Rect) Size() Size {
	return Size{Width: r.Width, Height: r.Height}
}

// MinX returns left edge.
func (r Rect) MinX() float64 { return r.X }

// MaxX returns right edge.
func (r Rect) MaxX() float64 { return r.X + r.Width }

// MinY returns top edge.
func (r Rect) MinY() float64 { return r.Y }

// MaxY returns bottom edge.
func (r Rect) MaxY() float64 { return r.Y + r.Height }

// Center returns the center point of the rectangle.
func (r Rect) Center() Point {
	return Point{
		X: r.X + r.Width/2,
		Y: r.Y + r.Height/2,
	}
}

// ContainsPoint checks if point (x, y) is inside the rectangle.
func (r Rect) ContainsPoint(p Point) bool {
	return p.X >= r.MinX() && p.X <= r.MaxX() && p.Y >= r.MinY() && p.Y <= r.MaxY()
}

// ContainsRect checks if another rect is fully enclosed.
func (r Rect) ContainsRect(other Rect) bool {
	return other.MinX() >= r.MinX() && other.MaxX() <= r.MaxX() &&
		other.MinY() >= r.MinY() && other.MaxY() <= r.MaxY()
}

// Intersects checks if this rectangle overlaps another.
func (r Rect) Intersects(other Rect) bool {
	return r.MinX() < other.MaxX() && r.MaxX() > other.MinX() &&
		r.MinY() < other.MaxY() && r.MaxY() > other.MinY()
}

// Intersection returns overlapping rectangle, or empty rect if no overlap.
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

// Union returns minimum bounding rectangle containing both rectangles.
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

// Inset shrinks the rectangle by given insets.
func (r Rect) Inset(insets Insets) Rect {
	return Rect{
		X:      r.X + insets.Left,
		Y:      r.Y + insets.Top,
		Width:  math.Max(0, r.Width-insets.Horizontal()),
		Height: math.Max(0, r.Height-insets.Vertical()),
	}
}

// Offset shifts the rectangle position by dx and dy.
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

// Insets represents padding/margin offsets on four sides.
type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// All creates equal insets on all sides.
func All(value float64) Insets {
	return Insets{Top: value, Right: value, Bottom: value, Left: value}
}

// Symmetric creates insets with symmetric vertical and horizontal padding.
func Symmetric(vertical, horizontal float64) Insets {
	return Insets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// TRBL creates insets specifying top, right, bottom, left explicitly.
func TRBL(top, right, bottom, left float64) Insets {
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Horizontal returns total left + right inset.
func (i Insets) Horizontal() float64 {
	return i.Left + i.Right
}

// Vertical returns total top + bottom inset.
func (i Insets) Vertical() float64 {
	return i.Top + i.Bottom
}

// CornerRadius represents corner radii for rounded rectangles.
type CornerRadius struct {
	TopLeft     float64
	TopRight    float64
	BottomRight float64
	BottomLeft  float64
}

// RadiusUniform creates equal corner radii.
func RadiusUniform(r float64) CornerRadius {
	return CornerRadius{TopLeft: r, TopRight: r, BottomRight: r, BottomLeft: r}
}

// RadiusSeparate creates corner radii for each corner individually.
func RadiusSeparate(tl, tr, br, bl float64) CornerRadius {
	return CornerRadius{TopLeft: tl, TopRight: tr, BottomRight: br, BottomLeft: bl}
}

// Matrix2D represents a 2D affine transformation matrix:
// [ A  C  Tx ]
// [ B  D  Ty ]
// [ 0  0  1  ]
type Matrix2D struct {
	A, B, C, D, Tx, Ty float64
}

// IdentityMatrix returns identity 2D matrix.
func IdentityMatrix() Matrix2D {
	return Matrix2D{A: 1, B: 0, C: 0, D: 1, Tx: 0, Ty: 0}
}

// TranslationMatrix creates a translation matrix.
func TranslationMatrix(tx, ty float64) Matrix2D {
	return Matrix2D{A: 1, B: 0, C: 0, D: 1, Tx: tx, Ty: ty}
}

// ScaleMatrix creates a scale matrix.
func ScaleMatrix(sx, sy float64) Matrix2D {
	return Matrix2D{A: sx, B: 0, C: 0, D: sy, Tx: 0, Ty: 0}
}

// RotationMatrix creates a rotation matrix (angle in radians).
func RotationMatrix(angle float64) Matrix2D {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Matrix2D{A: cos, B: sin, C: -sin, D: cos, Tx: 0, Ty: 0}
}

// Transform transforms point p by matrix.
func (m Matrix2D) Transform(p Point) Point {
	return Point{
		X: m.A*p.X + m.C*p.Y + m.Tx,
		Y: m.B*p.X + m.D*p.Y + m.Ty,
	}
}

// Multiply multiplies two matrices (m * other).
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
