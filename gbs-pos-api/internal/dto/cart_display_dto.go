package dto

import (
	"encoding/json"
	"errors"

	"gorm.io/datatypes"
)

// DefaultCartDisplay is returned by the public GET endpoint when no payload
// has been uploaded for the requested terminal.
var DefaultCartDisplay = datatypes.JSON([]byte(`{"Initial":{"NamaKasir":"DOMAR","KodeToko":"T14AB","NamaToko":"Indomaret Pusat","JenisToko":"POINT"},"DaftarBelanja":[],"Summary":{"Hemat":"0","Total":"0","Bayar":"0","Kembali":"0"},"TeksSelesai":"None"}`))

// DeviceInfo represents device information sent from the Android POS app.
type DeviceInfo struct {
	DeviceModel        string `json:"deviceModel"`
	DeviceManufacturer string `json:"deviceManufacturer"`
	DeviceBrand        string `json:"deviceBrand"`
	AndroidVersion     string `json:"androidVersion"`
	SDKInt             int    `json:"sdkInt"`
	AppVersion         string `json:"appVersion"`
	AppVersionCode     int64  `json:"appVersionCode"`
}

// SaveCartDisplayRequest represents the JSON uploaded by the Android POS.
// Only terminalId and deviceInfo are strongly-typed fields; every other field is kept as raw JSON.
type SaveCartDisplayRequest struct {
	TerminalID string                     `json:"terminalId" binding:"required"`
	DeviceInfo *DeviceInfo                `json:"deviceInfo,omitempty"`
	Payload    map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON extracts terminalId, deviceInfo, and keeps the remaining top-level fields
// as raw JSON so they can be persisted unchanged.
func (r *SaveCartDisplayRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract terminalId
	terminalID, ok := raw["terminalId"]
	if !ok {
		return errors.New("terminalId is required!!!")
	}

	var id string
	if err := json.Unmarshal(terminalID, &id); err != nil {
		return err
	}

	if id == "" {
		return errors.New("terminalId must not be empty")
	}

	delete(raw, "terminalId")

	// Extract deviceInfo (optional)
	if deviceInfoRaw, ok := raw["deviceInfo"]; ok && deviceInfoRaw != nil {
		var di DeviceInfo
		if err := json.Unmarshal(deviceInfoRaw, &di); err == nil {
			r.DeviceInfo = &di
		}
		delete(raw, "deviceInfo")
	}

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
