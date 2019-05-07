package main

import (
	"strings"

	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"time"

	"log"
)

type CommandGenerator struct {
	C          <-chan time.Time
	Instance   string
	CmdCreator func() *exec.Cmd
	IsJSON     bool
	IsArray    bool // used with IsJSON

	TrimSpace bool
}

// Generate ...
func (c CommandGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {

	fail := func(err error) {
		//w.Error = err
		log.Printf("Command failed for widget #%d: %s\n", index, err)
		return
	}

	gen := func() (e []Element, err error) {
		cmd := c.CmdCreator()
		outbytes, err := cmd.Output()
		if err != nil {
			fail(err)
			return
		}

		if !c.IsJSON {
			var out string
			if c.TrimSpace {
				out = strings.TrimSpace(string(outbytes))
			} else {
				out = string(outbytes)
			}
			e = append(e,
				Element{Name: "Command",
					Instance:  c.Instance,
					Alignment: AlignRight,
					FullText:  string(out)})
		} else if !c.IsArray {
			var elem Element
			err = json.Unmarshal(outbytes, &elem)
			if err != nil {
				fail(err)
				return
			}
			e = append(e, elem)
		} else {
			var elems []Element
			err = json.Unmarshal(outbytes, &elems)
			if err != nil {
				fail(err)
				return
			}

			for _, elem := range elems {
				e = append(e, elem)
			}
		}
		return
	}
	generator(w, index, ctx, c.C, gen)
}

// NOTE: Only JSON
type StreamingCommandGenerator struct {
	Instance   string
	Restart    bool
	CmdCreator func() *exec.Cmd
}

// Reads a stream where each line is a JSON encoded Element.
func (c StreamingCommandGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {

	failnow := func(err error) {
		log.Printf("Command failed for widget #%d: %s\n", index, err)
		w.Error = err
		ctx.errorch <- WidgetError{index, err}
		ctx.done <- true
		return
	}

	cmd := c.CmdCreator()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		failnow(err)
		return
	}
	err = cmd.Start()

	// Channel for lines read from stdin
	stdoutch := make(chan []byte, 1)
	tickerch := make(chan time.Time, 1)

	go func(ch chan []byte, ticker chan time.Time, f io.ReadCloser) {
		r := bufio.NewReader(f)

		for {
			bytes, err := r.ReadBytes('\n')
			if err != nil {
				failnow(err)
				close(ticker)
				close(ch)
				return
			}
			ticker <- time.Now()
			ch <- bytes
		}
	}(stdoutch, tickerch, stdout)

	gen := func() (e []Element, err error) {
		var elem Element
		err = json.Unmarshal(<-stdoutch, &elem)
		if err != nil {
			//failnow(err)
			return
		}
		e = append(e, elem)
		return
	}
	generator(w, index, ctx, tickerch, gen)
	//cmd.Wait()
}
