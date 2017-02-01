package apiutils

import "testing"

// Test the getVersion() function.
func TestGetVersion(t *testing.T) {
	v := "0.0.0"
	maj, min, patch := getVersion(v)

	if maj != 0 && min != 0 && patch != 0 {
		t.Error("maj, min or patch are not set to 0", maj, min, patch)
	}

	v = "1.2.4"

	maj, min, patch = getVersion(v)

	if maj != 1 && min != 2 && patch != 4 {
		t.Error("maj, min or patch are not set to 1, 2, 4", maj, min, patch)
	}
}
