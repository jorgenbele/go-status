package status

import (
	"fmt"
	"strings"
)

type lemonbar struct {
	Bar
	out BarWriter
}

// NewLemonbar implements the Bar interface and supports
// generating output for lemonbar.
func NewLemonbar(out BarWriter) Bar {
	return &lemonbar{out: out}
}

// Write ...
func (b *lemonbar) Write(v []Element) (err error) {
	bytes := make([]byte, 0)

	format := func(prefix byte, data []byte) (err error) {
		bytes = append(bytes, []byte{'%', '{', prefix}...)
		bytes = append(bytes, data...)
		bytes = append(bytes, '}')
		_, err = b.out.Write(bytes)
		return
	}

	var curalign AlignStr
	for _, e := range v {
		// Only output alignment formatting for every change, since
		// lemonbar wil overlap (overwrite) widgets otherwise.
		if e.Alignment != curalign {
			curalign = e.Alignment
			alignMap := map[AlignStr][]byte{
				AlignLeft:   []byte("%{l}"),
				AlignRight:  []byte("%{r}"),
				AlignCenter: []byte("%{c}"),
			}

			// Alignment.
			if e.Alignment != AlignNone {
				var alignbytes []byte
				alignbytes, ok := alignMap[e.Alignment]
				if !ok {
					return fmt.Errorf("invalid alignment")
				}
				bytes = append(bytes, alignbytes...)
			}
		}

		// Colors.
		if e.Color != nil {
			format('F', []byte(e.Color.String()))
		}
		if e.Background != nil {
			format('B', []byte(e.Background.String()))
		}

		// Contents.
		str := strings.Replace(e.FullText, "%", "%%", -1)
		bytes = append(bytes, []byte(str)...)
	}
	bytes = append(bytes, '\n')
	_, err = b.out.Write(bytes)

	// TODO: Support other formatting options.
	b.out.Flush()
	return
}
