package animation

import (
	"math"
	"time"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
)

// EasingFunc defines a timing interpolation curve mapping [0..1] to [0..1].
type EasingFunc func(t float64) float64

// Standard Easing Curves
func Linear(t float64) float64 {
	return t
}

func EaseInQuad(t float64) float64 {
	return t * t
}

func EaseOutQuad(t float64) float64 {
	return t * (2.0 - t)
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2.0 * t * t
	}
	return -1.0 + (4.0-2.0*t)*t
}

func EaseInCubic(t float64) float64 {
	return t * t * t
}

func EaseOutCubic(t float64) float64 {
	t -= 1.0
	return t*t*t + 1.0
}

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4.0 * t * t * t
	}
	t -= 1.0
	return 4.0*t*t*t + 1.0
}

func Spring(t float64) float64 {
	return math.Pow(2, -10*t)*math.Sin((t-0.075)*(2*math.Pi)/0.3) + 1.0
}

func Bounce(t float64) float64 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

// Controller manages timing progression of an animation.
type Controller struct {
	Duration   time.Duration
	Elapsed    time.Duration
	IsPlaying  bool
	IsReversed bool
	Curve      EasingFunc
	OnUpdate   func(progress float64)
}

// NewController creates an animation controller.
func NewController(duration time.Duration) *Controller {
	return &Controller{
		Duration: duration,
		Curve:    EaseInOutQuad,
	}
}

// Play starts animation playback.
func (c *Controller) Play() {
	c.IsPlaying = true
}

// Pause pauses playback.
func (c *Controller) Pause() {
	c.IsPlaying = false
}

// Reset resets animation progress to start.
func (c *Controller) Reset() {
	c.Elapsed = 0
	c.IsPlaying = false
}

// Progress returns normalized progress [0..1].
func (c *Controller) Progress() float64 {
	if c.Duration <= 0 {
		return 1.0
	}
	p := float64(c.Elapsed) / float64(c.Duration)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	if c.IsReversed {
		p = 1.0 - p
	}
	if c.Curve != nil {
		return c.Curve(p)
	}
	return p
}

// Step advances animation by dt.
func (c *Controller) Step(dt time.Duration) {
	if !c.IsPlaying {
		return
	}
	c.Elapsed += dt
	if c.Elapsed >= c.Duration {
		c.Elapsed = c.Duration
		c.IsPlaying = false
	}
	if c.OnUpdate != nil {
		c.OnUpdate(c.Progress())
	}
}

// Interpolation / Tween helpers
func LerpFloat(from, to, t float64) float64 {
	return from + (to-from)*t
}

func LerpPoint(from, to geom.Point, t float64) geom.Point {
	return geom.Pt(LerpFloat(from.X, to.X, t), LerpFloat(from.Y, to.Y, t))
}

func LerpSize(from, to geom.Size, t float64) geom.Size {
	return geom.Sz(LerpFloat(from.Width, to.Width, t), LerpFloat(from.Height, to.Height, t))
}

func LerpColor(from, to color.Color, t float64) color.Color {
	return from.Lerp(to, t)
}
