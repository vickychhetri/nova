package render

import (
	"image"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// CommandType defines the type of drawing instruction.
type CommandType int

const (
	CmdNone CommandType = iota
	CmdFillRect
	CmdStrokeRect
	CmdFillRoundedRect
	CmdStrokeRoundedRect
	CmdFillCircle
	CmdStrokeCircle
	CmdLine
	CmdText
	CmdImage
	CmdShadow
	CmdPushClip
	CmdPopClip
	CmdPushTransform
	CmdPopTransform
)

// ShadowParams defines drop/box shadow parameters.
type ShadowParams struct {
	Color   color.Color
	OffsetX float64
	OffsetY float64
	Blur    float64
	Spread  float64
}

// Command represents a single self-contained drawing instruction for the renderer backend.
type Command struct {
	Type         CommandType
	Bounds       geom.Rect
	Color        color.Color
	StrokeColor  color.Color
	StrokeWidth  float64
	Radius       geom.CornerRadius
	P1, P2       geom.Point
	Text         string
	FontSize     float64
	FontWeight   int // 400 = regular, 700 = bold
	Image        image.Image
	Shadow       ShadowParams
	Transform    geom.Matrix2D
}
