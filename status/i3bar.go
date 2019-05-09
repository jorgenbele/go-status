package status

import (
	"encoding/json"
)

// Header is the json header used for i3-bar output
type I3BarHeader struct {
	Version     int  `json:"version"`
	StopSignal  int  `json:"stop_signal,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	ClickEvents bool `json:"click_events,omitempty"`
}

type i3Bar struct {
	Bar

	header      I3BarHeader
	out         BarWriter
	wroteHeader bool
}

// NewI3BarWriter creates a BarWriter which outputs in i3bar format.
func NewI3Bar(header I3BarHeader, out BarWriter) Bar {
	return &i3Bar{header: header, out: out}
}

func (w *i3Bar) writeHeader() (n int, err error) {
	bytes := make([]byte, 0)

	header, err := json.Marshal(I3BarHeader{Version: 1})
	if err != nil {
		return
	}
	bytes = append(bytes, header...)
	bytes = append(bytes, []byte{'\n', '[', '\n', '[', ']', '\n'}...)
	w.wroteHeader = true
	return w.out.Write(bytes)
}

func (w *i3Bar) Write(v []Element) (err error) {
	bytes := make([]byte, 0)

	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	if !w.wroteHeader {
		_, err = w.writeHeader()
		if err != nil {
			return
		}
	}

	bytes = append(bytes, ',')
	bytes = append(bytes, data...)
	bytes = append(bytes, '\n')

	_, err = w.out.Write(bytes)
	w.out.Flush()
	return
}
