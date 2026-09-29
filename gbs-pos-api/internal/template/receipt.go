package template

import (
	"embed"
	"html/template"
)

//go:embed receipt.html
var receiptFS embed.FS

var receiptTmpl = template.Must(template.ParseFS(receiptFS, "receipt.html"))

// ReceiptPageData is the data model for receipt.html.
type ReceiptPageData struct {
	ID            string
	Timestamp     string
	PumpID        string
	NozzleID      string
	FuelName      string
	PricePerLiter string
	Liters        string
	TotalAmount   string
	PaymentMethod string
	Status        string
	AuthorizedAt  string
	// AuthQRDataURI is the dispenser authorization QR as a PNG data URI.
	// Declared as template.URL so html/template does not strip the data: scheme
	// (a plain string would be replaced with #ZgotmplZ by the urlFilter).
	AuthQRDataURI template.URL
	AlreadyUsed   bool // true when the authorization token was already scanned
	// AuthorizationToken is the raw single-use token encoded in the QR. Injected
	// into the page so the client-side status poller can watch for the PAID →
	// AUTHORIZED transition without a manual refresh.
	AuthorizationToken string
}

func RenderReceiptPage(data ReceiptPageData) (string, error) {
	var buf templateBuffer
	if err := receiptTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// templateBuffer is a tiny strings.Builder wrapper so RenderReceiptPage can return
// (string, error) cleanly.
type templateBuffer struct {
	b []byte
}

func (t *templateBuffer) Write(p []byte) (int, error) {
	t.b = append(t.b, p...)
	return len(p), nil
}

func (t *templateBuffer) String() string {
	return string(t.b)
}
