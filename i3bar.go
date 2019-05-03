package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
)

// Header is the json header used for i3-bar output
type I3BarHeader struct {
	Version     int  `json:"version"`
	StopSignal  int  `json:"stop_signal,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	ClickEvents bool `json:"click_events,omitempty"`
}

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

type AlignStr string

const (
	AlignLeft   AlignStr = "left"
	AlignRight  AlignStr = "right"
	AlignCenter AlignStr = "center"
)

type I3BarWriter interface {
	BarWriter
}

type i3Bar struct {
	I3BarWriter

	header      I3BarHeader
	out         bufio.Writer
	wroteHeader bool
}

func NewI3BarWriter(header I3BarHeader, out bufio.Writer) I3BarWriter {
	return &i3Bar{header: header, out: out}
}

func (w *i3Bar) writeHeader() (n int, err error) {
	header, err := json.Marshal(I3BarHeader{Version: 1})
	if err != nil {
		return
	}
	wrote, err := w.out.Write(header)
	if err != nil {
		return
	}
	n += wrote
	wrote, err = w.out.Write([]byte{'\n', '[', '\n', '[', ']', '\n'})
	if err != nil {
		return
	}
	n += wrote
	w.wroteHeader = true
	return
}

func (w *i3Bar) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		n, err = w.writeHeader()
		if err != nil {
			return
		}
	}

	var wrote int
	wrote, err = w.out.Write([]byte{','})
	if err != nil {
		return
	}
	n += wrote

	wrote, err = w.out.Write(p)

	if err != nil {
		return
	}
	n += wrote

	wrote, err = w.out.Write([]byte{'\n'})
	if err != nil {
		return
	}
	n += wrote

	w.out.Flush()
	return
}
