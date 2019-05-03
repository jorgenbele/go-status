package main

import (
	"bufio"
	"encoding/json"
)

// Header is the json header used for i3-bar output
type I3BarHeader struct {
	Version     int  `json:"version"`
	StopSignal  int  `json:"stop_signal,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	ClickEvents bool `json:"click_events,omitempty"`
}

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
