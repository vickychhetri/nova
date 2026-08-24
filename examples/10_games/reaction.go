package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
)

type ReactionState int

const (
	StateIdle ReactionState = iota
	StateWaiting
	StateClickNow
	StateTooEarly
	StateRoundResult
	StateGameOver
)

type ReactionGame struct {
	State        ReactionState
	Round        int
	MaxRounds    int
	StartTime    time.Time
	TriggerTime  time.Time
	Times        []int // millisecond times
	BestTime     int
	AvgTime      int
	TargetDelay  time.Duration
}

func NewReactionGame() *ReactionGame {
	g := &ReactionGame{
		MaxRounds: 5,
		BestTime:  9999,
	}
	g.Reset()
	return g
}

func (g *ReactionGame) Reset() {
	g.State = StateIdle
	g.Round = 1
	g.Times = nil
	g.AvgTime = 0
	g.BestTime = 9999
}

func (g *ReactionGame) StartRound() {
	g.State = StateWaiting
	g.StartTime = time.Now()
	// Random delay between 1.5s and 4.0s
	delayMs := 1500 + rand.Intn(2500)
	g.TargetDelay = time.Duration(delayMs) * time.Millisecond
	g.TriggerTime = g.StartTime.Add(g.TargetDelay)
}

func (g *ReactionGame) HandleClick() {
	now := time.Now()
	switch g.State {
	case StateIdle:
		g.StartRound()

	case StateWaiting:
		if now.Before(g.TriggerTime) {
			// Clicked too early!
			g.State = StateTooEarly
		} else {
			// Clicked in time
			elapsed := int(now.Sub(g.TriggerTime).Milliseconds())
			if elapsed <= 0 {
				elapsed = 1
			}
			g.recordResult(elapsed)
		}

	case StateClickNow:
		elapsed := int(now.Sub(g.TriggerTime).Milliseconds())
		if elapsed <= 0 {
			elapsed = 1
		}
		g.recordResult(elapsed)

	case StateTooEarly, StateRoundResult:
		if g.Round >= g.MaxRounds {
			g.State = StateGameOver
		} else {
			g.Round++
			g.StartRound()
		}

	case StateGameOver:
		g.Reset()
		g.StartRound()
	}
}

func (g *ReactionGame) CheckTimer() bool {
	if g.State == StateWaiting && time.Now().After(g.TriggerTime) {
		g.State = StateClickNow
		return true
	}
	return false
}

func (g *ReactionGame) recordResult(ms int) {
	g.Times = append(g.Times, ms)
	if ms < g.BestTime {
		g.BestTime = ms
	}

	total := 0
	for _, t := range g.Times {
		total += t
	}
	g.AvgTime = total / len(g.Times)

	if g.Round >= g.MaxRounds {
		g.State = StateGameOver
	} else {
		g.State = StateRoundResult
	}
}

func (g *ReactionGame) GetRank() string {
	if g.AvgTime == 0 {
		return "N/A"
	}
	if g.AvgTime < 210 {
		return "S Rank (Godlike Reflexes)"
	}
	if g.AvgTime < 260 {
		return "A Rank (Pro Gamer)"
	}
	if g.AvgTime < 340 {
		return "B Rank (Good Reaction)"
	}
	return "C Rank (Casual)"
}

func (g *ReactionGame) Render(onClick func(), onReset func()) ui.Component {
	g.CheckTimer()

	boxBg := color.Hex("#1E293B")
	boxBorder := color.Hex("#334155")
	msg := "Click Here to Start Reaction Test"
	subMsg := fmt.Sprintf("Round %d of %d", g.Round, g.MaxRounds)

	switch g.State {
	case StateWaiting:
		boxBg = color.Hex("#7F1D1D") // Deep Red
		boxBorder = color.Hex("#EF4444")
		msg = "WAIT FOR GREEN..."
		subMsg = "Do not click yet!"

	case StateClickNow:
		boxBg = color.Hex("#065F46") // Bright Green
		boxBorder = color.Hex("#10B981")
		msg = ">>> CLICK NOW! <<<"
		subMsg = "Fastest reflex wins!"

	case StateTooEarly:
		boxBg = color.Hex("#9A3412") // Amber Orange
		boxBorder = color.Hex("#F97316")
		msg = "TOO EARLY!"
		subMsg = "Click to try next round"

	case StateRoundResult:
		boxBg = color.Hex("#1E3A8A")
		boxBorder = color.Hex("#38BDF8")
		lastTime := 0
		if len(g.Times) > 0 {
			lastTime = g.Times[len(g.Times)-1]
		}
		msg = fmt.Sprintf("Reaction Time: %d ms", lastTime)
		subMsg = "Click anywhere to continue"

	case StateGameOver:
		boxBg = color.Hex("#312E81")
		boxBorder = color.Hex("#818CF8")
		msg = fmt.Sprintf("Final Average: %d ms (%s)", g.AvgTime, g.GetRank())
		subMsg = fmt.Sprintf("Best Time: %d ms | Click to Play Again", g.BestTime)
	}

	btn := widgets.Button(msg).Primary().OnClick(onClick)

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Reflex & Reaction Speed").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Basic Tier").Success(),
					widgets.Badge(fmt.Sprintf("Round %d/%d", g.Round, g.MaxRounds)).Info(),
					ui.Spacer(),
					widgets.Button("[Restart]").Secondary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					ifLen(g.BestTime < 9000, widgets.Badge(fmt.Sprintf("Best: %d ms", g.BestTime)).Success(), ui.Spacer()),
					ifLen(g.AvgTime > 0, widgets.Badge(fmt.Sprintf("Average: %d ms", g.AvgTime)).Info(), ui.Spacer()),
					ifLen(g.AvgTime > 0, widgets.Badge(g.GetRank()).Warning(), ui.Spacer()),
				).GapSpacing(8),

				ui.Container().
					WithWidth(940).
					Size(940, 260).
					Bg(boxBg).
					Border(boxBorder, 2.0).
					Pad(geom.All(20)).
					Rounded(geom.RadiusUniform(10)).
					WithChild(
						ui.Center(
							ui.Column(
								ui.Text(msg).Size(24).Weight(font.WeightBold).Col(color.Hex("#FFFFFF")),
								ui.Text(subMsg).Size(14).Col(color.Hex("#E2E8F0")),
								btn,
							).GapSpacing(12),
						),
					),
			).GapSpacing(12),
		)
}
