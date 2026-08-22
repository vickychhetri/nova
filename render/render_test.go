package render_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/render"
)

func TestCanvasAndCommandBuffer(t *testing.T) {
	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)

	canvas.FillRect(geom.NewRect(10, 10, 100, 50), color.Blue)
	canvas.Save()
	canvas.Translate(20, 30)
	canvas.FillRoundedRect(geom.NewRect(0, 0, 80, 40), geom.RadiusUniform(6), color.Green)
	canvas.DrawText("Hello Nova", geom.Pt(5, 20), 14, 400, color.White)
	canvas.Restore()

	commands := buf.Commands()
	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(commands))
	}

	cmd0 := commands[0]
	if cmd0.Type != render.CmdFillRect || cmd0.Bounds.X != 10 || cmd0.Bounds.Y != 10 {
		t.Fatalf("unexpected cmd0: %+v", cmd0)
	}

	cmd1 := commands[1]
	if cmd1.Type != render.CmdFillRoundedRect || cmd1.Bounds.X != 20 || cmd1.Bounds.Y != 30 {
		t.Fatalf("unexpected translated cmd1: %+v", cmd1)
	}

	cmd2 := commands[2]
	if cmd2.Type != render.CmdText || cmd2.P1.X != 25 || cmd2.P1.Y != 50 || cmd2.Text != "Hello Nova" {
		t.Fatalf("unexpected translated cmd2: %+v", cmd2)
	}

	buf.Clear()
	if buf.Len() != 0 {
		t.Fatalf("expected 0 commands after clear, got %d", buf.Len())
	}
}
