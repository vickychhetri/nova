package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/color"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/font"
	"github.com/vickychhetri/nova/render"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/text"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/forms"
)

type StickyTheme struct {
	Name      string
	BgColor   color.Color
	BorderCol color.Color
	HeaderCol color.Color
	TextCol   color.Color
	BadgeBg   color.Color
}

func getTheme(colorName string) StickyTheme {
	switch strings.ToLower(colorName) {
	case "blue":
		return StickyTheme{
			Name:      "Blue",
			BgColor:   color.Hex("#F0F9FF"),
			BorderCol: color.Hex("#BAE6FD"),
			HeaderCol: color.Hex("#0284C7"),
			TextCol:   color.Hex("#0C4A6E"),
			BadgeBg:   color.Hex("#E0F2FE"),
		}
	case "green":
		return StickyTheme{
			Name:      "Green",
			BgColor:   color.Hex("#F0FDF4"),
			BorderCol: color.Hex("#BBF7D0"),
			HeaderCol: color.Hex("#16A34A"),
			TextCol:   color.Hex("#14532D"),
			BadgeBg:   color.Hex("#DCFCE7"),
		}
	case "pink":
		return StickyTheme{
			Name:      "Pink",
			BgColor:   color.Hex("#FFF1F2"),
			BorderCol: color.Hex("#FECDD3"),
			HeaderCol: color.Hex("#E11D48"),
			TextCol:   color.Hex("#881337"),
			BadgeBg:   color.Hex("#FFE4E6"),
		}
	case "purple":
		return StickyTheme{
			Name:      "Purple",
			BgColor:   color.Hex("#FAF5FF"),
			BorderCol: color.Hex("#E9D5FF"),
			HeaderCol: color.Hex("#9333EA"),
			TextCol:   color.Hex("#581C87"),
			BadgeBg:   color.Hex("#F3E8FF"),
		}
	default: // Yellow
		return StickyTheme{
			Name:      "Yellow",
			BgColor:   color.Hex("#FEFCE8"),
			BorderCol: color.Hex("#FDE047"),
			HeaderCol: color.Hex("#CA8A04"),
			TextCol:   color.Hex("#713F12"),
			BadgeBg:   color.Hex("#FEF08A"),
		}
	}
}

func main() {
	// 1. Initialize SQLite Database
	dbPath := "sticky.db"
	db, err := OpenDatabase(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer db.Close()

	// 2. Initialize Nova Application with Light Theme
	app := nova.New()
	win := app.Window(
		nova.Title("Nova Sticky — Smart Clipboard & Sticky Notes"),
		nova.Size(1080, 840),
		nova.Theme(theme.Light()),
	)

	// 3. Application State Signals
	searchQuery := state.String("")
	activeFilter := state.String("All") // "All", "Pinned", "Yellow", "Blue", "Green", "Pink", "Purple"
	toastMsg := state.String("Live Clipboard Watcher active: Copy anything in your OS (Ctrl+C) to automatically capture here.")
	toastSuccess := state.Bool(true)

	// Pagination & View Mode Signals
	pageIndex := state.Int(0)
	pageSize := state.Int(4)
	viewMode := state.String("cards") // "cards" or "compact"

	// New Note Form Signals
	newNoteContent := state.String("")
	newNoteColor := state.String("Yellow")
	newNotePinned := state.Bool(false)
	showAddPanel := state.Bool(true)

	clipsList := state.New([]StickyClip{})

	var rebuildView func()

	// Refresh List from SQLite
	refreshData := func() {
		filter := activeFilter.Get()
		onlyPinned := (filter == "Pinned")
		colorFilter := "All"
		if filter != "All" && filter != "Pinned" {
			colorFilter = filter
		}

		items, _ := db.GetClips(searchQuery.Get(), colorFilter, onlyPinned)
		clipsList.Set(items)
	}

	refreshData()

	// 4. Logo Component Helper (Sticky Note Icon)
	renderStickyLogo := func(size float64) ui.Component {
		return widgets.Canvas(size, size, func(canvas *render.Canvas, bounds geom.Rect) {
			w := bounds.Width
			h := bounds.Height
			radius := geom.RadiusUniform(w * 0.2)

			// Yellow Sticky Square
			rect := geom.NewRect(0, 0, w, h)
			canvas.FillRoundedRect(rect, radius, color.Hex("#FACC15"))
			canvas.StrokeRoundedRect(rect, radius, color.Hex("#EAB308"), 1.5)

			// Pin Accent
			canvas.FillCircle(geom.Pt(w*0.5, h*0.2), w*0.08, color.Hex("#DC2626"))

			// Note text lines
			canvas.FillRoundedRect(geom.NewRect(w*0.2, h*0.40, w*0.6, h*0.08), geom.RadiusUniform(2), color.Hex("#713F12"))
			canvas.FillRoundedRect(geom.NewRect(w*0.2, h*0.55, w*0.45, h*0.08), geom.RadiusUniform(2), color.Hex("#713F12"))
			canvas.FillRoundedRect(geom.NewRect(w*0.2, h*0.70, w*0.35, h*0.08), geom.RadiusUniform(2), color.Hex("#713F12"))
		})
	}

	// 5. Manual & Background Clipboard Import Function
	importClipboard := func() bool {
		clip := strings.TrimSpace(forms.ReadClipboard())
		if clip == "" {
			toastMsg.Set("System clipboard is currently empty.")
			if rebuildView != nil {
				rebuildView()
			}
			return false
		}
		_, err := db.AddClip(clip, "", "Yellow", false)
		if err == nil {
			toastMsg.Set(fmt.Sprintf("✓ Captured clipboard (%d chars). Click note to copy anytime!", len(clip)))
			toastSuccess.Set(true)
			pageIndex.Set(0)
			refreshData()
			if rebuildView != nil {
				rebuildView()
			}
			return true
		}
		return false
	}

	// Live Clipboard Poller (Background Watcher)
	lastTrackedClip := ""
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			currentClip := strings.TrimSpace(forms.ReadClipboard())
			if currentClip != "" && currentClip != lastTrackedClip {
				lastTrackedClip = currentClip
				_, err := db.AddClip(currentClip, "", "Yellow", false)
				if err == nil {
					toastMsg.Set(fmt.Sprintf("✓ Auto-captured from OS (%d chars). Ready to paste!", len(currentClip)))
					toastSuccess.Set(true)
					refreshData()
					if rebuildView != nil {
						rebuildView()
					}
				}
			}
		}
	}()

	// 6. Build Main Application UI
	buildMainView := func() ui.Component {
		curFilter := activeFilter.Get()

		makeFilterBtn := func(label string, val string) *ui.ButtonComponent {
			if curFilter == val {
				return widgets.Button(label).Primary()
			}
			btn := widgets.Button(label).Secondary()
			btn.OnClick(func() {
				activeFilter.Set(val)
				pageIndex.Set(0)
				refreshData()
				rebuildView()
			})
			return btn
		}

		// Top Header Bar
		headerBar := widgets.Card("",
			ui.Row(
				renderStickyLogo(42),
				ui.Column(
					ui.Row(
						ui.Text("Nova Sticky").Size(20).Weight(font.WeightBold).Col(color.Hex("#0F172A")),
						widgets.Badge("Smart Clipboard History").Info(),
						widgets.Badge("Live Watcher Active").Success(),
					).GapSpacing(8),
					ui.Text("Capture OS copies automatically. Click any sticky to copy & paste anywhere.").Size(12).Col(color.Hex("#475569")),
				).GapSpacing(2),

				ui.Spacer(),

				ui.Row(
					widgets.Button("[+ Paste from Clipboard]").Primary().OnClick(func() {
						importClipboard()
					}),
					widgets.Button("[+ New Sticky]").Secondary().OnClick(func() {
						showAddPanel.Set(!showAddPanel.Get())
						rebuildView()
					}),
					widgets.Button("[Clear History]").Danger().OnClick(func() {
						_ = db.ClearHistory(true) // Keeps pinned
						toastMsg.Set("Cleared unpinned clipboard history.")
						refreshData()
						rebuildView()
					}),
				).GapSpacing(8),
			).GapSpacing(12),
		)

		// Filter & Search Bar
		filterBar := ui.Row(
			ui.Text("View:").Weight(font.WeightBold).Col(color.Hex("#334155")),
			makeFilterBtn("All Stickies", "All"),
			makeFilterBtn("Pinned (Top)", "Pinned"),
			makeFilterBtn("Yellow", "Yellow"),
			makeFilterBtn("Blue", "Blue"),
			makeFilterBtn("Green", "Green"),
			makeFilterBtn("Pink", "Pink"),
			makeFilterBtn("Purple", "Purple"),

			ui.Spacer(),

			widgets.TextField("Search notes, code, links...").
				Bind(searchQuery).
				WithWidth(260).
				OnChange(func(_ string) {
					pageIndex.Set(0)
					refreshData()
					rebuildView()
				}),
		).GapSpacing(6)

		// Quick Add Sticky Note Panel
		var addNoteCard ui.Component
		if showAddPanel.Get() {
			curColor := newNoteColor.Get()
			makeColorSelectBtn := func(cName string) *ui.ButtonComponent {
				if curColor == cName {
					return widgets.Button("[" + cName + "]").Primary()
				}
				btn := widgets.Button(cName).Secondary()
				btn.OnClick(func() {
					newNoteColor.Set(cName)
					rebuildView()
				})
				return btn
			}

			addNoteCard = widgets.Card("Create New Sticky Note",
				ui.Column(
					widgets.TextField("Write a note, paste code snippet, URL, or reminder...").
						WithLabel("Note Content").
						WithWidth(980).
						Bind(newNoteContent),

					ui.Row(
						ui.Text("Color Tag:").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
						makeColorSelectBtn("Yellow"),
						makeColorSelectBtn("Blue"),
						makeColorSelectBtn("Green"),
						makeColorSelectBtn("Pink"),
						makeColorSelectBtn("Purple"),

						ui.Spacer(),

						widgets.Button(func() string {
							if newNotePinned.Get() {
								return "[Pinned to Top]"
							}
							return "[Pin to Top]"
						}()).Secondary().OnClick(func() {
							newNotePinned.Set(!newNotePinned.Get())
							rebuildView()
						}),

						widgets.Button("[+] Save Sticky Note").Primary().OnClick(func() {
							txt := strings.TrimSpace(newNoteContent.Get())
							if txt == "" {
								toastMsg.Set("Please enter note text to save.")
								rebuildView()
								return
							}

							_, _ = db.AddClip(txt, "", newNoteColor.Get(), newNotePinned.Get())
							newNoteContent.Set("")
							newNotePinned.Set(false)
							pageIndex.Set(0)
							toastMsg.Set("Sticky note saved and pinned to board!")
							toastSuccess.Set(true)
							refreshData()
							rebuildView()
						}),
					).GapSpacing(8),
				).GapSpacing(10),
			)
		}

		// ----------------------------------------------------
		// PAGINATION & SCROLL SLICE CALCULATION
		// ----------------------------------------------------
		allItems := clipsList.Get()
		totalItems := len(allItems)
		ps := pageSize.Get()
		if ps <= 0 {
			ps = 4
		}
		if viewMode.Get() == "compact" && ps < 10 {
			ps = 10
		}

		totalPages := (totalItems + ps - 1) / ps
		if totalPages < 1 {
			totalPages = 1
		}

		curPage := pageIndex.Get()
		if curPage >= totalPages {
			curPage = totalPages - 1
			pageIndex.Set(curPage)
		}
		if curPage < 0 {
			curPage = 0
			pageIndex.Set(curPage)
		}

		startIdx := curPage * ps
		endIdx := startIdx + ps
		if endIdx > totalItems {
			endIdx = totalItems
		}

		var displayItems []StickyClip
		if startIdx < totalItems {
			displayItems = allItems[startIdx:endIdx]
		}

		// Scroll / Pagination Navigation Bar
		makePageSizeBtn := func(sz int) *ui.ButtonComponent {
			label := fmt.Sprintf("%d", sz)
			if pageSize.Get() == sz {
				return widgets.Button(label).Primary()
			}
			btn := widgets.Button(label).Secondary()
			btn.OnClick(func() {
				pageSize.Set(sz)
				pageIndex.Set(0)
				rebuildView()
			})
			return btn
		}

		startDisplay := 0
		if totalItems > 0 {
			startDisplay = startIdx + 1
		}

		scrollNavRow := widgets.Card("",
			ui.Row(
				widgets.Badge(fmt.Sprintf("Showing %d–%d of %d clips", startDisplay, endIdx, totalItems)).Info(),
				widgets.Badge(fmt.Sprintf("Page %d of %d", curPage+1, totalPages)).Success(),

				ui.Spacer(),

				widgets.Button("[First]").Secondary().OnClick(func() {
					pageIndex.Set(0)
					rebuildView()
				}),
				widgets.Button("[▲ Prev]").Primary().OnClick(func() {
					if pageIndex.Get() > 0 {
						pageIndex.Set(pageIndex.Get() - 1)
						rebuildView()
					}
				}),
				widgets.Button("[▼ Next]").Primary().OnClick(func() {
					if pageIndex.Get() < totalPages-1 {
						pageIndex.Set(pageIndex.Get() + 1)
						rebuildView()
					}
				}),
				widgets.Button("[Last]").Secondary().OnClick(func() {
					pageIndex.Set(totalPages - 1)
					rebuildView()
				}),

				ui.Spacer(),

				ui.Text("Per page:").Size(11).Col(color.Hex("#475569")),
				makePageSizeBtn(4),
				makePageSizeBtn(8),
				makePageSizeBtn(16),
				makePageSizeBtn(50),

				ui.Spacer(),

				widgets.Button(func() string {
					if viewMode.Get() == "cards" {
						return "[Cards View]"
					}
					return "Cards View"
				}()).OnClick(func() {
					viewMode.Set("cards")
					rebuildView()
				}),
				widgets.Button(func() string {
					if viewMode.Get() == "compact" {
						return "[Compact View]"
					}
					return "Compact View"
				}()).OnClick(func() {
					viewMode.Set("compact")
					rebuildView()
				}),
			).GapSpacing(4),
		)

		// ----------------------------------------------------
		// RENDER STICKY CLIPS (CARDS OR COMPACT ROWS)
		// ----------------------------------------------------
		var cardRows []ui.Component

		for _, it := range displayItems {
			item := it
			th := getTheme(item.ColorTag)

			// Pin state badge / button
			pinBtn := widgets.Button("[Pin to Top]").Ghost()
			if item.IsPinned {
				pinBtn = widgets.Button("[Pinned]").Primary()
			}
			pinBtn.OnClick(func() {
				_ = db.TogglePin(item.ID)
				refreshData()
				rebuildView()
			})

			// Copy action
			copyBtn := widgets.Button("[Click to Copy]").Primary().OnClick(func() {
				forms.WriteClipboard(item.Content)
				_ = db.RecordCopy(item.ID)
				toastMsg.Set(fmt.Sprintf("Copied note (%d chars) to clipboard! Ready to paste (Ctrl+V).", len(item.Content)))
				toastSuccess.Set(true)
				refreshData()
				rebuildView()
			})

			// Color change cycle button
			colorCycleBtn := widgets.Button(th.Name).Ghost().OnClick(func() {
				nextColor := "Yellow"
				switch item.ColorTag {
				case "Yellow":
					nextColor = "Blue"
				case "Blue":
					nextColor = "Green"
				case "Green":
					nextColor = "Pink"
				case "Pink":
					nextColor = "Purple"
				default:
					nextColor = "Yellow"
				}
				_ = db.UpdateColor(item.ID, nextColor)
				refreshData()
				rebuildView()
			})

			// Delete button
			delBtn := widgets.Button("[x]").Danger().OnClick(func() {
				_ = db.DeleteClip(item.ID)
				refreshData()
				rebuildView()
			})

			if viewMode.Get() == "compact" {
				// COMPACT ONE-LINE ROW
				cleanOneLine := strings.ReplaceAll(strings.ReplaceAll(item.Content, "\n", "  |  "), "\t", " ")
				compactRow := ui.Container().
					WithWidth(980).
					Bg(th.BgColor).
					Border(th.BorderCol, 1.0).
					Pad(geom.Insets{Top: 6, Bottom: 6, Left: 10, Right: 10}).
					Rounded(geom.RadiusUniform(6)).
					WithChild(
						ui.Row(
							widgets.Badge(item.Category).Info(),
							widgets.Badge(th.Name).Info(),
							ui.Text(item.UpdatedAt.Format("15:04")).Size(11).Col(color.Hex("#64748B")),
							ui.Text(text.TruncateWithEllipsis(cleanOneLine, 450, 11, font.WeightMedium)).Size(11).Col(th.TextCol),
							ui.Spacer(),
							ifLen(item.CopyCount > 0, widgets.Badge(fmt.Sprintf("%dx", item.CopyCount)).Success(), ui.Spacer()),
							pinBtn,
							copyBtn,
							delBtn,
						).GapSpacing(8),
					)
				cardRows = append(cardRows, compactRow)
			} else {
				// DETAILED CARD VIEW
				wrappedLines := text.WrapLines(item.Content, 960, 12, font.WeightRegular)
				if len(wrappedLines) > 6 {
					wrappedLines = wrappedLines[:6]
					wrappedLines = append(wrappedLines, fmt.Sprintf("... (%d characters total)", len(item.Content)))
				}

				var textElements []ui.Component
				for _, wl := range wrappedLines {
					textElements = append(textElements, ui.Text(wl).Size(12).Weight(font.WeightMedium).Col(th.TextCol))
				}

				stickyCard := ui.Container().
					Bg(th.BgColor).
					Border(th.BorderCol, 1.5).
					Pad(geom.All(12)).
					Rounded(geom.RadiusUniform(8)).
					WithChild(
						ui.Column(
							// Header Row
							ui.Row(
								widgets.Badge(item.Category).Info(),
								ui.Text(fmt.Sprintf("%d words • %d chars", len(strings.Fields(item.Content)), len(item.Content))).Size(11).Col(color.Hex("#64748B")),
								ifLen(item.CopyCount > 0, widgets.Badge(fmt.Sprintf("Used %dx", item.CopyCount)).Success(), ui.Spacer()),
								ui.Text("• "+item.UpdatedAt.Format("15:04:05")).Size(11).Col(color.Hex("#94A3B8")),

								ui.Spacer(),

								colorCycleBtn,
								pinBtn,
								delBtn,
							).GapSpacing(8),

							// Content Box (Clickable to copy)
							ui.Container().
								WithWidth(980).
								Bg(color.Hex("#FFFFFF").WithAlpha(0.6)).
								Border(th.BorderCol.WithAlpha(0.5), 1.0).
								Pad(geom.All(10)).
								Rounded(geom.RadiusUniform(6)).
								WithChild(
									ui.Column(textElements...).GapSpacing(3),
								),

							// Footer Action Row
							ui.Row(
								copyBtn,
								ui.Spacer(),
								ui.Text("Click button to copy into OS system clipboard").Size(11).Col(color.Hex("#64748B")),
							).GapSpacing(8),
						).GapSpacing(8),
					)

				cardRows = append(cardRows, stickyCard)
			}
		}

		if len(cardRows) == 0 {
			cardRows = append(cardRows, widgets.Alert("No Sticky Notes Found", "Copy any text anywhere on your computer (Ctrl+C) or click [+ New Sticky] above to add notes.", widgets.AlertInfo))
		}

		// Toast / Status Banner
		toastBanner := ui.Container().
			Bg(color.Hex("#0F172A")).
			Border(color.Hex("#334155"), 1.0).
			Pad(geom.Insets{Top: 6, Bottom: 6, Left: 14, Right: 14}).
			Rounded(geom.RadiusUniform(6)).
			WithChild(
				ui.Row(
					ui.Text(toastMsg.Get()).Size(12).Weight(font.WeightMedium).Col(color.Hex("#38BDF8")),
					ui.Spacer(),
					ui.Text("Nova Sticky v1.0 | Offline & Instant").Size(11).Col(color.Hex("#94A3B8")),
				),
			)

		return ui.Padding(geom.Insets{Top: 10, Bottom: 10, Left: 16, Right: 16},
			ui.Column(
				headerBar,
				filterBar,
				ifLen(addNoteCard != nil, addNoteCard, ui.Spacer()),
				scrollNavRow,
				ui.Column(cardRows...).GapSpacing(8),
				ui.Spacer(),
				toastBanner,
			).GapSpacing(10),
		)
	}

	// 7. Dynamic Rebuild Function
	rebuildView = func() {
		win.Content(buildMainView())
	}

	rebuildView()

	// 8. Generate Preview Screenshots
	_ = win.SaveScreenshot("sticky_preview.png")
	_ = win.SaveScreenshot("examples/09_sticky/sticky_preview.png")

	// Page 2 Preview
	pageIndex.Set(1)
	rebuildView()
	_ = win.SaveScreenshot("sticky_page2_preview.png")
	_ = win.SaveScreenshot("examples/09_sticky/sticky_page2_preview.png")

	// Compact View Preview
	viewMode.Set("compact")
	pageSize.Set(10)
	pageIndex.Set(0)
	rebuildView()
	_ = win.SaveScreenshot("sticky_compact_preview.png")
	_ = win.SaveScreenshot("examples/09_sticky/sticky_compact_preview.png")

	// Reset to Default Cards View
	viewMode.Set("cards")
	pageSize.Set(4)
	pageIndex.Set(0)
	rebuildView()

	fmt.Println("🚀 Running Nova Sticky — Smart Clipboard & Sticky Notes...")
	fmt.Println("📋 Live clipboard watcher is active! Copy anything in your OS to see it appear.")

	if err := app.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func ifLen(cond bool, a, b ui.Component) ui.Component {
	if cond {
		return a
	}
	return b
}
