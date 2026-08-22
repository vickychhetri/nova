package animation_test

import (
	"testing"
	"time"

	"github.com/vickychhetri/nova/animation"
	"github.com/vickychhetri/nova/core/color"
)

func TestAnimationController(t *testing.T) {
	ctrl := animation.NewController(100 * time.Millisecond)
	ctrl.Play()

	if ctrl.Progress() != 0 {
		t.Fatalf("expected initial progress 0, got %f", ctrl.Progress())
	}

	ctrl.Step(50 * time.Millisecond)
	pMid := ctrl.Progress()
	if pMid <= 0 || pMid >= 1.0 {
		t.Fatalf("expected progress between 0 and 1 at midpoint, got %f", pMid)
	}

	ctrl.Step(60 * time.Millisecond)
	if ctrl.Progress() != 1.0 || ctrl.IsPlaying {
		t.Fatalf("expected finished progress 1.0 and isPlaying false, got %f (%v)", ctrl.Progress(), ctrl.IsPlaying)
	}
}

func TestAnimationLerp(t *testing.T) {
	v := animation.LerpFloat(10, 20, 0.5)
	if v != 15 {
		t.Fatalf("expected 15, got %f", v)
	}

	c := animation.LerpColor(color.Black, color.White, 0.5)
	if c.R < 0.49 || c.R > 0.51 {
		t.Fatalf("expected ~0.5, got %f", c.R)
	}
}
