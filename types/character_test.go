package types

import (
	"encoding/json"
	"testing"
)

func TestCharacterJSONOmitsEmptyGenderTags(t *testing.T) {
	// Test: marshal Character with empty Gender/Tags, assert no "gender"/"tags" keys
	c := Character{ID: "1"}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal empty: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if _, hasGender := m["gender"]; hasGender {
		t.Errorf("empty Character should not have 'gender' key")
	}
	if _, hasTags := m["tags"]; hasTags {
		t.Errorf("empty Character should not have 'tags' key")
	}

	// Test: marshal Character with Gender and Tags, assert both present
	c2 := Character{
		ID:     "2",
		Gender: "female",
		Tags:   []string{"student"},
	}
	data2, err := json.Marshal(c2)
	if err != nil {
		t.Fatalf("marshal with values: %v", err)
	}
	var m2 map[string]interface{}
	if err := json.Unmarshal(data2, &m2); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if gender, ok := m2["gender"]; !ok {
		t.Errorf("Character with Gender should have 'gender' key")
	} else if gender != "female" {
		t.Errorf("gender = %v, want 'female'", gender)
	}
	if tags, ok := m2["tags"]; !ok {
		t.Errorf("Character with Tags should have 'tags' key")
	} else if tagsSlice, ok := tags.([]interface{}); !ok {
		t.Errorf("tags is not []interface{}, got %T", tags)
	} else if len(tagsSlice) != 1 || tagsSlice[0] != "student" {
		t.Errorf("tags = %v, want ['student']", tagsSlice)
	}
}
