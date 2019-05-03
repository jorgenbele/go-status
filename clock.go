package main

import (
	"time"
)

type ClockGenerator struct{}

func (c ClockGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {
	gen := func() (e []Element, err error) {
		t := time.Now()
		fmt := t.Format("Mon Jan 2 15:04:05")
		e = append(e, Element{Name: "Clock", Alignment: AlignRight, FullText: fmt})
		return
	}
	ticker := time.NewTicker(time.Second)
	generator(w, index, ctx, ticker.C, gen)
	ticker.Stop()
}
