package main

type Widget struct {
	Generator Generator
	Error     error // Only modified by generator
	//ErrorMsg  *WidgetElem // Element to be displayed when Error != nil.
}

type WidgetElem struct {
	Index int
	e     []Element
}

type WidgetError struct {
	Index int
	Error error
}

type Generator interface {
	Generate(w *Widget, index int, ctx *GeneratorCtx)
}

type GeneratorCtx struct {
	ch         chan WidgetElem
	stop, done chan bool
	errorch    chan WidgetError
}

func NewGeneratorCtx(widgetcount int) GeneratorCtx {
	return GeneratorCtx{
		ch:      make(chan WidgetElem),
		stop:    make(chan bool, widgetcount),
		done:    make(chan bool),
		errorch: make(chan WidgetError),
	}
}
