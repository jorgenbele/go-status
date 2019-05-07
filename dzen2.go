package main

import (
	"strings"
)

type dzen2 struct {
	Bar
	out BarWriter
}

// NewDzen2Bar implements the Bar interface and supports
// generating output for dzen2.
func NewDzen2Bar(out BarWriter) Bar {
	return &dzen2{out: out}
}

// Write ...
func (b *dzen2) Write(v []Element) (err error) {
	// TODO: Support other formatting options.
	bytes := make([]byte, 0)

	format := func(command, data []byte) (err error) {
		bytes = append(bytes, '^')
		bytes = append(bytes, command...)
		bytes = append(bytes, '(')
		bytes = append(bytes, data...)
		bytes = append(bytes, ')')
		_, err = b.out.Write(bytes)
		return
	}

	for _, e := range v {
		// TODO: Fix alignment support.

		// Colors.
		if e.Color != nil {
			err = format([]byte("fg"), []byte(e.Color.String()))
			if err != nil {
				return
			}
		}
		if e.Background != nil {
			err = format([]byte("bg"), []byte(e.Color.String()))
			if err != nil {
				return
			}
		}

		// Contents.
		str := strings.Replace(e.FullText, "^", "^^", -1)
		bytes = append(bytes, []byte(str)...)
	}
	bytes = append(bytes, '\n')
	_, err = b.out.Write(bytes)
	b.out.Flush()
	return
}
