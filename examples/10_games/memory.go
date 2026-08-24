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

type MemoryCard struct {
	ID        int
	Symbol    string
	IsFlipped bool
	IsMatched bool
}

type MemoryGame struct {
	Cards        []MemoryCard
	FirstFlipped int // index of first card flipped in turn (-1 if none)
	Moves        int
	Matches      int
	IsWon        bool
	StartTime    time.Time
	ElapsedSecs  int
}

func NewMemoryGame() *MemoryGame {
	g := &MemoryGame{
		FirstFlipped: -1,
	}
	g.Reset()
	return g
}

func (g *MemoryGame) Reset() {
	symbols := []string{"[Go]", "[Rust]", "[Py]", "[JS]", "[DB]", "[API]", "[Git]", "[UI]"}
	var deck []string
	for _, s := range symbols {
		deck = append(deck, s, s) // pairs
	}

	// Shuffle
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	g.Cards = make([]MemoryCard, 16)
	for i, sym := range deck {
		g.Cards[i] = MemoryCard{
			ID:        i,
			Symbol:    sym,
			IsFlipped: false,
			IsMatched: false,
		}
	}

	g.FirstFlipped = -1
	g.Moves = 0
	g.Matches = 0
	g.IsWon = false
	g.StartTime = time.Now()
	g.ElapsedSecs = 0
}

func (g *MemoryGame) FlipCard(idx int) bool {
	if idx < 0 || idx >= len(g.Cards) || g.Cards[idx].IsMatched || g.Cards[idx].IsFlipped || g.IsWon {
		return false
	}

	if g.FirstFlipped == -1 {
		// First card in turn
		g.Cards[idx].IsFlipped = true
		g.FirstFlipped = idx
		return true
	}

	// Second card in turn
	g.Cards[idx].IsFlipped = true
	g.Moves++

	firstIdx := g.FirstFlipped
	if g.Cards[firstIdx].Symbol == g.Cards[idx].Symbol {
		// Match found!
		g.Cards[firstIdx].IsMatched = true
		g.Cards[idx].IsMatched = true
		g.Matches++
		g.FirstFlipped = -1

		if g.Matches == len(g.Cards)/2 {
			g.IsWon = true
			g.ElapsedSecs = int(time.Since(g.StartTime).Seconds())
		}
	} else {
		// No match - reset both
		g.Cards[firstIdx].IsFlipped = false
		g.Cards[idx].IsFlipped = false
		g.FirstFlipped = -1
	}

	return true
}

func (g *MemoryGame) Render(onFlip func(int), onReset func()) ui.Component {
	var cardRows []ui.Component

	for r := 0; r < 4; r++ {
		var cols []ui.Component
		for c := 0; c < 4; c++ {
			idx := r*4 + c
			card := g.Cards[idx]

			cardBg := color.Hex("#1E293B")
			cardBorder := color.Hex("#334155")
			btnLabel := "[ ? ]"

			if card.IsMatched {
				cardBg = color.Hex("#064E3B")
				cardBorder = color.Hex("#10B981")
				btnLabel = card.Symbol
			} else if card.IsFlipped {
				cardBg = color.Hex("#1E3A8A")
				cardBorder = color.Hex("#38BDF8")
				btnLabel = card.Symbol
			}

			cardIdx := idx
			btn := widgets.Button(btnLabel).OnClick(func() {
				onFlip(cardIdx)
			})

			cardBox := ui.Container().
				Size(100, 75).
				Bg(cardBg).
				Border(cardBorder, 1.5).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Center(btn),
				)

			cols = append(cols, cardBox)
		}
		cardRows = append(cardRows, ui.Row(cols...).GapSpacing(8))
	}

	statusBadge := widgets.Badge(fmt.Sprintf("Matched: %d / 8 Pairs", g.Matches)).Info()
	if g.IsWon {
		statusBadge = widgets.Badge(fmt.Sprintf("Victory in %d Moves!", g.Moves)).Success()
	}

	return ui.Container().
		Bg(color.Hex("#0F172A")).
		Border(color.Hex("#1E293B"), 1.5).
		Pad(geom.All(16)).
		Rounded(geom.RadiusUniform(12)).
		WithChild(
			ui.Column(
				ui.Row(
					ui.Text("Memory Card Match").Size(18).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
					widgets.Badge("Basic Tier").Success(),
					statusBadge,
					ui.Spacer(),
					widgets.Button("[Restart Game]").Primary().OnClick(onReset),
				).GapSpacing(8),

				ui.Row(
					widgets.Badge(fmt.Sprintf("Moves: %d", g.Moves)).Info(),
					widgets.Badge(fmt.Sprintf("Pairs: %d/8", g.Matches)).Success(),
					ui.Spacer(),
					ui.Text("Find all 8 matching pairs!").Size(12).Col(color.Hex("#94A3B8")),
				).GapSpacing(8),

				ui.Center(
					ui.Column(cardRows...).GapSpacing(8),
				),
			).GapSpacing(12),
		)
}
