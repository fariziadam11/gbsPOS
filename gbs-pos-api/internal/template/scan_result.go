package template

import (
	"embed"
	"html/template"
)

//go:embed scan_result.html
var scanResultFS embed.FS

var scanResultTmpl = template.Must(template.ParseFS(scanResultFS, "scan_result.html"))

// ScanResultOutcome describes the result of opening a dispenser authorization
// scan URL.
const (
	ScanOutcomeSuccess     = "success"      // PAID → AUTHORIZED happened now
	ScanOutcomeAlreadyUsed = "already_used" // token was used before
	ScanOutcomeNotFound    = "not_found"    // unknown token
)

// ScanResultPageData is the data model for scan_result.html.
type ScanResultPageData struct {
	Outcome      string // one of ScanOutcome*
	ShowDetails  bool   // show sale details (only on success)
	PumpID       string
	NozzleID     string
	FuelName     string
	Liters       string
	TotalAmount  string
}

func RenderScanResultPage(data ScanResultPageData) (string, error) {
	var buf templateBuffer
	if err := scanResultTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
