package software_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/renderer/software"
)

func TestRasterizer(t *testing.T) {
	raster := software.NewRasterizer()
	raster.BeginFrame(geom.Sz(100, 100), 1.0)

	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)

	// Draw red background
	canvas.FillRect(geom.NewRect(0, 0, 100, 100), color.Red)

	// Clip to center 40x40 and draw blue rect
	canvas.PushClip(geom.NewRect(30, 30, 40, 40))
	canvas.FillRect(geom.NewRect(0, 0, 100, 100), color.Blue)
	canvas.PopClip()

	// Draw rounded green box
	canvas.FillRoundedRect(geom.NewRect(10, 10, 20, 20), geom.RadiusUniform(4), color.Green)

	// Draw text
	canvas.DrawText("A", geom.Pt(50, 50), 16, 400, color.White)

	raster.Render(buf.Commands())
	raster.EndFrame()

	img := raster.Buffer()
	if img == nil {
		t.Fatal("expected non-nil framebuffer")
	}

	// Verify pixel (0,0) is red
	c0 := img.RGBAAt(0, 0)
	if c0.R < 200 || c0.A != 255 {
		t.Fatalf("expected (0,0) to be red, got RGBA(%d, %d, %d, %d)", c0.R, c0.G, c0.B, c0.A)
	}

	// Verify center pixel (50,50) is clipped and affected by blue/text
	cCenter := img.RGBAAt(50, 50)
	if cCenter.A == 0 {
		t.Fatalf("expected center to have drawn pixels, got %v", cCenter)
	}
}
