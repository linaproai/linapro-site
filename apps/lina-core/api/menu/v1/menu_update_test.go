// This file verifies the menu update HTTP contract does not publish sidebar
// icon uniqueness as a host integrity rule.

package v1

import (
	"reflect"
	"strings"
	"testing"
)

// TestUpdateReqDoesNotPublishIconUniqueness inspects the shipped PUT /menu
// documentation tags so uniqueness cannot return as a published host rule.
func TestUpdateReqDoesNotPublishIconUniqueness(t *testing.T) {
	typ := reflect.TypeOf(UpdateReq{})
	meta, ok := typ.FieldByName("Meta")
	if !ok {
		t.Fatal("UpdateReq is missing embedded Meta")
	}
	assertMenuContractAllowsDuplicateIcons(t, "g.Meta dc", meta.Tag.Get("dc"))

	icon, ok := typ.FieldByName("Icon")
	if !ok {
		t.Fatal("UpdateReq is missing Icon")
	}
	assertMenuContractAllowsDuplicateIcons(t, "icon dc", icon.Tag.Get("dc"))
}

// assertMenuContractAllowsDuplicateIcons rejects uniqueness-as-integrity wording
// on published menu contract documentation.
func assertMenuContractAllowsDuplicateIcons(t *testing.T, field string, dc string) {
	t.Helper()
	lower := strings.ToLower(dc)
	for _, token := range []string{
		"must remain globally unique",
		"duplicate icons will be rejected",
		"verified to be globally unique",
	} {
		if strings.Contains(lower, token) {
			t.Fatalf("%s still publishes icon uniqueness: %q", field, dc)
		}
	}
	if !strings.Contains(lower, "not required to be globally unique") &&
		!strings.Contains(lower, "duplicate icons are allowed") {
		t.Fatalf("%s must say duplicate icons are allowed, got %q", field, dc)
	}
}
