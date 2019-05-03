package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type BarWriter interface {
	io.Writer
	Flush() error
}

// Element contains the fields returned from the widget generators.
type Element struct {
	Name                string   `json:"name,omitempty"`
	Instance            string   `json:"instance,omitempty"`
	Alignment           AlignStr `json:"align,omitempty"`
	FullText            string   `json:"full_text,omitempty"`
	ShortText           string   `json:"short_text,omitempty"`
	Color               *Color   `json:"color,omitempty"`
	Background          *Color   `json:"background,omitempty"`
	Border              *Color   `json:"border,omitempty"`
	MinWidth            int      `json:"min_width,omitempty"`
	Urgent              bool     `json:"urgent,omitempty"`
	Separator           bool     `json:"separator,omitempty"`
	SeparatorBlockWidth int      `json:"separator_block_width,omitempty"`
}

type Status struct {
	started bool
	widgets []Widget
	cache   [][]Element
	w       BarWriter
}

func NewStatus(w BarWriter) Status {
	return Status{w: w}
}

func (s *Status) AddWidget(w Widget) {
	if s.started {
		panic(fmt.Errorf("cannot add a widget to a started status"))
	}
	s.widgets = append(s.widgets, w)
}

func (s *Status) Start() {
	if s.started {
		panic(fmt.Errorf("already started"))
	}
	s.started = true
	s.cache = make([][]Element, len(s.widgets))

	update := func() {
		v := make([]Element, 0, len(s.cache))

		for i, _ := range s.widgets {
			elems := s.cache[i]
			for _, e := range elems {
				v = append(v, e)
			}
		}
		data, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		s.w.Write(data)
	}

	ctx := NewGeneratorCtx(len(s.widgets))

	// Start goroutines.
	for i, widget := range s.widgets {
		go widget.Generator.Generate(&widget, i, &ctx)
	}

	// Register sigtermhandler
	sigtermchan := make(chan os.Signal)
	signal.Notify(sigtermchan, os.Interrupt, syscall.SIGTERM)

	// Loop.
	running := true
	for running {
		select {
		case we := <-ctx.ch:
			s.cache[we.Index] = we.e
			update()
			break

		case <-sigtermchan:
			log.Println("Catched SIGTERM, stopping!")
			running = false
			break

		case werror := <-ctx.errorch:
			log.Printf("Catched error, updating: %d, %v\n", werror.Index, werror.Error)
			red := ColorFromHex("#FF0000")
			s.cache[werror.Index] = []Element{Element{Name: "error",
				Alignment: AlignRight,
				Color:     &red,
				FullText:  fmt.Sprintf("ERROR: %v", werror.Error)}}
			update()
			break
		}
	}

}
