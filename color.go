package main

import (
	"bytes"
	"fmt"
)

// Color represents a rgb color
type Color struct {
	R uint8
	G uint8
	B uint8
}

// HexStr converts a color struct to a hex string repr.
func (c Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func (c Color) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString("\"")
	b.WriteString(c.String())
	b.WriteString("\"")
	return b.Bytes(), nil
}

func (c *Color) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) != 9 {
		return fmt.Errorf("invalid hex color: %s", s)
	}
	*c = ColorFromHex(string(data)[1:8])
	return nil
}
