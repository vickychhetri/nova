package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vickychhetri/nova"
	"github.com/vickychhetri/nova/core/geom"
	"github.com/vickychhetri/nova/state"
	"github.com/vickychhetri/nova/theme"
	"github.com/vickychhetri/nova/ui"
	"github.com/vickychhetri/nova/widgets"
	"github.com/vickychhetri/nova/widgets/forms"
)

func main() {
	// 1. Initialize Nova Desktop Application in Enterprise Light Theme
	app := nova.New()

	win := app.Window(
		nova.Title("Nova Payroll Enterprise — Payslip PDF Generator"),
		nova.Size(1280, 860),
		nova.Theme(theme.Light()),
	)

	// 2. Reactive UI State Signals
	empName := state.String("Alexander Wright")
	empID := state.String("EMP-80429")
	department := state.String("Engineering & Systems")
	designation := state.String("Principal Software Architect")
	payPeriod := state.String("August 2026")
	bankAccount := state.String("CHASE-XXXX-9842")
	taxID := state.String("TAX-99420-US")
	workingDays := state.Float(22)

	// Financial Earnings ($)
	basicSalary := state.Float(8500)
	hraAllowance := state.Float(2400)
	specialAllowance := state.Float(1600)
	performanceBonus := state.Float(1200)

	// Deductions ($)
	providentFund := state.Float(850)
	incomeTax := state.Float(1850)
	healthInsurance := state.Float(350)
	profTax := state.Float(150)

	statusMsg := state.String("Ready — Enter payroll metrics and click 'Generate PDF Payslip'.")
	generatedFileName := state.String("")

	// 3. Computed Financial Totals
	grossEarnings := state.Compute(func() float64 {
		return basicSalary.Get() + hraAllowance.Get() + specialAllowance.Get() + performanceBonus.Get()
	})

	totalDeductions := state.Compute(func() float64 {
		return providentFund.Get() + incomeTax.Get() + healthInsurance.Get() + profTax.Get()
	})

	netSalary := state.Compute(func() float64 {
		return grossEarnings.Get() - totalDeductions.Get()
	})

	// 4. PDF Generation Routine (Pure Go Vector PDF 1.4)
	generatePDF := func() (string, error) {
		filename := fmt.Sprintf("payslip_%s_%s.pdf", empID.Get(), strings.ReplaceAll(payPeriod.Get(), " ", "_"))

		pdfData := buildPayslipPDF(
			empName.Get(),
			empID.Get(),
			department.Get(),
			designation.Get(),
			payPeriod.Get(),
			bankAccount.Get(),
			taxID.Get(),
			int(workingDays.Get()),
			basicSalary.Get(),
			hraAllowance.Get(),
			specialAllowance.Get(),
			performanceBonus.Get(),
			providentFund.Get(),
			incomeTax.Get(),
			healthInsurance.Get(),
			profTax.Get(),
			grossEarnings.Get(),
			totalDeductions.Get(),
			netSalary.Get(),
		)

		err := os.WriteFile(filename, pdfData, 0644)
		if err != nil {
			return "", err
		}
		return filename, nil
	}

	// 5. Desktop Menu Bar (Qt QMenuBar Style with Interactive Dropdown Menus)
	loadSarah := func() {
		empName.Set("Dr. Sarah Jenkins")
		empID.Set("EMP-91823")
		department.Set("AI Research & Analytics")
		designation.Set("Chief Data Scientist")
		basicSalary.Set(9800)
		hraAllowance.Set(3000)
		specialAllowance.Set(2200)
		performanceBonus.Set(1500)
		providentFund.Set(980)
		incomeTax.Set(2400)
		healthInsurance.Set(400)
		profTax.Set(200)
		statusMsg.Set("Loaded profile: Dr. Sarah Jenkins (EMP-91823).")
	}

	loadAlex := func() {
		empName.Set("Alexander Wright")
		empID.Set("EMP-80429")
		department.Set("Engineering & Systems")
		designation.Set("Principal Software Architect")
		basicSalary.Set(8500)
		hraAllowance.Set(2400)
		specialAllowance.Set(1600)
		performanceBonus.Set(1200)
		providentFund.Set(850)
		incomeTax.Set(1850)
		healthInsurance.Set(350)
		profTax.Set(150)
		statusMsg.Set("Loaded profile: Alexander Wright (EMP-80429).")
	}

	resetForm := func() {
		empName.Set("")
		empID.Set("")
		department.Set("")
		designation.Set("")
		basicSalary.Set(0)
		hraAllowance.Set(0)
		specialAllowance.Set(0)
		performanceBonus.Set(0)
		providentFund.Set(0)
		incomeTax.Set(0)
		healthInsurance.Set(0)
		profTax.Set(0)
		statusMsg.Set("Form fields reset.")
	}

	topMenuBar := widgets.MenuBar(
		widgets.NewMenu("File",
			widgets.ShortcutItem("New Payslip Draft", "Ctrl+N", func() {
				resetForm()
				statusMsg.Set("Created new blank payslip draft.")
			}),
			widgets.ShortcutItem("Export PDF Document", "Ctrl+P", func() {
				fn, err := generatePDF()
				if err == nil {
					generatedFileName.Set(fn)
					statusMsg.Set("Exported PDF -> " + fn)
				}
			}),
			widgets.DividerItem(),
			widgets.ShortcutItem("Load Sample: Sarah", "Ctrl+1", loadSarah),
			widgets.ShortcutItem("Load Sample: Alex", "Ctrl+2", loadAlex),
			widgets.DividerItem(),
			widgets.ShortcutItem("Reset Form", "Ctrl+R", resetForm),
			widgets.ShortcutItem("Exit", "Ctrl+Q", func() {
				app.Quit()
			}),
		),
		widgets.NewMenu("Edit",
			widgets.ShortcutItem("Undo Changes", "Ctrl+Z", func() {
				statusMsg.Set("Undo executed.")
			}),
			widgets.ShortcutItem("Redo Changes", "Ctrl+Y", func() {
				statusMsg.Set("Redo executed.")
			}),
			widgets.DividerItem(),
			widgets.ShortcutItem("Select All Inputs", "Ctrl+A", nil),
		),
		widgets.NewMenu("Payroll",
			widgets.ActionItem("Recalculate CTC Allowances", func() {
				statusMsg.Set(fmt.Sprintf("Recalculated: Gross $%.2f | Net Take-Home $%.2f", grossEarnings.Get(), netSalary.Get()))
			}),
			widgets.ActionItem("Verify Tax & Statutory Compliance", func() {
				statusMsg.Set("Audit Passed: All 2026 tax withholdings comply with IRS regulations.")
			}),
			widgets.ActionItem("Authorize Direct Deposit Batch", func() {
				statusMsg.Set("Direct Deposit batch #892 authorized for automated disbursement.")
			}),
		),
		widgets.NewMenu("Reports",
			widgets.ActionItem("Monthly Payroll Summary", func() {
				statusMsg.Set("Report Generated: August 2026 Executive Summary.")
			}),
			widgets.ActionItem("Statutory Tax Withholding Register", func() {
				statusMsg.Set("Exported: Form W-2 / Withholding Tax Register.")
			}),
			widgets.ActionItem("Audit Trail & Compliance Log", func() {
				statusMsg.Set("Audit trail logs exported successfully.")
			}),
		),
		widgets.NewMenu("Settings",
			widgets.ActionItem("Theme: Light Enterprise", func() {
				statusMsg.Set("Active Theme: High-Contrast Light Enterprise.")
			}),
			widgets.ActionItem("Currency: USD ($)", func() {
				statusMsg.Set("Default Currency: United States Dollar (USD).")
			}),
			widgets.ActionItem("PDF Engine: Pure Vector 1.4", func() {
				statusMsg.Set("Vector rendering engine active.")
			}),
		),
		widgets.NewMenu("Help",
			widgets.ActionItem("Nova Documentation", func() {
				statusMsg.Set("See docs/widgets/ for detailed UI component guides.")
			}),
			widgets.ActionItem("About Payslip Generator", func() {
				statusMsg.Set("Nova Enterprise Payslip Generator v2.0 (Pure Go Desktop GUI).")
			}),
		),
	)

	// 6. Action Toolbar with High-Fidelity Desktop Controls
	mainToolbar := widgets.Toolbar(
		widgets.Button("Generate PDF Payslip").OnClick(func() {
			fn, err := generatePDF()
			if err == nil {
				generatedFileName.Set(fn)
				statusMsg.Set(fmt.Sprintf("PDF generated successfully -> %s", fn))
				fmt.Printf("✅ Generated PDF: %s\n", fn)
			} else {
				statusMsg.Set("Error generating PDF: " + err.Error())
			}
		}),

		widgets.Button("Load Sarah Jenkins").Secondary().OnClick(func() {
			empName.Set("Dr. Sarah Jenkins")
			empID.Set("EMP-91823")
			department.Set("AI Research & Analytics")
			designation.Set("Chief Data Scientist")
			basicSalary.Set(9800)
			hraAllowance.Set(3000)
			specialAllowance.Set(2200)
			performanceBonus.Set(1500)
			providentFund.Set(980)
			incomeTax.Set(2400)
			healthInsurance.Set(400)
			profTax.Set(200)
			statusMsg.Set("Loaded profile: Dr. Sarah Jenkins (EMP-91823).")
		}),

		widgets.Button("Load Alex Wright").Secondary().OnClick(func() {
			empName.Set("Alexander Wright")
			empID.Set("EMP-80429")
			department.Set("Engineering & Systems")
			designation.Set("Principal Software Architect")
			basicSalary.Set(8500)
			hraAllowance.Set(2400)
			specialAllowance.Set(1600)
			performanceBonus.Set(1200)
			providentFund.Set(850)
			incomeTax.Set(1850)
			healthInsurance.Set(350)
			profTax.Set(150)
			statusMsg.Set("Loaded profile: Alexander Wright (EMP-80429).")
		}),

		widgets.Button("Reset Form").Secondary().OnClick(func() {
			empName.Set("")
			empID.Set("")
			department.Set("")
			designation.Set("")
			basicSalary.Set(0)
			hraAllowance.Set(0)
			specialAllowance.Set(0)
			performanceBonus.Set(0)
			providentFund.Set(0)
			incomeTax.Set(0)
			healthInsurance.Set(0)
			profTax.Set(0)
			statusMsg.Set("Form fields reset.")
		}),

		widgets.Badge("ACME Global Technologies Corp").Info(),
		widgets.Badge("August 2026").Success(),
	)

	// 7. Left Navigator Sidebar
	sidebar := widgets.Sidebar("Payroll Manager",
		widgets.SidebarItem{Title: "Employee Slip", Selected: true},
		widgets.SidebarItem{Title: "Batch Processing", Badge: "480"},
		widgets.SidebarItem{Title: "Salary Structure"},
		widgets.SidebarItem{Title: "Tax & Compliance"},
		widgets.SidebarItem{Title: "Departments"},
		widgets.SidebarItem{Title: "Audit Logs"},
	)

	// 8. Employee Details Form GroupBox
	empDetailsGroupBox := widgets.GroupBox("1. Employee Profile & Position Information",
		ui.Column(
			ui.Row(
				widgets.TextField("Enter employee name").
					WithLabel("Employee Full Name").
					WithWidth(300).
					Bind(empName),

				widgets.TextField("EMP-XXXXX").
					WithLabel("Employee ID / Badge").
					WithWidth(300).
					Bind(empID),

				widgets.TextField("Engineering").
					WithLabel("Department").
					WithWidth(300).
					Bind(department),
			).GapSpacing(14),

			ui.Row(
				widgets.TextField("Job Designation").
					WithLabel("Designation / Title").
					WithWidth(300).
					Bind(designation),

				widgets.TextField("Month Year").
					WithLabel("Pay Period").
					WithWidth(300).
					Bind(payPeriod),

				widgets.TextField("Bank Account").
					WithLabel("Bank A/C & Tax ID").
					WithWidth(300).
					Bind(bankAccount),
			).GapSpacing(14),
		).GapSpacing(10),
	)

	// 9. Earnings & Deductions Form GroupBoxes (Balanced 50/50 Split)
	earningsGroupBox := widgets.GroupBox("2. Monthly Earnings & Allowances ($)",
		ui.Column(
			ui.Row(
				widgets.NumberInput(8500).Bind(basicSalary).WithLabel("Basic Salary ($)").WithPrefix("$").WithStep(500).WithWidth(210),
				widgets.NumberInput(2400).Bind(hraAllowance).WithLabel("HRA Allowance ($)").WithPrefix("$").WithStep(200).WithWidth(210),
			).GapSpacing(14),

			ui.Row(
				widgets.NumberInput(1600).Bind(specialAllowance).WithLabel("Special Allowance ($)").WithPrefix("$").WithStep(200).WithWidth(210),
				widgets.NumberInput(1200).Bind(performanceBonus).WithLabel("Performance Bonus ($)").WithPrefix("$").WithStep(100).WithWidth(210),
			).GapSpacing(14),
		).GapSpacing(10),
	)

	deductionsGroupBox := widgets.GroupBox("3. Statutory & Tax Deductions ($)",
		ui.Column(
			ui.Row(
				widgets.NumberInput(850).Bind(providentFund).WithLabel("Provident Fund / 401(k)").WithPrefix("$").WithStep(50).WithWidth(210),
				widgets.NumberInput(1850).Bind(incomeTax).WithLabel("Federal Income Tax").WithPrefix("$").WithStep(100).WithWidth(210),
			).GapSpacing(14),

			ui.Row(
				widgets.NumberInput(350).Bind(healthInsurance).WithLabel("Health Insurance").WithPrefix("$").WithStep(50).WithWidth(210),
				widgets.NumberInput(150).Bind(profTax).WithLabel("Professional Tax").WithPrefix("$").WithStep(25).WithWidth(210),
			).GapSpacing(14),
		).GapSpacing(10),
	)

	// 10. Financial Summary Cards (3 Expanded KPI Cards)
	summaryRow := ui.Row(
		ui.Expanded(widgets.Card("GROSS MONTHLY EARNINGS",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("$%.2f", grossEarnings.Get())
				})).Size(20).Weight(700),
				widgets.Badge("4 CTC Allowances").Info(),
			).GapSpacing(4),
		)),

		ui.Expanded(widgets.Card("TOTAL STATUTORY DEDUCTIONS",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("-$%.2f", totalDeductions.Get())
				})).Size(20).Weight(700),
				widgets.Badge("Tax & Compliance").Warning(),
			).GapSpacing(4),
		)),

		ui.Expanded(widgets.Card("NET PAYABLE TAKE-HOME",
			ui.Column(
				ui.Text(state.Compute(func() string {
					return fmt.Sprintf("$%.2f", netSalary.Get())
				})).Size(22).Weight(700),
				widgets.Badge("Direct Deposit Approved").Success(),
			).GapSpacing(4),
		)),
	).GapSpacing(14)

	// 11. Live Interactive Payslip Document Preview
	previewGroupBox := widgets.GroupBox("4. Official Payslip Document Preview (Live Reactive)",
		ui.Column(
			widgets.Alert("Payroll Verification & Audit", statusMsg.Get(), widgets.AlertSuccess),

			ui.Row(
				ui.Column(
					ui.Text("EMPLOYEE DETAILS").Size(11).Weight(700),
					ui.Text(state.Compute(func() string { return "Name: " + empName.Get() })).Size(12),
					ui.Text(state.Compute(func() string { return "ID: " + empID.Get() })).Size(12),
					ui.Text(state.Compute(func() string { return "Dept: " + department.Get() })).Size(12),
					ui.Text(state.Compute(func() string { return "Title: " + designation.Get() })).Size(12),
				).GapSpacing(4),

				ui.Column(
					ui.Text("EARNINGS SUMMARY").Size(11).Weight(700),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Basic: $%.2f", basicSalary.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("HRA: $%.2f", hraAllowance.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Special: $%.2f", specialAllowance.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Bonus: $%.2f", performanceBonus.Get()) })).Size(12),
				).GapSpacing(4),

				ui.Column(
					ui.Text("DEDUCTIONS SUMMARY").Size(11).Weight(700),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("PF (401k): $%.2f", providentFund.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Income Tax: $%.2f", incomeTax.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Health Ins: $%.2f", healthInsurance.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Prof Tax: $%.2f", profTax.Get()) })).Size(12),
				).GapSpacing(4),

				ui.Column(
					ui.Text("NET DISBURSEMENT").Size(11).Weight(700),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Gross: $%.2f", grossEarnings.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Deductions: -$%.2f", totalDeductions.Get()) })).Size(12),
					ui.Text(state.Compute(func() string { return fmt.Sprintf("Net Pay: $%.2f", netSalary.Get()) })).Size(14).Weight(700),
					widgets.Badge("Direct Deposit").Success(),
				).GapSpacing(4),
			).GapSpacing(36),
		).GapSpacing(10),
	)

	// 12. Bottom Status Bar (Qt QStatusBar)
	bottomStatusBar := widgets.StatusBar("System Ready — All payroll rules validated against 2026 compliance matrices.",
		widgets.StatusSegment{Text: "Theme: Light Enterprise", Width: 160},
		widgets.StatusSegment{Text: "PDF Engine: Pure Vector 1.4", Width: 180},
		widgets.StatusSegment{Text: "Currency: USD ($)", Width: 120},
		widgets.StatusSegment{Text: "Direct Deposit: Active", Width: 150},
	)

	// 13. Assemble Workspace
	rightContent := ui.Padding(geom.All(14),
		ui.Column(
			summaryRow,
			empDetailsGroupBox,
			ui.Row(
				ui.Expanded(earningsGroupBox),
				ui.Expanded(deductionsGroupBox),
			).GapSpacing(14),
			previewGroupBox,
		).GapSpacing(10),
	)

	centerLayout := ui.Row(
		sidebar,
		ui.Expanded(rightContent),
	)

	rootLayout := ui.Column(
		topMenuBar,
		mainToolbar,
		ui.Expanded(centerLayout),
		bottomStatusBar,
	)

	win.Content(rootLayout)

	fmt.Println("==========================================================")
	fmt.Println("🚀 Nova Enterprise Payslip PDF Generator Running...")
	fmt.Println("   • Theme:         White Light Enterprise Design")
	fmt.Println("   • Feature:       Interactive Payroll Forms + PDF 1.4 Output")
	fmt.Println("==========================================================")

	// Automatically generate sample PDF on run
	_, _ = generatePDF()

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
	}

	_ = win.SaveScreenshot("examples/07_payslip_generator/payslip_generator_preview.png")
	_ = win.SaveScreenshot("payslip_generator_preview.png")
	_ = forms.Required
}

// buildPayslipPDF generates valid PDF 1.4 binary data containing the official corporate payslip.
func buildPayslipPDF(
	name, id, dept, title, period, bank, tax string,
	days int,
	basic, hra, special, bonus float64,
	pf, taxDed, ins, profTax float64,
	gross, totalDed, netPay float64,
) []byte {
	var buf bytes.Buffer

	// PDF Header
	buf.WriteString("%PDF-1.4\n")
	buf.WriteString("%\xE2\xE3\xCF\xD3\n")

	// 1 0 obj: Catalog
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// 2 0 obj: Pages
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// 4 0 obj: Font Helvetica
	buf.WriteString("4 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// 5 0 obj: Font Helvetica-Bold
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>\nendobj\n")

	// Content Stream
	var stream bytes.Buffer
	dateStr := time.Now().Format("02-Jan-2006 15:04")

	stream.WriteString("q\n")

	// Page Header Banner Box
	stream.WriteString("0.15 0.35 0.75 rg\n") // Corporate Blue
	stream.WriteString("50 720 500 70 re f\n")

	// Header Text
	stream.WriteString("BT\n")
	stream.WriteString("/F2 18 Tf\n")
	stream.WriteString("1 1 1 rg\n") // White text
	stream.WriteString("65 762 Td (ACME GLOBAL TECHNOLOGIES CORP) Tj\n")
	stream.WriteString("/F1 11 Tf\n")
	stream.WriteString("0 -18 Td (Enterprise Payroll & Human Resources Division | Confidential Slip) Tj\n")
	stream.WriteString("ET\n")

	// Pay Period Title
	stream.WriteString("BT\n")
	stream.WriteString("/F2 14 Tf\n")
	stream.WriteString("0.1 0.1 0.15 rg\n")
	stream.WriteString(fmt.Sprintf("50 690 Td (MONTHLY PAYSLIP - %s) Tj\n", strings.ToUpper(period)))
	stream.WriteString("ET\n")

	// Employee Info Box Frame
	stream.WriteString("0.85 0.88 0.92 rg\n")
	stream.WriteString("50 580 500 95 re f\n")
	stream.WriteString("0.7 0.75 0.8 RG 1 w\n")
	stream.WriteString("50 580 500 95 re S\n")

	// Employee Info Text
	stream.WriteString("BT\n")
	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.2 0.2 0.25 rg\n")
	stream.WriteString(fmt.Sprintf("65 655 Td (Employee Name: %s) Tj\n", name))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Employee ID:   %s) Tj\n", id))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Department:    %s) Tj\n", dept))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Designation:   %s) Tj\n", title))

	stream.WriteString(fmt.Sprintf("320 655 Td (Pay Period:    %s) Tj\n", period))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Bank A/C:      %s) Tj\n", bank))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Tax ID / PAN:  %s) Tj\n", tax))
	stream.WriteString(fmt.Sprintf("0 -16 Td (Working Days:  %d Days) Tj\n", days))
	stream.WriteString("ET\n")

	// Earnings & Deductions Tables
	// Table Headers
	stream.WriteString("0.2 0.3 0.5 rg\n")
	stream.WriteString("50 545 240 22 re f\n") // Earnings Header
	stream.WriteString("310 545 240 22 re f\n") // Deductions Header

	stream.WriteString("BT\n")
	stream.WriteString("/F2 11 Tf\n")
	stream.WriteString("1 1 1 rg\n")
	stream.WriteString("60 552 Td (EARNINGS & ALLOWANCES) Tj\n")
	stream.WriteString("320 552 Td (STATUTORY DEDUCTIONS) Tj\n")
	stream.WriteString("ET\n")

	// Table Body Rows
	stream.WriteString("BT\n")
	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0.15 0.15 0.2 rg\n")

	// Earnings column
	stream.WriteString(fmt.Sprintf("60 525 Td (Basic Salary:           $%.2f) Tj\n", basic))
	stream.WriteString(fmt.Sprintf("0 -18 Td (House Rent Allowance:   $%.2f) Tj\n", hra))
	stream.WriteString(fmt.Sprintf("0 -18 Td (Special Allowance:      $%.2f) Tj\n", special))
	stream.WriteString(fmt.Sprintf("0 -18 Td (Performance Bonus:      $%.2f) Tj\n", bonus))

	// Deductions column
	stream.WriteString(fmt.Sprintf("320 525 Td (Provident Fund / 401k:  $%.2f) Tj\n", pf))
	stream.WriteString(fmt.Sprintf("0 -18 Td (Income Tax / TDS:        $%.2f) Tj\n", taxDed))
	stream.WriteString(fmt.Sprintf("0 -18 Td (Medical & Health Ins:    $%.2f) Tj\n", ins))
	stream.WriteString(fmt.Sprintf("0 -18 Td (Professional Tax:        $%.2f) Tj\n", profTax))
	stream.WriteString("ET\n")

	// Horizontal Separator Line
	stream.WriteString("0.7 0.75 0.8 RG 1 w\n")
	stream.WriteString("50 450 500 0 re S\n")

	// Totals Row
	stream.WriteString("BT\n")
	stream.WriteString("/F2 11 Tf\n")
	stream.WriteString("0.1 0.1 0.15 rg\n")
	stream.WriteString(fmt.Sprintf("60 432 Td (Gross Earnings: $%.2f) Tj\n", gross))
	stream.WriteString(fmt.Sprintf("320 432 Td (Total Deductions: $%.2f) Tj\n", totalDed))
	stream.WriteString("ET\n")

	// Net Salary Highlight Box (Green Banner)
	stream.WriteString("0.12 0.6 0.3 rg\n")
	stream.WriteString("50 350 500 55 re f\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F2 15 Tf\n")
	stream.WriteString("1 1 1 rg\n")
	stream.WriteString(fmt.Sprintf("65 382 Td (NET PAYABLE SALARY: $%.2f) Tj\n", netPay))
	stream.WriteString("/F1 10 Tf\n")
	stream.WriteString("0 -18 Td (Directly Credited to Registered Corporate Salary Account) Tj\n")
	stream.WriteString("ET\n")

	// Signatory & Compliance Footer
	stream.WriteString("BT\n")
	stream.WriteString("/F1 9 Tf\n")
	stream.WriteString("0.45 0.45 0.5 rg\n")
	stream.WriteString("65 240 Td (Employer Signature: _______________________) Tj\n")
	stream.WriteString("330 240 Td (Employee Signature: _______________________) Tj\n")
	stream.WriteString(fmt.Sprintf("50 160 Td (Generated on %s | System Generated Slip - No Physical Signature Required) Tj\n", dateStr))
	stream.WriteString("ET\n")

	stream.WriteString("Q\n")

	streamBytes := stream.Bytes()

	// 3 0 obj: Page
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 600 820] /Resources << /Font << /F1 4 0 R /F2 5 0 R >> >> /Contents 6 0 R >>\nendobj\n")

	// 6 0 obj: Stream
	buf.WriteString(fmt.Sprintf("6 0 obj\n<< /Length %d >>\nstream\n", len(streamBytes)))
	buf.Write(streamBytes)
	buf.WriteString("\nendstream\nendobj\n")

	// Cross-Reference Table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 7\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000015 00000 n \n")
	buf.WriteString("0000000068 00000 n \n")
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", 120))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", 240))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", 320))
	buf.WriteString(fmt.Sprintf("%010d 00000 n \n", 400))

	// Trailer
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size 7 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

	return buf.Bytes()
}
