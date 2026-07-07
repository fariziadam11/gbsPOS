package dto

import (
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
)

// DefaultCartDisplay is returned by the public GET endpoint when no payload
// has been uploaded for the requested terminal.
var DefaultCartDisplay = datatypes.JSON([]byte(`{"Initial":{"NamaKasir":"DOMAR","KodeToko":"T14AB","NamaToko":"Indomaret Pusat","JenisToko":"POINT"},"DaftarBelanja":[],"Summary":{"Hemat":"0","Total":"0","Bayar":"0","Kembali":"0"},"TeksSelesai":"None"}`))

// SaveCartDisplayRequest represents the JSON uploaded by the Android POS.
// Only terminalId is a strongly-typed field; every other field is kept as raw JSON.
type SaveCartDisplayRequest struct {
	TerminalID string                     `json:"terminalId" binding:"required"`
	Payload    map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON extracts terminalId and keeps the remaining top-level fields
// as raw JSON so they can be persisted unchanged.
func (r *SaveCartDisplayRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	terminalID, ok := raw["terminalId"]
	if !ok {
		return errors.New("terminalId is required")
	}

	var id string
	if err := json.Unmarshal(terminalID, &id); err != nil {
		return err
	}

	if id == "" {
		return errors.New("terminalId must not be empty")
	}

	delete(raw, "terminalId")

	r.TerminalID = id
	r.Payload = raw
	return nil
}

// MarshalPayload serializes the raw payload fields into a JSON document.
func (r *SaveCartDisplayRequest) MarshalPayload() (datatypes.JSON, error) {
	if r.Payload == nil {
		r.Payload = map[string]json.RawMessage{}
	}

	data, err := json.Marshal(r.Payload)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(data), nil
}
