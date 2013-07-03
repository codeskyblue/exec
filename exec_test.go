package exec

import "testing"

func TestOK(t *testing.T) {
	if 1 != 1 {
		t.Error("my god")
	}
}
