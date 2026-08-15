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

	// Decoding silently rewrites invalid UTF-8 inside quoted strings to
	// U+FFFD, so a malformed name can match a configured name and a
	// malformed status can misreport the lifecycle state.
	document := output[:decoder.InputOffset()]
	if !utf8.Valid(document) {
		return sandboxListing{}, errors.New("invalid UTF-8 in machine-readable listing")
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
