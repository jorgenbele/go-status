package main

import (
	"time"
	"status1/status"
)

type ClockGen struct {
	Format    string
	Alignment status.AlignStr
	Every     time.Duration
}

func (c ClockGen) Generate(w *status.Widget, index int, ctx *status.GeneratorCtx) {
	gen := func() (e []status.Element, err error) {
		t := time.Now()
		fmt := t.Format(c.Format)
		e = append(e, status.Element{Name: "Clock", Alignment: c.Alignment, FullText: fmt})
		return
	}
	ticker := time.NewTicker(c.Every)
	status.Generatorfunc(w, index, ctx, ticker.C, gen)
	ticker.Stop()
}
