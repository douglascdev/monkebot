package command

import "testing"

func TestParseArgs(t *testing.T) {
	if parseArgs("") != nil || parseArgs("\test") != nil || parseArgs("\test:") != nil || parseArgs("\test:hi") != nil {
		t.Fatal("failed to parse single argument to nil")
	}
}
