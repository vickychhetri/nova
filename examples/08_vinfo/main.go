package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func main() {
	// 1. Initialize SQLite Database
	dbPath := "vinfo.db"
	db, err := OpenDatabase(dbPath)
	if err != nil {
		fmt.Printf("Error opening database storage: %v\n", err)
		return
	}
	defer db.Close()

	// 2. Initialize Nova Application with Light Theme
	app := nova.New()
	win := app.Window(
		nova.Title("V-Info - Personal Knowledge & Secure Information Vault"),
		nova.Size(1280, 860),
		nova.Theme(theme.Light()),
	)

	// 3. Application State Signals
	isUnlocked := state.Bool(false)
	isUnlocking := state.Bool(false)
	unlockProgress := state.Float(0.0)
	unlockStepText := state.String("Verifying master passcode...")

	// Active Views: 0: Tasks, 1: Secrets/Passwords, 2: Contacts, 3: Files, 4: Dashboard, 5: Security
	activeTab := state.Int(0)

	// Authentication Signals
	enteredPIN := state.String("")
	authStatusMsg := state.String("Enter your 4-digit master passcode to unlock (Default: 2212)")
	authStatusIsError := state.Bool(false)

	// Revealed Items Maps (Passcode-protected re-verification)
	revealedSecrets := make(map[int64]bool)
	revealedContacts := make(map[int64]bool)
	revealedEvidence := make(map[int64]bool)

	// Media Player State (Play/Pause, Scrubber)
	mediaPlaying := make(map[int64]bool)
	mediaProgress := make(map[int64]float64)

	promptTargetType := state.String("") // "secret", "contact", "evidence"
	promptTargetID := state.Int(-1)
	promptPINInput := state.String("")
	promptErrorMsg := state.String("")

	// File Upload Modal Dialog Signals
	showFileModal := state.Bool(false)
	fileModalCurrentDir := state.String(".")
	selectedFilesMap := make(map[string]bool)
	modalTagInput := state.String("VAULT-2026")
	modalNotesInput := state.String("Encrypted attachment uploaded from local system.")
	modalStatusMsg := state.String("Select one or multiple files from the list below to encrypt and upload.")

	// Passcode Change Signals
	currentPINInput := state.String("")
	newPINInput := state.String("")
	confirmPINInput := state.String("")
	pinChangeMsg := state.String("Enter current passcode and choose a new 4+ character passcode.")
	pinChangeSuccess := state.Bool(false)
	pinChangeError := state.Bool(false)

	// Task Inputs
	taskCaseInput := state.String("PRJ-2026-")
	taskTitleInput := state.String("")
	taskCategoryInput := state.String("Work")
	taskPriorityInput := state.String("High")
	taskNotesInput := state.String("")
	taskFilterCategory := state.String("All")
	taskSearchQuery := state.String("")
	taskOnlyPending := state.Bool(false)

	// Secret / Password Inputs
	secTitleInput := state.String("")
	secTypeInput := state.String("Credential")
	secUsernameInput := state.String("")
	secValueInput := state.String("")
	secTargetURIInput := state.String("")
	secNotesInput := state.String("")
	secFilterType := state.String("All")
	secSearchQuery := state.String("")

	// Person / Contact Inputs
	cntFullNameInput := state.String("")
	cntAliasInput := state.String("")
	cntRoleInput := state.String("Key Contact")
	cntPhoneInput := state.String("")
	cntEmailInput := state.String("")
	cntAddressInput := state.String("")
	cntOrgInput := state.String("")
	cntIntelInput := state.String("")
	cntFilterRole := state.String("All")
	cntSearchQuery := state.String("")

	// File & Media Inputs
	evCaseInput := state.String("DOC-2026-")
	evFilePathInput := state.String("")
	evFileNameInput := state.String("")
	evFileTypeInput := state.String("Image/Photo")
	evFileContentInput := state.String("")
	evDetailsInput := state.String("")
	evFilterType := state.String("All")
	evSearchQuery := state.String("")

	statusBannerMsg := state.String("Secure Vault Active - AES-256-GCM Encrypted | 3-Strike Protection Active.")

	// Data Lists & Stats Signals
	tasksList := state.New([]TaskItem{})
	secretsList := state.New([]VaultSecret{})
	contactsList := state.New([]PersonContact{})
	evidenceList := state.New([]EvidenceFile{})
	vaultStats := state.New(VaultStats{})

	// Refresh All Vault Data
	refreshData := func() {
		ts, _ := db.GetTasks(taskFilterCategory.Get(), taskSearchQuery.Get(), taskOnlyPending.Get())
		tasksList.Set(ts)

		ss, _ := db.GetSecrets(secFilterType.Get(), secSearchQuery.Get())
		secretsList.Set(ss)

		cs, _ := db.GetContacts(cntFilterRole.Get(), cntSearchQuery.Get())
		contactsList.Set(cs)

		es, _ := db.GetEvidence(evFilterType.Get(), evSearchQuery.Get())
		evidenceList.Set(es)

		st, _ := db.GetStats(dbPath)
		vaultStats.Set(st)
	}

	// Ensure zero unencrypted traces on filesystem (strictly offline in-memory decryption)
	_ = os.RemoveAll("/tmp/vinfo_vault_export")
	_ = os.RemoveAll("/tmp/vinfo_temp")

	// 5. Logo Component Helper (Vault Shield)
	renderVaultLogo := func(size float64) ui.Component {
		return widgets.Canvas(size, size, func(canvas *render.Canvas, bounds geom.Rect) {
			w := bounds.Width
			h := bounds.Height
			radius := geom.RadiusUniform(w * 0.22)

			// Shield Background (Deep Indigo / Gold Accent)
			shieldRect := geom.NewRect(0, 0, w, h)
			canvas.FillRoundedRect(shieldRect, radius, color.Hex("#1E3A8A"))
			canvas.StrokeRoundedRect(shieldRect, radius, color.Hex("#D97706"), 2.0)

			// Inner Accent Ring
			innerRect := geom.NewRect(w*0.08, h*0.08, w*0.84, h*0.84)
			canvas.StrokeRoundedRect(innerRect, geom.RadiusUniform(w*0.16), color.Hex("#F59E0B").WithAlpha(0.4), 1.0)

			// Emblem 'V'
			cx := w / 2.0
			cy := h / 2.0
			vText := "V"
			fontSize := w * 0.52
			tSz := text.MeasureText(vText, fontSize, font.WeightBold)
			canvas.DrawText(vText, geom.Pt(cx-tSz.Width/2.0, cy-tSz.Height/2.0+fontSize*0.08), fontSize, font.WeightBold, color.Hex("#F8FAFC"))

			// Gold Security Accent Dot
			canvas.FillCircle(geom.Pt(cx, cy+h*0.24), w*0.07, color.Hex("#FBBF24"))
		})
	}

	var rebuildView func()

	// 6. Authentication Routine
	verifyAndUnlock := func() {
		pin := strings.TrimSpace(enteredPIN.Get())
		if pin == "" {
			authStatusMsg.Set("Please enter your master passcode.")
			authStatusIsError.Set(true)
			rebuildView()
			return
		}

		ok, remaining, err := db.VerifyPIN(pin)
		if err == ErrVaultWiped {
			authStatusMsg.Set("🚨 MAXIMUM FAILED ATTEMPTS EXCEEDED (3/3). ALL VAULT DATA WIPED. Passcode reset to 2212.")
			authStatusIsError.Set(true)
			enteredPIN.Set("")
			refreshData()
			rebuildView()
			return
		}

		if err != nil && err != ErrInvalidPIN {
			authStatusMsg.Set("Authentication error. Please try again.")
			authStatusIsError.Set(true)
			rebuildView()
			return
		}

		if ok {
			authStatusIsError.Set(false)
			enteredPIN.Set("")
			isUnlocking.Set(true)
			unlockProgress.Set(0.15)
			unlockStepText.Set("Master passcode verified. Decrypting secure vault...")
			rebuildView()

			go func() {
				time.Sleep(250 * time.Millisecond)
				unlockProgress.Set(0.45)
				unlockStepText.Set("Decrypting AES-256 records, secrets, and documents...")
				rebuildView()

				time.Sleep(300 * time.Millisecond)
				unlockProgress.Set(0.80)
				unlockStepText.Set("Loading contacts directory, passwords, and encrypted files...")
				refreshData()
				rebuildView()

				time.Sleep(250 * time.Millisecond)
				unlockProgress.Set(1.0)
				unlockStepText.Set("Secure Information Vault Ready.")
				rebuildView()

				time.Sleep(200 * time.Millisecond)
				isUnlocking.Set(false)
				isUnlocked.Set(true)
				rebuildView()
			}()
		} else {
			if remaining == 1 {
				authStatusMsg.Set("🚨 WARNING: 1 ATTEMPT REMAINING! Next failed attempt will permanently wipe all vault data.")
			} else {
				authStatusMsg.Set(fmt.Sprintf("⚠️ Incorrect passcode (%d attempts remaining before full vault wipe).", remaining))
			}
			authStatusIsError.Set(true)
			enteredPIN.Set("")
			rebuildView()
		}
	}

	lockVault := func() {
		isUnlocked.Set(false)
		isUnlocking.Set(false)
		enteredPIN.Set("")
		revealedSecrets = make(map[int64]bool)
		revealedContacts = make(map[int64]bool)
		revealedEvidence = make(map[int64]bool)
		showFileModal.Set(false)
		authStatusMsg.Set("Vault locked. Enter master passcode to continue (Default: 2212).")
		authStatusIsError.Set(false)
		rebuildView()
	}

	// Passcode Change
	updatePasscode := func() {
		cur := strings.TrimSpace(currentPINInput.Get())
		newP := strings.TrimSpace(newPINInput.Get())
		conf := strings.TrimSpace(confirmPINInput.Get())

		if cur == "" || newP == "" || conf == "" {
			pinChangeMsg.Set("All passcode fields are required.")
			pinChangeError.Set(true)
			pinChangeSuccess.Set(false)
			return
		}

		if len(newP) < 4 {
			pinChangeMsg.Set("New passcode must be at least 4 characters long.")
			pinChangeError.Set(true)
			pinChangeSuccess.Set(false)
			return
		}

		if newP != conf {
			pinChangeMsg.Set("New passcode and confirmation do not match.")
			pinChangeError.Set(true)
			pinChangeSuccess.Set(false)
			return
		}

		err := db.UpdateLoginCode(cur, newP)
		if err != nil {
			pinChangeMsg.Set("Error: Current passcode is incorrect.")
			pinChangeError.Set(true)
			pinChangeSuccess.Set(false)
			return
		}

		pinChangeMsg.Set("Passcode successfully updated! All database records re-encrypted.")
		pinChangeSuccess.Set(true)
		pinChangeError.Set(false)
		currentPINInput.Set("")
		newPINInput.Set("")
		confirmPINInput.Set("")
		statusBannerMsg.Set("Master passcode updated. Re-encryption complete.")
	}

	// 7. Lock Screen UI
	buildLockScreen := func() ui.Component {
		return ui.Center(
			ui.Padding(geom.All(32),
				ui.Column(
					ui.Row(
						renderVaultLogo(76),
						ui.Column(
							ui.Row(
								ui.Text("V-Info").Size(30).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
								widgets.Badge("Confidential Vault").Info(),
								widgets.Badge("AES-256 Protected").Success(),
							).GapSpacing(10),
							ui.Text("Personal Knowledge, Secret & File Storage Vault").Size(14).Col(color.Hex("#475569")),
						).GapSpacing(4),
					).GapSpacing(18),

					widgets.Card("Security Authentication Required",
						ui.Column(
							func() ui.Component {
								if authStatusIsError.Get() {
									return widgets.Alert("Security Warning", authStatusMsg.Get(), widgets.AlertError)
								}
								return widgets.Alert("Authentication Required", authStatusMsg.Get(), widgets.AlertInfo)
							}(),

							ui.Row(
								widgets.PasswordField("Enter 4-digit PIN").
									WithLabel("Master Passcode").
									Bind(enteredPIN).
									OnSubmit(verifyAndUnlock),
								widgets.Button("[Unlock] Open Vault").OnClick(verifyAndUnlock),
							).GapSpacing(12),

							ui.Column(
								ui.Row(
									widgets.Button("1").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "1") }),
									widgets.Button("2").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "2") }),
									widgets.Button("3").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "3") }),
								).GapSpacing(8),
								ui.Row(
									widgets.Button("4").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "4") }),
									widgets.Button("5").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "5") }),
									widgets.Button("6").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "6") }),
								).GapSpacing(8),
								ui.Row(
									widgets.Button("7").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "7") }),
									widgets.Button("8").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "8") }),
									widgets.Button("9").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "9") }),
								).GapSpacing(8),
								ui.Row(
									widgets.Button("[Clear]").Danger().OnClick(func() { enteredPIN.Set("") }),
									widgets.Button("0").Secondary().OnClick(func() { enteredPIN.Set(enteredPIN.Get() + "0") }),
									widgets.Button("PIN: 2212").Secondary().OnClick(func() { enteredPIN.Set("2212") }),
								).GapSpacing(8),
							).GapSpacing(8),

							widgets.Alert("Security Protocol", "All records, passwords, and files are encrypted with AES-256-GCM. 3 consecutive failed attempts will permanently wipe all vault data.", widgets.AlertWarning),
						).GapSpacing(16),
					),
				).GapSpacing(20),
			),
		)
	}

	// 8. Unlocking Screen UI
	buildUnlockingScreen := func() ui.Component {
		p := unlockProgress.Get()
		step := unlockStepText.Get()

		return ui.Center(
			ui.Container().Size(580, 270).WithChild(
				widgets.Card("Secure Information Vault",
					ui.Column(
						ui.Row(
							renderVaultLogo(64),
							ui.Column(
								ui.Text("V-Info Secure Vault").Size(20).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
								ui.Text("Decrypting confidential records, passwords, and files...").Size(13).Col(color.Hex("#475569")),
							).GapSpacing(4),
						).GapSpacing(16),

						ui.Row(
							ui.Text(step).Size(13).Weight(font.WeightMedium).Col(color.Hex("#0284C7")),
							ui.Spacer(),
							ui.Text(fmt.Sprintf("%.0f%%", p*100)).Size(13).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
						),

						widgets.Progress(p),

						ui.Column(
							ui.Row(
								ui.Text("[v]").Col(color.Hex("#16A34A")).Weight(font.WeightBold),
								ui.Text("Master passcode authenticated").Size(12).Col(color.Hex("#334155")),
							).GapSpacing(8),
							ui.Row(
								ui.Text("[v]").Col(color.Hex("#16A34A")).Weight(font.WeightBold),
								ui.Text("AES-256 encrypted database & files decrypted").Size(12).Col(color.Hex("#334155")),
							).GapSpacing(8),
							ui.Row(
								ui.Text("[v]").Col(color.Hex("#16A34A")).Weight(font.WeightBold),
								ui.Text("Loading tasks, passwords, contacts, and media files").Size(12).Col(color.Hex("#334155")),
							).GapSpacing(8),
						).GapSpacing(6),
					).GapSpacing(14),
				),
			),
		)
	}

	// 9. Build File Open / Upload Dialog Modal
	buildFileOpenDialogModal := func() ui.Component {
		curDir := fileModalCurrentDir.Get()
		absDir, _ := filepath.Abs(curDir)

		entries, err := os.ReadDir(absDir)
		if err != nil {
			entries = nil
		}

		var dirButtons []ui.Component
		var fileRows []ui.Component

		// Up Directory Button
		parentDir := filepath.Dir(absDir)
		if parentDir != absDir {
			dirButtons = append(dirButtons, widgets.Button("[.. Up Folder]").Secondary().OnClick(func() {
				fileModalCurrentDir.Set(parentDir)
				selectedFilesMap = make(map[string]bool)
				rebuildView()
			}))
		}

		// List Directories (Max 6 shown to avoid overflow)
		dirCount := 0
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				dirCount++
				if dirCount <= 5 {
					entryName := e.Name()
					fullPath := filepath.Join(absDir, entryName)
					dirButtons = append(dirButtons, widgets.Button("["+entryName+"]").Ghost().OnClick(func() {
						fileModalCurrentDir.Set(fullPath)
						selectedFilesMap = make(map[string]bool)
						rebuildView()
					}))
				}
			}
		}

		// List Files
		for _, e := range entries {
			entry := e
			fullPath := filepath.Join(absDir, entry.Name())

			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				fi, _ := entry.Info()
				szStr := "0 B"
				if fi != nil {
					sz := fi.Size()
					if sz < 1024 {
						szStr = fmt.Sprintf("%d B", sz)
					} else if sz < 1024*1024 {
						szStr = fmt.Sprintf("%.1f KB", float64(sz)/1024.0)
					} else {
						szStr = fmt.Sprintf("%.1f MB", float64(sz)/(1024.0*1024.0))
					}
				}

				isSelected := selectedFilesMap[fullPath]
				checkIcon := "[ ]"
				rowBtn := widgets.Button(checkIcon + " " + entry.Name() + " (" + szStr + ")").Ghost()
				if isSelected {
					checkIcon = "[X]"
					rowBtn = widgets.Button(checkIcon + " " + entry.Name() + " (" + szStr + ")").Primary()
				}

				rowBtn.OnClick(func() {
					if selectedFilesMap[fullPath] {
						delete(selectedFilesMap, fullPath)
					} else {
						selectedFilesMap[fullPath] = true
					}
					rebuildView()
				})

				fileRows = append(fileRows, ui.Row(
					rowBtn,
					ui.Spacer(),
					widgets.Badge(strings.ToUpper(filepath.Ext(entry.Name()))).Info(),
				).GapSpacing(8))
			}
		}

		selectedCount := len(selectedFilesMap)

		// Limit displayed file rows to top 4 to prevent vertical overflow
		displayFileRows := fileRows
		if len(displayFileRows) > 4 {
			displayFileRows = displayFileRows[:4]
		}

		return ui.Center(
			ui.Container().Size(840, 640).WithChild(
				widgets.Card("Open File Dialog - Select & Upload Files to Secure Vault",
					ui.Column(
						widgets.Alert("File Browser", modalStatusMsg.Get(), widgets.AlertInfo),

						// Current Directory Navigation Bar
						ui.Container().
							Bg(color.Hex("#F8FAFC")).
							Border(color.Hex("#CBD5E1"), 1.0).
							Pad(geom.All(8)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Column(
									ui.Row(
										ui.Text("Current Folder:").Size(12).Weight(font.WeightBold).Col(color.Hex("#334155")),
										ui.Text(absDir).Size(12).Col(color.Hex("#0284C7")),
										ui.Spacer(),
										widgets.Button("[Select All Files]").Secondary().OnClick(func() {
											for _, e := range entries {
												if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
													selectedFilesMap[filepath.Join(absDir, e.Name())] = true
												}
											}
											rebuildView()
										}),
										widgets.Button("[Deselect All]").Secondary().OnClick(func() {
											selectedFilesMap = make(map[string]bool)
											rebuildView()
										}),
									).GapSpacing(8),

									// Subfolders row
									ifLen(len(dirButtons) > 0, ui.Row(dirButtons...).GapSpacing(6), ui.Spacer()),
								).GapSpacing(6),
							),

						// Files Area
						ui.Container().
							Bg(color.Hex("#FFFFFF")).
							Border(color.Hex("#E2E8F0"), 1.0).
							Pad(geom.All(8)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Column(
									func() []ui.Component {
										if len(displayFileRows) == 0 {
											return []ui.Component{ui.Text("No files in this folder.").Size(12).Col(color.Hex("#94A3B8"))}
										}
										res := displayFileRows
										if len(fileRows) > 4 {
											res = append(res, ui.Text(fmt.Sprintf("... and %d more files in this directory (click [Select All Files] to upload all).", len(fileRows)-4)).Size(11).Col(color.Hex("#64748B")))
										}
										return res
									}()...,
								).GapSpacing(4),
							),

						// Target Tag & Notes Inputs
						ui.Row(
							widgets.TextField("Tag / Reference (e.g. DOC-2026)").
								WithLabel("Target Tag / Reference").
								WithWidth(240).
								Bind(modalTagInput),

							widgets.TextField("Confidential description or notes...").
								WithLabel("File Notes / Description").
								WithWidth(480).
								Bind(modalNotesInput),
						).GapSpacing(12),

						// Upload and Close Buttons
						ui.Row(
							widgets.Button(fmt.Sprintf("[+] Upload & Encrypt Selected Files (%d)", selectedCount)).
								OnClick(func() {
									if selectedCount == 0 {
										modalStatusMsg.Set("Please select at least 1 file to upload.")
										rebuildView()
										return
									}

									uploaded := 0
									for path := range selectedFilesMap {
										_, err := db.UploadAndEncryptFile(modalTagInput.Get(), path, modalNotesInput.Get())
										if err == nil {
											uploaded++
										}
									}

									showFileModal.Set(false)
									selectedFilesMap = make(map[string]bool)
									statusBannerMsg.Set(fmt.Sprintf("Successfully encrypted and uploaded %d file(s) to vault.", uploaded))
									refreshData()
									rebuildView()
								}),

							widgets.Button("[Cancel / Close]").Secondary().OnClick(func() {
								showFileModal.Set(false)
								rebuildView()
							}),
						).GapSpacing(12),
					).GapSpacing(12),
				),
			),
		)
	}

	// 10. Rich Decrypted Content Viewer & Player Component
	buildDecryptedContentViewer := func(item EvidenceFile) ui.Component {
		raw := item.FileContent

		// 1. Sanitize text by removing carriage returns \r and control chars
		sanitized := strings.ReplaceAll(raw, "\r\n", "\n")
		sanitized = strings.ReplaceAll(sanitized, "\r", "\n")

		// 2. If content is JSON, pretty print it!
		if strings.HasSuffix(strings.ToLower(item.FileName), ".json") || strings.HasPrefix(strings.TrimSpace(sanitized), "{") || strings.HasPrefix(strings.TrimSpace(sanitized), "[") {
			var prettyBuf bytes.Buffer
			if err := json.Indent(&prettyBuf, []byte(sanitized), "", "  "); err == nil {
				sanitized = prettyBuf.String()
			}
		}

		// Action Bar for Decrypted File
		topToolbar := ui.Row(
			ui.Text(fmt.Sprintf("Decrypted Content Viewer — %s", item.FileName)).Size(12).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
			widgets.Badge(item.FileType).Info(),
			widgets.Badge("AES-256 Decrypted").Success(),
			widgets.Badge("In-App Only").Warning(),
			ui.Spacer(),
			widgets.Button("[Copy Text]").Secondary().OnClick(func() {
				forms.WriteClipboard(sanitized)
				statusBannerMsg.Set("Copied decrypted content to system clipboard.")
				rebuildView()
			}),
		).GapSpacing(8)

		// ----------------------------------------------------
		// A. VIDEO & AUDIO PLAYER COMPONENT (IN-APP PLAYBACK ONLY)
		// ----------------------------------------------------
		if item.FileType == "Audio/Recording" || strings.HasSuffix(strings.ToLower(item.FileName), ".mp4") || strings.HasSuffix(strings.ToLower(item.FileName), ".webm") || strings.HasSuffix(strings.ToLower(item.FileName), ".mkv") || strings.HasSuffix(strings.ToLower(item.FileName), ".mp3") || strings.HasSuffix(strings.ToLower(item.FileName), ".wav") {
			isPlaying := mediaPlaying[item.ID]
			prog := mediaProgress[item.ID]
			if prog <= 0.0 {
				prog = 0.35
			}

			playBtnText := "[Play Media]"
			if isPlaying {
				playBtnText = "[Pause]"
			}

			return ui.Container().
				Bg(color.Hex("#0F172A")).
				Border(color.Hex("#334155"), 1.0).
				Pad(geom.All(14)).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Column(
						ui.Row(
							ui.Text("Internal Media Player: "+item.FileName).Size(13).Weight(font.WeightBold).Col(color.Hex("#F8FAFC")),
							widgets.Badge(item.FileType).Info(),
							widgets.Badge("Encrypted Stream").Success(),
							ui.Spacer(),
							widgets.Badge("Zero Disk Leakage").Warning(),
						).GapSpacing(8),

						// Waveform / Video Canvas
						widgets.Canvas(920, 70, func(canvas *render.Canvas, bounds geom.Rect) {
							canvas.FillRoundedRect(bounds, geom.RadiusUniform(6), color.Hex("#1E293B"))
							w := bounds.Width
							h := bounds.Height

							// Draw animated soundwave / equalizer bars
							numBars := 40
							barW := w / float64(numBars) * 0.6
							gap := w / float64(numBars) * 0.4

							for i := 0; i < numBars; i++ {
								bx := float64(i)*(barW+gap) + 10
								// Generate variable height
								barH := 10.0 + float64((i*17)%45)
								if isPlaying {
									barH = 15.0 + float64((i*23)%50)
								}
								by := (h - barH) / 2.0
								barCol := color.Hex("#38BDF8")
								if float64(i)/float64(numBars) > prog {
									barCol = color.Hex("#475569")
								}
								canvas.FillRoundedRect(geom.NewRect(bx, by, barW, barH), geom.RadiusUniform(2), barCol)
							}
						}),

						// Scrubber & Controls
						ui.Row(
							widgets.Button(playBtnText).Primary().OnClick(func() {
								mediaPlaying[item.ID] = !mediaPlaying[item.ID]
								if mediaPlaying[item.ID] {
									mediaProgress[item.ID] = 0.45
								}
								rebuildView()
							}),

							widgets.Button("[-10s]").Secondary().OnClick(func() {
								mediaProgress[item.ID] = 0.15
								rebuildView()
							}),

							widgets.Button("[+10s]").Secondary().OnClick(func() {
								mediaProgress[item.ID] = 0.75
								rebuildView()
							}),

							ui.Text("01:42 / 03:50").Size(12).Weight(font.WeightBold).Col(color.Hex("#E2E8F0")),
							ui.Spacer(),
							widgets.Badge("Volume: 80%").Info(),
						).GapSpacing(10),
					).GapSpacing(10),
				)
		}

		// ----------------------------------------------------
		// B. PDF & DOCUMENT VIEWER COMPONENT (IN-APP VIEWER ONLY)
		// ----------------------------------------------------
		if item.FileType == "PDF/Report" {
			// Wrap lines so document stays within container frame
			wrappedLines := text.WrapLines(sanitized, 920, 12, font.WeightRegular)
			if len(wrappedLines) > 12 {
				wrappedLines = wrappedLines[:12]
				wrappedLines = append(wrappedLines, "... [Document continued in complete file]")
			}

			var docLineRows []ui.Component
			for _, line := range wrappedLines {
				docLineRows = append(docLineRows, ui.Text(line).Size(12).Col(color.Hex("#1E293B")))
			}

			return ui.Container().
				Bg(color.Hex("#FFFFFF")).
				Border(color.Hex("#CBD5E1"), 1.5).
				Pad(geom.All(14)).
				Rounded(geom.RadiusUniform(8)).
				WithChild(
					ui.Column(
						topToolbar,

						ui.Container().
							Bg(color.Hex("#F8FAFC")).
							Border(color.Hex("#E2E8F0"), 1.0).
							Pad(geom.All(12)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Column(
									ui.Row(
										ui.Text("Document Reader View (Formatted)").Size(11).Weight(font.WeightBold).Col(color.Hex("#475569")),
										ui.Spacer(),
										widgets.Badge("Page 1 of 1").Info(),
									),
									ui.Column(docLineRows...).GapSpacing(4),
								).GapSpacing(8),
							),
					).GapSpacing(10),
				)
		}

		// ----------------------------------------------------
		// C. TEXT / JSON / CODE / ARCHIVE VIEWER WITH WORD WRAP
		// ----------------------------------------------------
		// Split lines and wrap each line to max container width (860px)
		rawLines := strings.Split(sanitized, "\n")
		var displayLines []string
		for _, rl := range rawLines {
			wrapped := text.WrapLines(rl, 860, 11, font.WeightRegular)
			for _, wl := range wrapped {
				displayLines = append(displayLines, wl)
			}
		}

		// Cap display to top 12 lines with indicator
		totalLines := len(displayLines)
		if len(displayLines) > 12 {
			displayLines = displayLines[:12]
		}

		var lineItems []ui.Component
		for i, line := range displayLines {
			lineNum := fmt.Sprintf("%02d | ", i+1)
			lineItems = append(lineItems, ui.Row(
				ui.Text(lineNum).Size(11).Col(color.Hex("#94A3B8")),
				ui.Text(line).Size(11).Weight(font.WeightMedium).Col(color.Hex("#0F172A")),
			).GapSpacing(4))
		}

		if totalLines > 12 {
			lineItems = append(lineItems, ui.Text(fmt.Sprintf("... (%d total lines — use [Copy Text] to copy full content)", totalLines)).Size(11).Col(color.Hex("#64748B")))
		}

		return ui.Container().
			Bg(color.Hex("#F8FAFC")).
			Border(color.Hex("#CBD5E1"), 1.0).
			Pad(geom.All(12)).
			Rounded(geom.RadiusUniform(8)).
			WithChild(
				ui.Column(
					topToolbar,

					ui.Container().
						Bg(color.Hex("#FFFFFF")).
						Border(color.Hex("#E2E8F0"), 1.0).
						Pad(geom.All(10)).
						Rounded(geom.RadiusUniform(6)).
						WithChild(
							ui.Column(lineItems...).GapSpacing(3),
						),
				).GapSpacing(10),
			)
	}

	// 11. Build Main Application UI
	buildAuthenticatedApp := func() ui.Component {
		curTab := activeTab.Get()

		// Top Windows-Style Desktop Menu Bar with Submenus
		makeNavMenu := func(label string, tabIndex int) *ui.ButtonComponent {
			btn := widgets.Button(label).Ghost()
			if curTab == tabIndex {
				btn = widgets.Button(label).Secondary()
			}
			btn.OnClick(func() {
				activeTab.Set(tabIndex)
				rebuildView()
			})
			return btn
		}

		topMenuBar := ui.Container().
			Bg(color.Hex("#F1F5F9")).
			Border(color.Hex("#CBD5E1"), 1.0).
			Pad(geom.Insets{Top: 4, Bottom: 4, Left: 8, Right: 8}).
			WithChild(
				ui.Row(
					widgets.Button("File").Ghost().OnClick(func() {
						refreshData()
						rebuildView()
					}),

					// Primary Navigation Menus
					makeNavMenu("Tasks & To-Dos", 0),
					makeNavMenu("Confidential Secrets", 1),
					makeNavMenu("Contacts & Directory", 2),
					makeNavMenu("Files & Media", 3),
					makeNavMenu("Dashboard", 4),
					makeNavMenu("Security", 5),

					ui.Spacer(),

					ui.Text("Status: AES-256 Encrypted | 3-Strike Policy Active").Size(11).Col(color.Hex("#64748B")),
					widgets.Button("[Lock Vault]").Danger().OnClick(lockVault),
				).GapSpacing(4),
			)

		// Header Bar
		headerBar := widgets.Card("",
			ui.Row(
				renderVaultLogo(40),
				ui.Column(
					ui.Row(
						ui.Text("V-Info").Size(18).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
						ui.Text("- Personal Knowledge & Confidential Information Vault").Size(14).Weight(font.WeightMedium).Col(color.Hex("#1E293B")),
						widgets.Badge("Confidential").Warning(),
						widgets.Badge("AES-256 Protected").Success(),
					).GapSpacing(10),
					ui.Text("Secure Task Management, Password Manager, Contacts Directory & Encrypted Files.").Size(12).Col(color.Hex("#64748B")),
				).GapSpacing(2),

				ui.Spacer(),

				ui.Row(
					widgets.Button("[+ Upload Files]").Primary().OnClick(func() {
						showFileModal.Set(true)
						modalStatusMsg.Set("Select one or multiple files from the list below to encrypt and upload.")
						rebuildView()
					}),
					widgets.Button("[Refresh Vault]").Secondary().OnClick(func() {
						refreshData()
						rebuildView()
					}),
				).GapSpacing(8),
			).GapSpacing(12),
		)

		// Passcode Re-Verification Modal
		buildPasscodeVerificationModal := func(title string, description string, onSuccess func()) ui.Component {
			return widgets.Card("⚠️ Re-Enter Master Passcode to View Protected Data",
				ui.Column(
					ui.Text(description).Size(13).Col(color.Hex("#334155")),

					func() ui.Component {
						if promptErrorMsg.Get() != "" {
							return widgets.Alert("Verification Failed", promptErrorMsg.Get(), widgets.AlertError)
						}
						return ui.Spacer()
					}(),

					ui.Row(
						widgets.PasswordField("Enter Passcode").
							WithLabel("Master Passcode").
							Bind(promptPINInput).
							OnSubmit(func() {
								ok, _, _ := db.VerifyPIN(promptPINInput.Get())
								if ok {
									onSuccess()
									promptTargetID.Set(-1)
									promptTargetType.Set("")
									rebuildView()
								} else {
									promptErrorMsg.Set("Incorrect passcode. Access denied.")
									rebuildView()
								}
							}),

						widgets.Button("[Unlock & View]").Primary().OnClick(func() {
							ok, _, _ := db.VerifyPIN(promptPINInput.Get())
							if ok {
								onSuccess()
								promptTargetID.Set(-1)
								promptTargetType.Set("")
								rebuildView()
							} else {
								promptErrorMsg.Set("Incorrect passcode. Access denied.")
								rebuildView()
							}
						}),

						widgets.Button("[Cancel]").Secondary().OnClick(func() {
							promptTargetID.Set(-1)
							promptTargetType.Set("")
							rebuildView()
						}),
					).GapSpacing(12),
				).GapSpacing(10),
			)
		}

		// ----------------------------------------------------
		// TAB 0: TASKS & TO-DOS
		// ----------------------------------------------------
		buildTasksTab := func() ui.Component {
			curCat := taskCategoryInput.Get()
			makeCatBtn := func(name string) *ui.ButtonComponent {
				if curCat == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					taskCategoryInput.Set(name)
					rebuildView()
				})
				return btn
			}

			curPri := taskPriorityInput.Get()
			makePriBtn := func(name string) *ui.ButtonComponent {
				if curPri == name {
					switch name {
					case "High":
						return widgets.Button(name + " Priority").Danger()
					case "Medium":
						return widgets.Button(name + " Priority").Primary()
					default:
						return widgets.Button(name + " Priority").Secondary()
					}
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					taskPriorityInput.Set(name)
					rebuildView()
				})
				return btn
			}

			addTaskCard := widgets.Card("Add Task or Project Milestone",
				ui.Column(
					ui.Row(
						widgets.TextField("Tag (e.g. PRJ-2026-089)").
							WithLabel("Reference Tag").
							WithWidth(180).
							Bind(taskCaseInput),

						widgets.TextField("Task description or action item...").
							WithLabel("Task Title").
							WithWidth(320).
							Bind(taskTitleInput),

						ui.Column(
							ui.Text("Category").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
							ui.Row(
								makeCatBtn("Work"),
								makeCatBtn("Personal"),
								makeCatBtn("Urgent"),
								makeCatBtn("Legal"),
								makeCatBtn("Research"),
							).GapSpacing(6),
						).GapSpacing(4),

						ui.Column(
							ui.Text("Priority").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
							ui.Row(
								makePriBtn("High"),
								makePriBtn("Medium"),
								makePriBtn("Low"),
							).GapSpacing(6),
						).GapSpacing(4),

						widgets.Button("[+] Save Task").OnClick(func() {
							title := strings.TrimSpace(taskTitleInput.Get())
							if title == "" {
								statusBannerMsg.Set("Please enter a task title.")
								return
							}
							_, _ = db.AddTask(TaskItem{
								CaseID:      taskCaseInput.Get(),
								Title:       title,
								Category:    taskCategoryInput.Get(),
								Notes:       taskNotesInput.Get(),
								Priority:    taskPriorityInput.Get(),
								IsCompleted: false,
							})
							taskTitleInput.Set("")
							taskNotesInput.Set("")
							statusBannerMsg.Set(fmt.Sprintf("Task '%s' added to vault.", title))
							refreshData()
							rebuildView()
						}),
					).GapSpacing(12),

					widgets.TextField("Confidential details, instructions, or notes...").
						WithLabel("Task Notes & Instructions").
						WithWidth(800).
						Bind(taskNotesInput),
				).GapSpacing(10),
			)

			// Filter Bar
			curF := taskFilterCategory.Get()
			makeFilterBtn := func(name string) *ui.ButtonComponent {
				if curF == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					taskFilterCategory.Set(name)
					refreshData()
					rebuildView()
				})
				return btn
			}

			filterBar := ui.Row(
				ui.Text("Filter:").Weight(font.WeightBold),
				makeFilterBtn("All"),
				makeFilterBtn("Work"),
				makeFilterBtn("Personal"),
				makeFilterBtn("Urgent"),
				makeFilterBtn("Legal"),
				makeFilterBtn("Research"),

				ui.Spacer(),

				widgets.TextField("Search tasks & projects...").
					Bind(taskSearchQuery).
					WithWidth(260).
					OnChange(func(_ string) {
						refreshData()
						rebuildView()
					}),
			).GapSpacing(8)

			items := tasksList.Get()
			var taskRows []ui.Component

			for _, it := range items {
				item := it
				statusIcon := "[ ]"
				statusText := "Mark Done"
				titleStyle := font.WeightMedium
				titleColor := color.Hex("#0F172A")

				if item.IsCompleted {
					statusIcon = "[v]"
					statusText = "Completed"
					titleStyle = font.WeightRegular
					titleColor = color.Hex("#64748B")
				}

				var pBadge ui.Component
				switch item.Priority {
				case "High":
					pBadge = widgets.Badge("HIGH PRIORITY").Error()
				case "Medium":
					pBadge = widgets.Badge("MED").Warning()
				default:
					pBadge = widgets.Badge("LOW").Info()
				}

				taskCard := widgets.Card("",
					ui.Row(
						widgets.Button(statusIcon).Secondary().OnClick(func() {
							_ = db.ToggleTaskCompletion(item.ID)
							refreshData()
							rebuildView()
						}),

						ui.Column(
							ui.Row(
								widgets.Badge(item.CaseID).Info(),
								ui.Text(item.Title).Size(14).Weight(titleStyle).Col(titleColor),
								widgets.Badge(item.Category),
								pBadge,
							).GapSpacing(8),
							ui.Row(
								ui.Text(item.Notes).Size(12).Col(color.Hex("#475569")),
								ui.Text("— Saved: " + item.CreatedAt.Format("02 Jan 15:04")).Size(11).Col(color.Hex("#94A3B8")),
							).GapSpacing(8),
						).GapSpacing(4),

						ui.Spacer(),

						widgets.Button(statusText).Secondary().OnClick(func() {
							_ = db.ToggleTaskCompletion(item.ID)
							refreshData()
							rebuildView()
						}),

						widgets.Button("[x] Delete").Danger().OnClick(func() {
							_ = db.DeleteTask(item.ID)
							refreshData()
							rebuildView()
						}),
					).GapSpacing(12),
				)
				taskRows = append(taskRows, taskCard)
			}

			if len(taskRows) == 0 {
				taskRows = append(taskRows, widgets.Alert("No Tasks Found", "No tasks match the selected filters.", widgets.AlertInfo))
			}

			return ui.Column(
				addTaskCard,
				filterBar,
				ui.Column(taskRows...).GapSpacing(8),
			).GapSpacing(14)
		}

		// ----------------------------------------------------
		// TAB 1: CONFIDENTIAL SECRETS & PASSWORDS
		// ----------------------------------------------------
		buildVaultTab := func() ui.Component {
			curType := secTypeInput.Get()
			makeTypeBtn := func(name string) *ui.ButtonComponent {
				if curType == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					secTypeInput.Set(name)
					rebuildView()
				})
				return btn
			}

			addSecretCard := widgets.Card("Add Encrypted Password, Access Key, or Confidential Secret",
				ui.Column(
					ui.Row(
						widgets.TextField("e.g. Production Cloud Console, Safehouse PIN").
							WithLabel("Secret Title").
							WithWidth(300).
							Bind(secTitleInput),

						ui.Column(
							ui.Text("Secret Type").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
							ui.Row(
								makeTypeBtn("Credential"),
								makeTypeBtn("Access Key"),
								makeTypeBtn("Recovery Key"),
								makeTypeBtn("API Token"),
								makeTypeBtn("Secure Note"),
							).GapSpacing(6),
						).GapSpacing(4),

						widgets.Button("[+] Save Password / Secret").OnClick(func() {
							title := strings.TrimSpace(secTitleInput.Get())
							val := strings.TrimSpace(secValueInput.Get())
							if title == "" || val == "" {
								statusBannerMsg.Set("Please enter both title and password/secret value.")
								return
							}
							_, _ = db.AddSecret(VaultSecret{
								Title:       title,
								SecretType:  secTypeInput.Get(),
								Username:    secUsernameInput.Get(),
								SecretValue: val,
								TargetURI:   secTargetURIInput.Get(),
								Notes:       secNotesInput.Get(),
							})
							secTitleInput.Set("")
							secUsernameInput.Set("")
							secValueInput.Set("")
							secTargetURIInput.Set("")
							secNotesInput.Set("")
							statusBannerMsg.Set(fmt.Sprintf("Secret '%s' saved and encrypted with AES-256.", title))
							refreshData()
							rebuildView()
						}),
					).GapSpacing(12),

					ui.Row(
						widgets.TextField("Username / Account").
							WithLabel("Account / Username").
							WithWidth(220).
							Bind(secUsernameInput),

						widgets.PasswordField("Password / Secret Key / PIN").
							WithLabel("Secret Value / Password").
							WithWidth(260).
							Bind(secValueInput),

						widgets.TextField("System / Portal / Service").
							WithLabel("Target System / URI").
							WithWidth(260).
							Bind(secTargetURIInput),
					).GapSpacing(12),

					widgets.TextField("Confidential usage instructions or security protocols...").
						WithLabel("Security Notes").
						WithWidth(800).
						Bind(secNotesInput),
				).GapSpacing(10),
			)

			// Filter Bar
			curF := secFilterType.Get()
			makeFilterBtn := func(name string) *ui.ButtonComponent {
				if curF == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					secFilterType.Set(name)
					refreshData()
					rebuildView()
				})
				return btn
			}

			filterBar := ui.Row(
				ui.Text("Filter:").Weight(font.WeightBold),
				makeFilterBtn("All"),
				makeFilterBtn("Credential"),
				makeFilterBtn("Access Key"),
				makeFilterBtn("Recovery Key"),
				makeFilterBtn("API Token"),

				ui.Spacer(),

				widgets.TextField("Search passwords & secrets...").
					Bind(secSearchQuery).
					WithWidth(260).
					OnChange(func(_ string) {
						refreshData()
						rebuildView()
					}),
			).GapSpacing(8)

			secrets := secretsList.Get()
			var secRows []ui.Component

			for _, it := range secrets {
				item := it
				isRevealed := revealedSecrets[item.ID]

				displayedVal := "••••••••••••••••"
				if isRevealed {
					displayedVal = item.SecretValue
				}

				secCard := widgets.Card(item.Title,
					ui.Column(
						ui.Row(
							widgets.Badge(item.SecretType).Info(),
							ui.Text("Username: " + item.Username).Weight(font.WeightBold),
							ui.Text("Target: " + item.TargetURI).Size(12).Col(color.Hex("#475569")),
							ui.Spacer(),

							func() ui.Component {
								if isRevealed {
									return widgets.Button("[Hide Details]").Secondary().OnClick(func() {
										delete(revealedSecrets, item.ID)
										rebuildView()
									})
								}
								return widgets.Button("[View Details / Reveal]").Primary().OnClick(func() {
									promptTargetType.Set("secret")
									promptTargetID.Set(int(item.ID))
									promptPINInput.Set("")
									promptErrorMsg.Set("")
									rebuildView()
								})
							}(),

							widgets.Button("[x] Delete").Danger().OnClick(func() {
								_ = db.DeleteSecret(item.ID)
								delete(revealedSecrets, item.ID)
								refreshData()
								rebuildView()
							}),
						).GapSpacing(10),

						ui.Container().
							Bg(color.Hex("#F8FAFC")).
							Border(color.Hex("#CBD5E1"), 1.0).
							Pad(geom.All(10)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Row(
									ui.Text("Secret / Password:").Size(12).Weight(font.WeightBold).Col(color.Hex("#334155")),
									ui.Text(displayedVal).Size(13).Weight(font.WeightBold).Col(func() color.Color {
										if isRevealed {
											return color.Hex("#16A34A")
										}
										return color.Hex("#64748B")
									}()),
									ui.Spacer(),
									ui.Text(item.Notes).Size(12).Col(color.Hex("#475569")),
								).GapSpacing(12),
							),
					).GapSpacing(8),
				)

				secRows = append(secRows, secCard)
			}

			if len(secRows) == 0 {
				secRows = append(secRows, widgets.Alert("No Secrets Found", "No passwords or confidential records found.", widgets.AlertInfo))
			}

			// Passcode Re-Verification Prompt Modal if clicked
			var promptModal ui.Component
			if promptTargetType.Get() == "secret" && promptTargetID.Get() >= 0 {
				targetID := int64(promptTargetID.Get())
				promptModal = buildPasscodeVerificationModal(
					"⚠️ Passcode Verification",
					"Enter your master passcode to decrypt and view this confidential password/secret:",
					func() {
						revealedSecrets[targetID] = true
					},
				)
			}

			return ui.Column(
				func() ui.Component {
					if promptModal != nil {
						return promptModal
					}
					return addSecretCard
				}(),
				filterBar,
				ui.Column(secRows...).GapSpacing(8),
			).GapSpacing(14)
		}

		// ----------------------------------------------------
		// TAB 2: CONTACTS & DIRECTORY
		// ----------------------------------------------------
		buildContactsTab := func() ui.Component {
			curRole := cntRoleInput.Get()
			makeRoleBtn := func(name string) *ui.ButtonComponent {
				if curRole == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					cntRoleInput.Set(name)
					rebuildView()
				})
				return btn
			}

			addContactCard := widgets.Card("Add Contact, Associate, or Key Partner",
				ui.Column(
					ui.Row(
						widgets.TextField("Full Legal Name").
							WithLabel("Full Name").
							WithWidth(240).
							Bind(cntFullNameInput),

						widgets.TextField("Preferred Name / Alias").
							WithLabel("Alias / Nickname").
							WithWidth(180).
							Bind(cntAliasInput),

						ui.Column(
							ui.Text("Role").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
							ui.Row(
								makeRoleBtn("Key Contact"),
								makeRoleBtn("Associate"),
								makeRoleBtn("Partner"),
								makeRoleBtn("Executive"),
								makeRoleBtn("Client"),
							).GapSpacing(6),
						).GapSpacing(4),

						widgets.Button("[+] Save Contact").OnClick(func() {
							name := strings.TrimSpace(cntFullNameInput.Get())
							if name == "" {
								statusBannerMsg.Set("Please enter contact name.")
								return
							}
							_, _ = db.AddContact(PersonContact{
								FullName:          name,
								Alias:             cntAliasInput.Get(),
								Role:              cntRoleInput.Get(),
								Phone:             cntPhoneInput.Get(),
								Email:             cntEmailInput.Get(),
								Address:           cntAddressInput.Get(),
								Organization:      cntOrgInput.Get(),
								ConfidentialIntel: cntIntelInput.Get(),
							})
							cntFullNameInput.Set("")
							cntAliasInput.Set("")
							cntPhoneInput.Set("")
							cntEmailInput.Set("")
							cntAddressInput.Set("")
							cntOrgInput.Set("")
							cntIntelInput.Set("")
							statusBannerMsg.Set(fmt.Sprintf("Contact record '%s' saved and encrypted.", name))
							refreshData()
							rebuildView()
						}),
					).GapSpacing(12),

					ui.Row(
						widgets.TextField("Phone number").
							WithLabel("Phone").
							WithWidth(200).
							Bind(cntPhoneInput),

						widgets.TextField("Email address").
							WithLabel("Email").
							WithWidth(240).
							Bind(cntEmailInput),

						widgets.TextField("Physical Address / Office Location").
							WithLabel("Address / Location").
							WithWidth(280).
							Bind(cntAddressInput),

						widgets.TextField("Organization / Company / Department").
							WithLabel("Organization").
							WithWidth(200).
							Bind(cntOrgInput),
					).GapSpacing(12),

					widgets.TextField("Private notes, background, or confidential details...").
						WithLabel("Confidential Notes (Passcode Protected)").
						WithWidth(800).
						Bind(cntIntelInput),
				).GapSpacing(10),
			)

			// Filter Bar
			curF := cntFilterRole.Get()
			makeFilterBtn := func(name string) *ui.ButtonComponent {
				if curF == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					cntFilterRole.Set(name)
					refreshData()
					rebuildView()
				})
				return btn
			}

			filterBar := ui.Row(
				ui.Text("Filter:").Weight(font.WeightBold),
				makeFilterBtn("All"),
				makeFilterBtn("Key Contact"),
				makeFilterBtn("Associate"),
				makeFilterBtn("Partner"),
				makeFilterBtn("Executive"),
				makeFilterBtn("Client"),

				ui.Spacer(),

				widgets.TextField("Search contacts, emails, phones...").
					Bind(cntSearchQuery).
					WithWidth(260).
					OnChange(func(_ string) {
						refreshData()
						rebuildView()
					}),
			).GapSpacing(8)

			contacts := contactsList.Get()
			var cntRows []ui.Component

			for _, it := range contacts {
				item := it
				isRevealed := revealedContacts[item.ID]

				displayedIntel := "••••••••••••••••••••••••••••••••••••••••"
				if isRevealed {
					displayedIntel = item.ConfidentialIntel
				}

				var roleBadge ui.Component
				switch item.Role {
				case "Executive":
					roleBadge = widgets.Badge("EXECUTIVE").Error()
				case "Partner":
					roleBadge = widgets.Badge("KEY PARTNER").Warning()
				case "Key Contact":
					roleBadge = widgets.Badge("PRIMARY CONTACT").Success()
				default:
					roleBadge = widgets.Badge(item.Role).Info()
				}

				cntCard := widgets.Card(item.FullName+" "+item.Alias,
					ui.Column(
						ui.Row(
							roleBadge,
							ui.Text("Phone: "+item.Phone).Size(12).Weight(font.WeightMedium),
							ui.Text("Email: "+item.Email).Size(12).Col(color.Hex("#475569")),
							ui.Text("Location: "+item.Address).Size(12).Col(color.Hex("#475569")),
							ui.Spacer(),

							func() ui.Component {
								if isRevealed {
									return widgets.Button("[Hide Notes]").Secondary().OnClick(func() {
										delete(revealedContacts, item.ID)
										rebuildView()
									})
								}
								return widgets.Button("[View Details / Reveal]").Primary().OnClick(func() {
									promptTargetType.Set("contact")
									promptTargetID.Set(int(item.ID))
									promptPINInput.Set("")
									promptErrorMsg.Set("")
									rebuildView()
								})
							}(),

							widgets.Button("[x] Delete").Danger().OnClick(func() {
								_ = db.DeleteContact(item.ID)
								delete(revealedContacts, item.ID)
								refreshData()
								rebuildView()
							}),
						).GapSpacing(10),

						ui.Container().
							Bg(color.Hex("#F8FAFC")).
							Border(color.Hex("#CBD5E1"), 1.0).
							Pad(geom.All(10)).
							Rounded(geom.RadiusUniform(6)).
							WithChild(
								ui.Row(
									ui.Text("Confidential Notes:").Size(12).Weight(font.WeightBold).Col(color.Hex("#334155")),
									ui.Text(displayedIntel).Size(12).Weight(font.WeightMedium).Col(func() color.Color {
										if isRevealed {
											return color.Hex("#1E3A8A")
										}
										return color.Hex("#64748B")
									}()),
								).GapSpacing(12),
							),
					).GapSpacing(8),
				)
				cntRows = append(cntRows, cntCard)
			}

			if len(cntRows) == 0 {
				cntRows = append(cntRows, widgets.Alert("No Contacts Found", "No contact records match search filters.", widgets.AlertInfo))
			}

			// Passcode Re-Verification Prompt Modal for Contacts
			var promptModal ui.Component
			if promptTargetType.Get() == "contact" && promptTargetID.Get() >= 0 {
				targetID := int64(promptTargetID.Get())
				promptModal = buildPasscodeVerificationModal(
					"⚠️ Passcode Verification",
					"Enter your master passcode to decrypt and view sensitive notes for this contact:",
					func() {
						revealedContacts[targetID] = true
					},
				)
			}

			return ui.Column(
				func() ui.Component {
					if promptModal != nil {
						return promptModal
					}
					return addContactCard
				}(),
				filterBar,
				ui.Column(cntRows...).GapSpacing(8),
			).GapSpacing(14)
		}

		// ----------------------------------------------------
		// TAB 3: FILES & MEDIA (WITH OPEN FILE DIALOG MODAL & RICH VIEWERS)
		// ----------------------------------------------------
		buildEvidenceTab := func() ui.Component {
			curType := evFileTypeInput.Get()
			makeTypeBtn := func(name string) *ui.ButtonComponent {
				if curType == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					evFileTypeInput.Set(name)
					rebuildView()
				})
				return btn
			}

			addEvidenceCard := widgets.Card("Upload & Encrypt Digital Files, Images, Reports, or Documents",
				ui.Column(
					// Row 1: File Dialog Button & Path Loader
					ui.Row(
						widgets.Button("[Browse Files / Open File Dialog]").Primary().OnClick(func() {
							showFileModal.Set(true)
							modalStatusMsg.Set("Select one or multiple files from the list below to encrypt and upload.")
							rebuildView()
						}),

						widgets.TextField("Path on disk or browse files above...").
							WithLabel("File Path on Disk").
							WithWidth(300).
							Bind(evFilePathInput),

						widgets.Button("[Read & Hash File]").Secondary().OnClick(func() {
							p := strings.TrimSpace(evFilePathInput.Get())
							if p == "" {
								statusBannerMsg.Set("Please enter a file path.")
								return
							}
							data, err := os.ReadFile(p)
							if err != nil {
								statusBannerMsg.Set(fmt.Sprintf("Failed to read file: %v", err))
								return
							}
							evFileNameInput.Set(filepath.Base(p))
							evFileContentInput.Set(string(data))
							statusBannerMsg.Set(fmt.Sprintf("Loaded %s (%d bytes). Ready to encrypt.", filepath.Base(p), len(data)))
							rebuildView()
						}),

						ui.Row(
							widgets.Button("Sample Photo").Ghost().OnClick(func() {
								evCaseInput.Set("DOC-2026-089")
								evFileNameInput.Set("Warehouse_Design_Cam03.png")
								evFileTypeInput.Set("Image/Photo")
								evFileContentInput.Set("[Binary Photo Stream - 1920x1080 resolution architectural snapshot]")
								evDetailsInput.Set("High resolution design schematic.")
								rebuildView()
							}),
							widgets.Button("Sample Audio").Ghost().OnClick(func() {
								evCaseInput.Set("DOC-2026-104")
								evFileNameInput.Set("Audio_Recording_Meeting.mp3")
								evFileTypeInput.Set("Audio/Recording")
								evFileContentInput.Set("AUDIO MEETING RECORDING:\n[00:01] Speaker A: 'Production launch approved.'\n[00:15] Speaker B: 'Proceed with release.'")
								evDetailsInput.Set("Executive board meeting audio recording.")
								rebuildView()
							}),
							widgets.Button("Sample Report").Ghost().OnClick(func() {
								evCaseInput.Set("DOC-2026-077")
								evFileNameInput.Set("Quarterly_Audit_Report.pdf")
								evFileTypeInput.Set("PDF/Report")
								evFileContentInput.Set("FINANCIAL AUDIT REPORT:\nTotal Revenue: $4.2M\nNet Growth: +18.4%\nStatus: Verified and Sealed.")
								evDetailsInput.Set("Official financial report statement.")
								rebuildView()
							}),
						).GapSpacing(6),
					).GapSpacing(12),

					// Row 2: Metadata Fields
					ui.Row(
						widgets.TextField("Tag # (e.g. DOC-2026-089)").
							WithLabel("Reference Tag").
							WithWidth(180).
							Bind(evCaseInput),

						widgets.TextField("File Name (e.g. Document.pdf)").
							WithLabel("File Name").
							WithWidth(260).
							Bind(evFileNameInput),

						ui.Column(
							ui.Text("File Type").Size(12).Weight(font.WeightMedium).Col(color.Hex("#475569")),
							ui.Row(
								makeTypeBtn("Image/Photo"),
								makeTypeBtn("PDF/Report"),
								makeTypeBtn("Audio/Recording"),
								makeTypeBtn("Document"),
								makeTypeBtn("Data/Archive"),
							).GapSpacing(6),
						).GapSpacing(4),

						widgets.Button("[+] Encrypt & Save to Vault").OnClick(func() {
							name := strings.TrimSpace(evFileNameInput.Get())
							cnt := strings.TrimSpace(evFileContentInput.Get())
							if name == "" {
								statusBannerMsg.Set("Please enter file name.")
								return
							}

							hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cnt)))
							_, _ = db.AddEvidence(EvidenceFile{
								CaseNumber:  evCaseInput.Get(),
								FileName:    name,
								FileType:    evFileTypeInput.Get(),
								FileSize:    int64(len(cnt)),
								FileHash:    hash,
								FileContent: cnt,
								FileDetails: evDetailsInput.Get(),
							})
							evFileNameInput.Set("")
							evFilePathInput.Set("")
							evFileContentInput.Set("")
							evDetailsInput.Set("")
							statusBannerMsg.Set(fmt.Sprintf("File '%s' encrypted with AES-256-GCM and saved to vault.", name))
							refreshData()
							rebuildView()
						}),
					).GapSpacing(12),

					widgets.TextField("Paste raw file content, text, JSON, or code to encrypt...").
						WithLabel("Raw File Content / Data (AES-256 Encrypted at Rest)").
						WithWidth(800).
						Bind(evFileContentInput),

					widgets.TextField("Description, custody details, or private notes...").
						WithLabel("File Description & Confidential Notes").
						WithWidth(800).
						Bind(evDetailsInput),
				).GapSpacing(10),
			)

			// Filter Bar
			curF := evFilterType.Get()
			makeFilterBtn := func(name string) *ui.ButtonComponent {
				if curF == name {
					return widgets.Button(name).Primary()
				}
				btn := widgets.Button(name).Secondary()
				btn.OnClick(func() {
					evFilterType.Set(name)
					refreshData()
					rebuildView()
				})
				return btn
			}

			filterBar := ui.Row(
				ui.Text("Filter:").Weight(font.WeightBold),
				makeFilterBtn("All"),
				makeFilterBtn("Image/Photo"),
				makeFilterBtn("PDF/Report"),
				makeFilterBtn("Audio/Recording"),
				makeFilterBtn("Document"),
				makeFilterBtn("Data/Archive"),

				ui.Spacer(),

				widgets.TextField("Search files & documents...").
					Bind(evSearchQuery).
					WithWidth(260).
					OnChange(func(_ string) {
						refreshData()
						rebuildView()
					}),
			).GapSpacing(8)

			files := evidenceList.Get()
			var evRows []ui.Component

			for _, it := range files {
				item := it
				isRevealed := revealedEvidence[item.ID]

				evCard := widgets.Card(item.CaseNumber+": "+item.FileName,
					ui.Column(
						ui.Row(
							widgets.Badge(item.FileType).Info(),
							ui.Text(fmt.Sprintf("Size: %.1f KB", float64(item.FileSize)/1024.0)).Size(11).Weight(font.WeightBold).Col(color.Hex("#475569")),
							ui.Text("Hash: "+item.FileHash).Size(11).Col(color.Hex("#64748B")),
							ui.Text("Logged: "+item.CreatedAt.Format("02 Jan 15:04")).Size(11).Col(color.Hex("#94A3B8")),
							ui.Spacer(),

							func() ui.Component {
								if isRevealed {
									return widgets.Button("[Hide Content]").Secondary().OnClick(func() {
										delete(revealedEvidence, item.ID)
										rebuildView()
									})
								}
								return widgets.Button("[View Details / Reveal]").Primary().OnClick(func() {
									promptTargetType.Set("evidence")
									promptTargetID.Set(int(item.ID))
									promptPINInput.Set("")
									promptErrorMsg.Set("")
									rebuildView()
								})
							}(),

							widgets.Button("[x] Delete").Danger().OnClick(func() {
								_ = db.DeleteEvidence(item.ID)
								delete(revealedEvidence, item.ID)
								refreshData()
								rebuildView()
							}),
						).GapSpacing(10),

						ui.Padding(geom.All(4),
							ui.Text(item.FileDetails).Size(12).Col(color.Hex("#334155")),
						),

						// Decrypted Content Box or Rich Player
						func() ui.Component {
							if isRevealed {
								return buildDecryptedContentViewer(item)
							}
							return ui.Container().
								Bg(color.Hex("#F8FAFC")).
								Border(color.Hex("#CBD5E1"), 1.0).
								Pad(geom.All(10)).
								Rounded(geom.RadiusUniform(6)).
								WithChild(
									ui.Row(
										ui.Text("Decrypted File Content / Payload:").Size(11).Weight(font.WeightBold).Col(color.Hex("#334155")),
										ui.Text("•••••••••••••••••••••••••••••••••••••••••••••••• (Encrypted with AES-256-GCM)").Size(12).Col(color.Hex("#64748B")),
										ui.Spacer(),
										widgets.Badge("Protected").Warning(),
									).GapSpacing(10),
								)
						}(),
					).GapSpacing(6),
				)
				evRows = append(evRows, evCard)
			}

			if len(evRows) == 0 {
				evRows = append(evRows, widgets.Alert("No Files Found", "No encrypted files logged in vault.", widgets.AlertInfo))
			}

			// Passcode Re-Verification Prompt Modal for Files
			var promptModal ui.Component
			if promptTargetType.Get() == "evidence" && promptTargetID.Get() >= 0 {
				targetID := int64(promptTargetID.Get())
				promptModal = buildPasscodeVerificationModal(
					"⚠️ Passcode Verification",
					"Enter your master passcode to decrypt and view this confidential file:",
					func() {
						revealedEvidence[targetID] = true
					},
				)
			}

			return ui.Column(
				func() ui.Component {
					if promptModal != nil {
						return promptModal
					}
					return addEvidenceCard
				}(),
				filterBar,
				ui.Column(evRows...).GapSpacing(8),
			).GapSpacing(14)
		}

		// ----------------------------------------------------
		// TAB 4: DASHBOARD & OVERVIEW
		// ----------------------------------------------------
		buildDashboardTab := func() ui.Component {
			st := vaultStats.Get()

			statCards := ui.Row(
				widgets.Card("Active Tasks",
					ui.Column(
						ui.Text(fmt.Sprintf("%d", st.PendingTasks)).Size(26).Weight(font.WeightBold).Col(color.Hex("#D97706")),
						widgets.Badge("Pending Items").Warning(),
					).GapSpacing(4),
				),
				widgets.Card("Confidential Secrets",
					ui.Column(
						ui.Text(fmt.Sprintf("%d", st.TotalSecrets)).Size(26).Weight(font.WeightBold).Col(color.Hex("#1E3A8A")),
						widgets.Badge("Passwords & Keys").Info(),
					).GapSpacing(4),
				),
				widgets.Card("Contacts & Directory",
					ui.Column(
						ui.Text(fmt.Sprintf("%d", st.TotalContacts)).Size(26).Weight(font.WeightBold).Col(color.Hex("#7C3AED")),
						widgets.Badge("Key Associates").Info(),
					).GapSpacing(4),
				),
				widgets.Card("Encrypted Files",
					ui.Column(
						ui.Text(fmt.Sprintf("%d", st.TotalEvidence)).Size(26).Weight(font.WeightBold).Col(color.Hex("#059669")),
						widgets.Badge("Files & Media").Success(),
					).GapSpacing(4),
				),
			).GapSpacing(12)

			securityOverviewCard := widgets.Card("Vault Security Architecture",
				ui.Column(
					ui.Row(
						ui.Text("Vault Encryption:").Weight(font.WeightBold),
						ui.Text("AES-256-GCM (All Records & Uploaded Files Encrypted at Rest)").Col(color.Hex("#16A34A")),
					).GapSpacing(12),
					ui.Row(
						ui.Text("Self-Destruct Protection:").Weight(font.WeightBold),
						ui.Text("Active (Permanent 3-strike wipe on unauthorized access)").Col(color.Hex("#DC2626")),
					).GapSpacing(12),
					ui.Row(
						ui.Text("Secret Re-Verification:").Weight(font.WeightBold),
						ui.Text("Enabled (Requires passcode re-entry to reveal passwords, files & notes)").Col(color.Hex("#2563EB")),
					).GapSpacing(12),
				).GapSpacing(8),
			)

			return ui.Column(
				statCards,
				securityOverviewCard,
				widgets.Alert("Strict Offline Security Notice", "This application operates strictly offline. All data is protected against forensic extraction via AES-256 encryption.", widgets.AlertSuccess),
			).GapSpacing(14)
		}

		// ----------------------------------------------------
		// TAB 5: SECURITY & PASSCODE MANAGEMENT
		// ----------------------------------------------------
		buildSecurityTab := func() ui.Component {
			var alertBanner ui.Component
			if pinChangeSuccess.Get() {
				alertBanner = widgets.Alert("Success", pinChangeMsg.Get(), widgets.AlertSuccess)
			} else if pinChangeError.Get() {
				alertBanner = widgets.Alert("Security Error", pinChangeMsg.Get(), widgets.AlertError)
			} else {
				alertBanner = widgets.Alert("Master Passcode", "The master passcode secures your entire private vault. Default is 2212.", widgets.AlertInfo)
			}

			changePinCard := widgets.Card("Update Master Passcode",
				ui.Column(
					alertBanner,

					ui.Row(
						widgets.PasswordField("Enter current code").
							WithLabel("Current Passcode").
							WithWidth(240).
							Bind(currentPINInput),

						widgets.PasswordField("Enter new passcode").
							WithLabel("New Passcode (min 4 chars)").
							WithWidth(240).
							Bind(newPINInput),

						widgets.PasswordField("Confirm new passcode").
							WithLabel("Confirm Passcode").
							WithWidth(240).
							Bind(confirmPINInput),
					).GapSpacing(16),

					ui.Row(
						widgets.Button("Update Master Passcode").OnClick(func() {
							updatePasscode()
							rebuildView()
						}),

						widgets.Button("Clear Form").Secondary().OnClick(func() {
							currentPINInput.Set("")
							newPINInput.Set("")
							confirmPINInput.Set("")
							pinChangeMsg.Set("Form cleared.")
							pinChangeError.Set(false)
							pinChangeSuccess.Set(false)
							rebuildView()
						}),
					).GapSpacing(12),
				).GapSpacing(16),
			)

			return ui.Column(
				changePinCard,
				widgets.Card("Security Guidelines",
					ui.Column(
						ui.Text("- All saved records, contacts, passwords, and uploaded files are encrypted with AES-256.").Col(color.Hex("#334155")),
						ui.Text("- 3-Strike Security Policy: 3 consecutive wrong passcodes immediately wipes all vault data.").Col(color.Hex("#DC2626")),
						ui.Text("- To reveal sensitive passwords, files, or confidential notes, click '[View Details / Reveal]'.").Col(color.Hex("#334155")),
						ui.Text("- Always lock the vault with '[Lock Vault]' when stepping away from the device.").Col(color.Hex("#334155")),
					).GapSpacing(6),
				),
			).GapSpacing(14)
		}

		// Active tab content container
		var currentTabContent ui.Component
		switch curTab {
		case 1:
			currentTabContent = buildVaultTab()
		case 2:
			currentTabContent = buildContactsTab()
		case 3:
			currentTabContent = buildEvidenceTab()
		case 4:
			currentTabContent = buildDashboardTab()
		case 5:
			currentTabContent = buildSecurityTab()
		default:
			currentTabContent = buildTasksTab()
		}

		// Status Bar (Standard Desktop Footer)
		statusBar := ui.Container().
			Bg(color.Hex("#F1F5F9")).
			Border(color.Hex("#CBD5E1"), 1.0).
			Pad(geom.Insets{Top: 4, Bottom: 4, Left: 12, Right: 12}).
			WithChild(
				ui.Row(
					ui.Text(statusBannerMsg).Size(11).Col(color.Hex("#475569")),
					ui.Spacer(),
					ui.Text("V-Info Desktop | AES-256-GCM | Ready").Size(11).Col(color.Hex("#64748B")),
				),
			)

		return ui.Column(
			topMenuBar,
			ui.Padding(geom.Insets{Top: 8, Bottom: 8, Left: 16, Right: 16},
				ui.Column(
					headerBar,
					currentTabContent,
				).GapSpacing(10),
			),
			ui.Spacer(),
			statusBar,
		)
	}

	// 12. Dynamic Rebuild Function
	rebuildView = func() {
		if isUnlocking.Get() {
			win.Content(buildUnlockingScreen())
		} else if !isUnlocked.Get() {
			win.Content(buildLockScreen())
		} else if showFileModal.Get() {
			win.Content(buildFileOpenDialogModal())
		} else {
			win.Content(buildAuthenticatedApp())
		}
	}

	// Set initial content
	rebuildView()

	// 13. Pre-generate and Export Clean Preview Screenshots
	// 13.1 Lock Screen Preview
	isUnlocked.Set(false)
	isUnlocking.Set(false)
	showFileModal.Set(false)
	rebuildView()
	_ = win.SaveScreenshot("vinfo_lock_preview.png")
	_ = win.SaveScreenshot("examples/08_vinfo/vinfo_lock_preview.png")

	// 13.2 Authenticated Files & Media with Revealed Video/Audio/PDF Viewers
	isUnlocking.Set(false)
	isUnlocked.Set(true)
	activeTab.Set(3)
	// Reveal JSON collection item
	evSearchQuery.Set("json")
	refreshData()
	revealedEvidence[2] = true
	rebuildView()
	_ = win.SaveScreenshot("vinfo_json_viewer_preview.png")
	_ = win.SaveScreenshot("examples/08_vinfo/vinfo_json_viewer_preview.png")

	// Full Evidence View
	evSearchQuery.Set("")
	refreshData()
	revealedEvidence[1] = true
	revealedEvidence[2] = true
	revealedEvidence[3] = true
	rebuildView()
	_ = win.SaveScreenshot("vinfo_evidence_preview.png")
	_ = win.SaveScreenshot("examples/08_vinfo/vinfo_evidence_preview.png")

	// 13.3 Tasks View
	activeTab.Set(0)
	rebuildView()
	_ = win.SaveScreenshot("vinfo_preview.png")
	_ = win.SaveScreenshot("examples/08_vinfo/vinfo_preview.png")

	// 13.4 Reset to Lock Screen for Interactive Desktop User Session
	revealedEvidence = make(map[int64]bool)
	isUnlocked.Set(false)
	isUnlocking.Set(false)
	enteredPIN.Set("")
	activeTab.Set(0)
	rebuildView()

	fmt.Println("🚀 Running V-Info Personal Knowledge & Secure Information Vault...")
	fmt.Println("🔑 Default Passcode is: 2212")

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
