package main

import (
	"time"
)

type ClockGenerator struct {
	Format    string
	Alignment AlignStr
	Every     time.Duration
}

func (c ClockGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {
	gen := func() (e []Element, err error) {
		t := time.Now()
		fmt := t.Format(c.Format)
		e = append(e, Element{Name: "Clock", Alignment: c.Alignment, FullText: fmt})
		return
	}
	ticker := time.NewTicker(c.Every)
	generator(w, index, ctx, ticker.C, gen)
	ticker.Stop()
}
