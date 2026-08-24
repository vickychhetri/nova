package render_test

import (
	"testing"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/render"
)

// TestCanvasAndCommandBuffer verifies the basic recording contract shared by
// Canvas and CommandBuffer. It checks that drawing calls become ordered
// commands, that Canvas translations are applied when commands are recorded,
// and that clearing the buffer removes the recorded work.
func TestCanvasAndCommandBuffer(t *testing.T) {
	// The test uses the public constructors so it exercises the same setup as a
	// caller outside the render package.
	buf := render.NewCommandBuffer()
	canvas := render.NewCanvas(buf)

	// This command is recorded at the initial zero offset.
	canvas.FillRect(geom.NewRect(10, 10, 100, 50), color.Blue)
	// Save and Translate create a local coordinate scope. The following two
	// commands should contain coordinates shifted by (20, 30).
	canvas.Save()
	canvas.Translate(20, 30)
	canvas.FillRoundedRect(geom.NewRect(0, 0, 80, 40), geom.RadiusUniform(6), color.Green)
	canvas.DrawText("Hello Nova", geom.Pt(5, 20), 14, 400, color.White)
	// Restore returns Canvas to the offset that was active before Save. This
	// test does not record another command afterward, but confirms the scoped
	// transform is balanced for future drawing calls.
	canvas.Restore()

	// Commands exposes the recorded sequence in insertion order. Three drawing
	// calls were made; Save and Restore affect Canvas state only and do not add
	// entries to the command buffer.
	commands := buf.Commands()
	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(commands))
	}

	// The first rectangle was recorded before translation and should retain its
	// original position and operation type.
	cmd0 := commands[0]
	if cmd0.Type != render.CmdFillRect || cmd0.Bounds.X != 10 || cmd0.Bounds.Y != 10 {
		t.Fatalf("unexpected cmd0: %+v", cmd0)
	}

	// The rounded rectangle uses local origin (0, 0), so its recorded bounds
	// demonstrate that the active offset was applied at record time.
	cmd1 := commands[1]
	if cmd1.Type != render.CmdFillRoundedRect || cmd1.Bounds.X != 20 || cmd1.Bounds.Y != 30 {
		t.Fatalf("unexpected translated cmd1: %+v", cmd1)
	}

	// Text uses P1 as its origin. Both coordinates include the active offset,
	// and the original string is preserved in the command payload.
	cmd2 := commands[2]
	if cmd2.Type != render.CmdText || cmd2.P1.X != 25 || cmd2.P1.Y != 50 || cmd2.Text != "Hello Nova" {
		t.Fatalf("unexpected translated cmd2: %+v", cmd2)
	}

	// Clear resets the logical command length while allowing the buffer's
	// backing storage to remain reusable for a later frame.
	buf.Clear()
	if buf.Len() != 0 {
		t.Fatalf("expected 0 commands after clear, got %d", buf.Len())
	}
}
