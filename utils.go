package main

import (
	"fmt"
	"strconv"
	"time"
)

// HBar generates a horizontal progressbar
func HBar(progress, size int, prune, other rune) string {
	v := make([]rune, 0, size)
	for i := 0; i < progress; i++ {
		v = append(v, prune)
	}
	for i := progress; i < size; i++ {
		v = append(v, other)
	}
	return string(v)
}

// ColorFromHex converts a hex color string (#RRGGBB) to a Color struct
func ColorFromHex(hex string) (c Color) {
	if len(hex) != 7 {
		panic(fmt.Sprintf("%s is not a valid hex color: invalid length %d",
			hex, len(hex)))
	}

	var rgb [3]uint8
	for i := 0; i < cap(rgb); i++ {
		c, err := strconv.ParseUint(hex[2*i+1:2*i+3], 16, 8)
		if err != nil {
			nerr := fmt.Errorf("%s is not a valid hex color: %v", hex, err)
			panic(nerr)
		}
		rgb[i] = uint8(c)
	}
	return Color{rgb[0], rgb[1], rgb[2]}
}

// Calls gen() every tick (timeout) until <-stop. On error the Error field
// of the widget is set and the goroutine signifies it is 'done' and returns.
func generator(w *Widget, index int, ctx *GeneratorCtx,
	tick <-chan time.Time, gen func() ([]Element, error)) {

	prod, err := gen()
	if err != nil {
		w.Error = err
		ctx.errorch <- WidgetError{index, err}
		ctx.done <- true
		return
	}
	ctx.ch <- WidgetElem{index, prod}

	for {
		select {
		case <-ctx.stop:
			ctx.done <- true
			return

		case _, ok := <-tick:
			if !ok {
				ctx.done <- true
				return
			}
			break
		}

		prod, err := gen()
		if err != nil {
			w.Error = err
			ctx.errorch <- WidgetError{index, err}
			ctx.done <- true
			return
		}
		ctx.ch <- WidgetElem{index, prod}
	}
}
