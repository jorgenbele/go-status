package status

// Widget contains information about a generator instance.
type Widget struct {
	Gen   Generator
	Error error // Only modified by generator
}

// WidgetElem is returned from the generators to the
// status instance through a channel.
type WidgetElem struct {
	Index int
	e     []Element
}

// WidgetError is passed to the status instance when
// a generator encounters an unrecoverable error.
type WidgetError struct {
	Index int
	Error error
}

// Generator is the interface all generators must implement.
type Generator interface {
	Generate(w *Widget, index int, ctx *GeneratorCtx)
}

// GeneratorCtx is used to
type GeneratorCtx struct {
	Ch         chan WidgetElem
	Stop, Done chan bool
	Errorch    chan WidgetError
}

// NewGeneratorCtx ...
func NewGeneratorCtx(widgetcount int) GeneratorCtx {
	return GeneratorCtx{
		Ch:      make(chan WidgetElem),
		Stop:    make(chan bool, widgetcount),
		Done:    make(chan bool),
		Errorch: make(chan WidgetError, widgetcount),
	}
}
