package serviceHelpImpl

import (
	"encoding/json"
	"fmt"
)

// metaName is the file every documentation directory carries.
const metaName = "meta.json"

// rootsKey names the configured documentation directories.
const rootsKey = "default.runtime.services.docs"

// decodeMeta reads one meta.json.
func decodeMeta(data []byte) (meta, error) {
	parsed := meta{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return meta{}, fmt.Errorf("decode %s: %w", metaName, err)
	}
	return parsed, nil
}
