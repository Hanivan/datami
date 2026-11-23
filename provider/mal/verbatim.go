package mal

import (
	"encoding/json"
)

// verbatimJSON re-marshals a struct into json.RawMessage so the upsert can
// persist the exact upstream shape. Returns nil when the struct holds no real
// data so the persisted column stays SQL NULL.
func verbatimJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil
	}
	if jsonLeafEmpty(decoded) {
		return nil
	}
	return json.RawMessage(b)
}

func jsonLeafEmpty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case bool:
		return !x
	case float64:
		return x == 0
	case []any:
		for _, e := range x {
			if !jsonLeafEmpty(e) {
				return false
			}
		}
		return true
	case map[string]any:
		for _, e := range x {
			if !jsonLeafEmpty(e) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func verbatimPictures(p []picture) json.RawMessage {
	if len(p) == 0 {
		return nil
	}
	return verbatimJSON(p)
}

func verbatimAltTitles(alt altTitles) json.RawMessage {
	if alt.Synonyms == nil {
		alt.Synonyms = []string{}
	}
	return verbatimJSON(alt)
}
