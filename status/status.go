package status

import (
	"fmt"
	"io"
	"log"
	"os"
)

// BarWriter is an interface wrapping an io.Writer with
// an additional Flush() function to ensure that the bar is
// updated on write. (can use bufio.Writer).
type BarWriter interface {
	io.Writer
	Flush() error
}

// Bar ...
type Bar interface {
	Write(e []Element) error
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

// AlignStr represents the various ways to aligning widgets.
type AlignStr string

const (
	AlignNone   AlignStr = ""
	AlignLeft   AlignStr = "left"
	AlignRight  AlignStr = "right"
	AlignCenter AlignStr = "center"
)

// Status is used to generate the statusline from a set of widgets.
type Status struct {
	started bool
	widgets []Widget
	cache   [][]Element
	b       Bar

	sigstopch <-chan os.Signal
	sigcontch <-chan os.Signal
	sigtermch <-chan os.Signal
}

// NewStatus creates a new status.
func NewStatus(b Bar) Status {
	return Status{b: b}
}

// AddWidget adds the given widget to the slice of widgets to be
// displayed on the statusline. The order of AddWidget calls is the
// order which is used for the statusline. (May differ based upon
// alignment.)
func (s *Status) AddWidget(w Widget) {
	if s.started {
		panic(fmt.Errorf("cannot add a widget to a started status"))
	}
	s.widgets = append(s.widgets, w)
}

// SetStopSignal sets stop signal, usually this would be SIGSTOP.
func (s *Status) SetStopSignal(c <-chan os.Signal) {
	if s.started {
		panic(fmt.Errorf("cannot set stop signal since status already started"))
	}
	s.sigstopch = c
}

// SetContSignal sets cont signal, usually this would be SIGCONT.
func (s *Status) SetContSignal(c <-chan os.Signal) {
	if s.started {
		panic(fmt.Errorf("cannot set cont signal since status already started"))
	}
	s.sigcontch = c
}

// SetTermSignal sets sigterm signal, usually this would be SIGTERM.
func (s *Status) SetTermSignal(c <-chan os.Signal) {
	if s.started {
		panic(fmt.Errorf("cannot set sigterm signal since status already started"))
	}
	s.sigtermch = c
}

// Start starts the status loop which will run until the
// term signal is recieved.
func (s *Status) Start() {
	if s.started {
		panic(fmt.Errorf("attempting to start an already started status"))
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

		//data, err := json.Marshal(v)
		//if err != nil {
		//	panic(err)
		//}
		//s.w.Write(data)
		err := s.b.Write(v)
		if err != nil {
			panic(err)
		}
	}

	ctx := NewGeneratorCtx(len(s.widgets))

	// Start goroutines.
	for i, widget := range s.widgets {
		go widget.Gen.Generate(&widget, i, &ctx)
	}

	// Loop until a term signal is recieved.
	running := true
	for running {
		select {
		case we := <-ctx.Ch:
			s.cache[we.Index] = we.e
			update()
			break

		case <-s.sigstopch:
			log.Println("Recieved stop signal, stopping!")

			// NOTE: Does not stop the running processes, but
			// ignores the we and werror channel until sigcont or
			// sigterm.
			select {
			case <-s.sigcontch:
				log.Println("Recieved cont signal while stopped, continuing!")
				break

			case <-s.sigstopch:
				log.Println("Recieved stop signal while stopped, ignoring!")
				break

			case <-s.sigtermch:
				running = false
				log.Println("Recieved term signal while stopped, shutting down!")
				break
			}
			break

		case <-s.sigcontch:
			log.Println("Recieved cont signal while running, ignoring!")
			break

		case <-s.sigtermch:
			log.Println("Recieved term signal, shutting down!")
			running = false
			break

		case werror := <-ctx.Errorch:
			log.Printf("Recieved widget error, updating: %d, %v\n", werror.Index, werror.Error)
			red := ColorFromHex("#FF0000")
			s.cache[werror.Index] = []Element{Element{Name: "error",
				Alignment: AlignRight,
				Color:     &red,
				FullText:  fmt.Sprintf("ERROR: %v", werror.Error)}}
			update()
			break
		}
	}

	// Stop widget generators.
	log.Println("Sending stop to all widgets.")
	for i := 0; i < len(s.widgets); i++ {
		ctx.Stop <- true
	}

	// Wait for all widgets to send done.
	// Consume products while waiting since some widgets
	// may still have uncomsumed channel messages.
	log.Println("Waiting for done messages.")
	remaining := len(s.widgets)
	for remaining != 0 {
		select {
		case <-ctx.Done:
			remaining--
			break
		case <-ctx.Ch:
			break
		case <-ctx.Errorch:
			break
		}
	}

	log.Println("Stopped all widgets. Shutting down.")
}
