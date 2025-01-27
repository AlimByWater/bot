package processing

import (
	"encoding/json"
	"testing"
)

func TestHelpers_ColorToHex(t *testing.T) {
	m := &Module{}
	bg := "black"

	hex := m.ColorToHex(bg)
	t.Log(hex)
}

func TestHelpers_ParseArgs(t *testing.T) {
	args := `w=[1] iphone=[true]
b=0XFFFFFF`

	m := &Module{}
	emojiArgs, err := m.ParseArgs(args)
	if err != nil {
		t.Error(err)
	}
	j, _ := json.MarshalIndent(emojiArgs, "", "  ")
	t.Log(string(j))
}
