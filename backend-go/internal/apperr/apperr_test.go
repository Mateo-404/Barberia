package apperr

import "testing"

func TestNewFormatsMessageAndKeepsCode(t *testing.T) {
	err := New(KindAlreadyExists, "customer.already_exists", Params{"phone": "115"},
		"a customer with phone %s already exists", "115")
	if err.Kind != KindAlreadyExists || err.Code != "customer.already_exists" {
		t.Errorf("kind/code = %v/%q", err.Kind, err.Code)
	}
	if err.Params["phone"] != "115" {
		t.Errorf("params = %v", err.Params)
	}
	if err.Error() != "a customer with phone 115 already exists" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestKindsAreDistinct(t *testing.T) {
	seen := map[Kind]bool{}
	for _, k := range []Kind{KindBusinessRule, KindNotFound, KindAlreadyExists, KindInvalidCredentials, KindValidation, KindConflict} {
		if seen[k] {
			t.Fatalf("duplicate kind: %v", k)
		}
		seen[k] = true
	}
}
