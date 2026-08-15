package sbx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

const dockerUpdateNoticeTitle = "Docker Sandboxes Update Available"

type sandboxListing struct {
	Sandboxes *[]sandboxListingEntry `json:"sandboxes"`
}

type sandboxListingEntry struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// decodeSandboxListing decodes one leading machine-readable listing and
// rejects every suffix except whitespace or Docker's known update notice.
func decodeSandboxListing(output []byte) (sandboxListing, error) {
	decoder := json.NewDecoder(bytes.NewReader(output))
	var listing sandboxListing
	if err := decoder.Decode(&listing); err != nil {
		return sandboxListing{}, err
	}

	remainder := bytes.TrimSpace(output[decoder.InputOffset():])
	if len(remainder) == 0 {
		return listing, nil
	}

	var additional json.RawMessage
	if err := json.NewDecoder(bytes.NewReader(remainder)).Decode(&additional); err == nil {
		return sandboxListing{}, errors.New("multiple JSON documents in machine-readable listing")
	}
	if !utf8.Valid(remainder) || !bytes.Contains(remainder, []byte(dockerUpdateNoticeTitle)) {
		return sandboxListing{}, fmt.Errorf("unexpected trailing output after machine-readable listing (%d bytes)", len(remainder))
	}
	return listing, nil
}
