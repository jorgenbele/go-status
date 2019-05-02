package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"syscall"
	"time"
)

// Header is the json header used for i3-bar output
type Header struct {
	Version     int  `json:"version"`
	StopSignal  int  `json:"stop_signal,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	ClickEvents bool `json:"click_events,omitempty"`
}

// Color represents a rgb color
type Color struct {
	R uint8
	G uint8
	B uint8
}

// HexStr converts a color struct to a hex string repr.
func (c Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

func (c Color) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString("\"")
	b.WriteString(c.String())
	b.WriteString("\"")
	return b.Bytes(), nil
}

func (c *Color) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) != 9 {
		return fmt.Errorf("invalid hex color: %s", s)
	}
	*c = ColorFromHex(string(data)[1:8])
	return nil
}

// AlignStr represents an alignment
type AlignStr string

const (
	AlignLeft   AlignStr = "left"
	AlignRight  AlignStr = "right"
	AlignCenter AlignStr = "center"
)

// nenerator interface used for generating widget strings
type Generator interface {
	Generate(w *Widget, index int, ch chan WidgetElem, stop, done chan bool)
}

// Element contains the fields returned from the widget generator
// and is passed to i3bar/swaybar.
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

// Widget contains all information about a widget
type Widget struct {
	Generator Generator
	Error     error // Only modified by generator
}

type WidgetElem struct {
	Index int
	e     []Element
}

// ColorFromHex converts a hex color string (#RRGGBB) to a Color struct
func ColorFromHex(hex string) (c Color) {
	if len(hex) != 7 {
		panic(fmt.Sprintf("%s is not a valid hex color: invalid length %d", hex, len(hex)))
	}

	var rgb [3]uint8
	for i := 0; i < cap(rgb); i++ {
		c, err := strconv.ParseUint(hex[2*i+1:2*i+3], 16, 8)
		rgb[i] = uint8(c)
		if err != nil {
			panic(fmt.Sprintf("%s is not a valid hex color: invalid hex digits", hex))
		}
	}
	return Color{rgb[0], rgb[1], rgb[2]}
}

const (
	// PowerSupplyPath ...
	PowerSupplyPath = "/sys/class/power_supply"
	// BatPrefix is the prefix string which identifies
	// power suppliers which are batteries
	BatPrefix = "BAT"
)

// BatStatus is used to represent the battery status
type BatStatus int

// BatCharge represents the battery charge level as
// a percentage of maximum charge.
type BatCharge int

const (
	// BatUnknown is used when battery is unknown
	BatUnknown BatStatus = iota
	// BatDischarging is used when battery is discharging
	BatDischarging
	// BatCharging is used when battery is charging
	BatCharging
	// BatFull is used when battery is full
	BatFull
)

func batteryStatus(name string) (status BatStatus, charge BatCharge, err error) {
	statusMap := map[string]BatStatus{
		"Unknown\n":     BatUnknown,
		"Charging\n":    BatCharging,
		"Discharging\n": BatDischarging,
		"Full\n":        BatFull,
	}

	batPath := fmt.Sprintf("%s/%s/", PowerSupplyPath, name)

	data, err := ioutil.ReadFile(batPath + "capacity")
	if err != nil {
		return
	}
	chargeInt, err := strconv.Atoi(strings.Trim(string(data), "\n"))
	if err != nil {
		return
	}
	charge = BatCharge(chargeInt)

	data, err = ioutil.ReadFile(batPath + "status")
	if err != nil {
		return
	}
	status, ok := statusMap[string(data)]
	if !ok {
		err = fmt.Errorf("Unknown battery status: %s", string(data))
		return
	}
	return
}

// Battery represents the status of a battery
type Battery struct {
	Path   string
	Status BatStatus
	Charge BatCharge
}

func GetColors() []Color {
	colors := []Color{
		ColorFromHex("#B82E34"), // very low
		ColorFromHex("#B82E34"), // low
		ColorFromHex("#8A8B8C"), // medium
		ColorFromHex("#8A8B8C"), // near full
		ColorFromHex("#8A8B8C"), // full
	}
	return colors
}

// GetVBar returns a vertical bar
func GetVBar() []string {
	vbar := []string{
		"_",
		"▂",
		"▄",
		"▆",
		"█",
	}
	return vbar
}

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

// Color returns a suitable color for the given battery capacity/state.
func (b Battery) Color() Color {
	colors := GetColors()
	return colors[int(b.Charge)/100.0*(len(colors)-1)]
}

// Symbol returns a suitable symbol for the given battery capacity/state.
func (b Battery) Symbol() string {
	prefix := map[BatStatus]string{
		BatUnknown:     "",
		BatCharging:    "+",
		BatDischarging: "-",
		BatFull:        "",
	}

	vbar := GetVBar()
	symb := vbar[int(float64(b.Charge)/100.0*float64(len(vbar)-1))]
	return fmt.Sprintf("%s%s", prefix[b.Status], symb)
}

// BatteryInfo returns a slice of batteries with name, status and charge.
func BatteryInfo() ([]Battery, error) {
	var bats []Battery

	files, err := ioutil.ReadDir(PowerSupplyPath)
	if err != nil {
		return bats, err
	}

	for _, file := range files {
		name := file.Name()
		if !strings.HasPrefix(name, BatPrefix) {
			continue
		}

		s, c, err := batteryStatus(name)
		if err != nil {
			return bats, err
		}

		bats = append(bats, Battery{name, s, c})
	}
	return bats, nil
}

// LoadAvg reads /proc/loadavg and returns it as a string slice
func LoadAvg() ([]string, error) {
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(strings.TrimSpace(string(data)), " "), nil
}

// CPU represents CPU usage.
type CPU struct {
	// CPU Usage as a fraction.
	Usage float64
	Cores uint
}

func cores() (n uint, err error) {
	data, err := ioutil.ReadFile("/sys/devices/system/cpu/present")
	if err != nil {
		return
	}

	// parse a list of ranges: x-y,z-w,...
	s := strings.TrimSpace(string(data))
	ranges := strings.Split(s, ",")
	for _, r := range ranges {
		rsplit := strings.Split(r, "-")
		var left, right uint64

		// assume length is 2
		left, err = strconv.ParseUint(rsplit[0], 10, 32)
		if err != nil {
			return
		}
		right, err = strconv.ParseUint(rsplit[1], 10, 32)
		if err != nil {
			return
		}

		n += uint((right - left) + 1)
	}
	return
}

// CPUInfo represents the CPU usage status.
func CPUInfo() (cpu CPU, err error) {
	l, err := LoadAvg()
	if err != nil {
		return
	}

	cpu.Usage, err = strconv.ParseFloat(l[0], 64)
	if err != nil {
		return
	}
	cpu.Cores, err = cores()
	return // includes err
}

func (c CPU) UsagePerc() int {
	return int(c.Usage / float64(c.Cores) * 100.0)
}

func (c CPU) Color() Color {
	colors := GetColors()
	return colors[c.UsagePerc()/100.0*(len(colors)-1)]
}

func (c CPU) Symbol() string {
	size := 6
	return HBar(int(c.Usage/float64(c.Cores)*float64(size)), size, '+', '-')
}

// Generate ...
type CPUGenerator struct{}

func (c CPUGenerator) Generate(w *Widget, index int, ch chan WidgetElem, stop, done chan bool) {
	gen := func() (e []Element, err error) {
		c, err := CPUInfo()
		if err != nil {
			return
		}
		color := c.Color()
		e = append(e, Element{Name: "CPU", Alignment: AlignRight, Color: &color, FullText: fmt.Sprintf("%d%% %s", c.UsagePerc(), c.Symbol())})
		return
	}
	generator(w, index, ch, stop, done, 3*time.Second, gen)
}

// BatteryGenerator Used to implement Generate()
type BatteryGenerator struct{}

// Call gen() every tick (timeout). On error the Error field of the widget is set and
// the goroutine signifies it is 'done' and returns.
func generator(w *Widget, index int, ch chan WidgetElem, stop, done chan bool, timeout time.Duration, gen func() ([]Element, error)) {
	prod, err := gen()
	if err != nil {
		w.Error = err
		done <- true
		return
	}
	ch <- WidgetElem{index, prod}

	for {
		tick := time.Tick(timeout)

		select {
		case <-stop:
			done <- true
			return

		case <-tick:
			break
		}

		prod, err := gen()
		if err != nil {
			w.Error = err
			done <- true
			return
		}
		ch <- WidgetElem{index, prod}
	}
}

// Generate ...
func (b BatteryGenerator) Generate(w *Widget, index int, ch chan WidgetElem, stop, done chan bool) {
	gen := func() (e []Element, err error) {
		bats, err := BatteryInfo()
		if err != nil {
			return
		}
		for _, b := range bats {
			color := b.Color()
			e = append(e, Element{Name: "Battery", Instance: b.Path, Alignment: AlignRight, Color: &color, FullText: fmt.Sprintf("%d%% %s", b.Charge, b.Symbol())})
		}
		return
	}
	generator(w, index, ch, stop, done, 2*time.Second, gen)
	return
}

type ClockGenerator struct{}

func (c ClockGenerator) Generate(w *Widget, index int, ch chan WidgetElem, stop, done chan bool) {
	gen := func() (e []Element, err error) {
		t := time.Now()
		fmt := t.Format("Mon Jan 2 15:04:05")

		e = append(e, Element{Name: "Clock", Alignment: AlignRight, FullText: fmt})
		return
	}
	generator(w, index, ch, stop, done, time.Second, gen)
}

type CommandGenerator struct {
	Tick       time.Duration
	Instance   string
	CmdCreator func() *exec.Cmd
	IsJSON     bool
	TrimSpace  bool
}

// Generate ...
func (c CommandGenerator) Generate(w *Widget, index int, ch chan WidgetElem, stop, done chan bool) {
	gen := func() (e []Element, err error) {
		cmd := c.CmdCreator()
		outbytes, err := cmd.Output()
		if err != nil {
			w.Error = err
			fmt.Fprintf(os.Stderr, "Command failed for widget #%d: %s\n", index, err)
			e = append(e, Element{Name: "Command", Instance: c.Instance, Alignment: AlignRight, FullText: fmt.Sprintf("ERROR: %s", err)})
			return
		}

		if !c.IsJSON {
			var out string
			if c.TrimSpace {
				out = strings.TrimSpace(string(outbytes))
			} else {
				out = string(outbytes)
			}
			e = append(e, Element{Name: "Command", Instance: c.Instance, Alignment: AlignRight, FullText: string(out)})
		} else {
			var elem Element
			err = json.Unmarshal(outbytes, &elem)
			if err != nil {
				w.Error = err
				e = append(e, Element{Name: "Command", Instance: c.Instance, Alignment: AlignRight, FullText: fmt.Sprintf("ERROR: %s", err)})
				return
			}
			e = append(e, elem)
		}
		return
	}
	generator(w, index, ch, stop, done, c.Tick, gen)
}

func main() {
	// TODO support SIGSTOP/SIGCONT
	// Catch SIGTERM
	sigtermchan := make(chan os.Signal)
	signal.Notify(sigtermchan, os.Interrupt, syscall.SIGTERM)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	// Start JSON output, beginning with header and followed with an
	// opening bracket (array).
	header, err := json.Marshal(Header{Version: 1})
	if err != nil {
		panic(err)
	}
	out.Write(header)
	out.Write([]byte{'\n', '[', '\n', '[', ']', '\n'})

	widgets := []Widget{
		//Widget{Generator: CommandGenerator{Instance: "mpc", Tick: time.Second * 10, IsJSON: true, CmdCreator: func() *exec.Cmd { return exec.Command("mpc", "-h", "localhost", "--format", "%title% - %album%,%artist", "current") }}},
		Widget{Generator: CommandGenerator{Instance: "wireless", Tick: time.Second * 10, IsJSON: true, CmdCreator: func() *exec.Cmd { return exec.Command("wireless_con") }}},
		Widget{Generator: CommandGenerator{Instance: "mullvadvpn", Tick: time.Second * 10, IsJSON: true, CmdCreator: func() *exec.Cmd { return exec.Command("mullvad_jsonblock") }}},
		Widget{Generator: BatteryGenerator{}},
		Widget{Generator: CPUGenerator{}},
		Widget{Generator: ClockGenerator{}},
		Widget{Generator: CommandGenerator{Instance: "sleeptest", Tick: time.Second, CmdCreator: func() *exec.Cmd { return exec.Command("sleeptest") }}},
		//Widget{Generator: CommandGenerator{Instance: "uptime", Tick: time.Second * 10, TrimSpace: true, CmdCreator: func() *exec.Cmd { return exec.Command("uptime", "-p") }}},
	}
	cache := make([][]Element, len(widgets))

	update := func() {
		v := make([]Element, 0, len(cache))

		for i, _ := range widgets {
			elems := cache[i]
			for _, e := range elems {
				v = append(v, e)
			}
		}

		data, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		out.Write([]byte{','})
		out.Write(data)
		out.Flush()
		out.Write([]byte{'\n'})
	}

	wechan := make(chan WidgetElem)
	stop := make(chan bool, len(widgets))
	done := make(chan bool)

	// Start widget generators.
	for i, widget := range widgets {
		go widget.Generator.Generate(&widget, i, wechan, stop, done)
	}

	// Main loop
	running := true
	for running {
		select {
		case we := <-wechan:
			//fmt.Fprintln(os.Stderr, "Recieved on wechan!", we, we.Index)
			cache[we.Index] = we.e
			update()
			break

		case <-sigtermchan:
			fmt.Fprintln(os.Stderr, "Catched SIGTERM, stopping!")
			running = false
			break
		}
	}

	// Stop widget generators.
	fmt.Fprintln(os.Stderr, "=== Sending stop")
	for i := 0; i < len(widgets); i++ {
		stop <- true
	}

	fmt.Fprintln(os.Stderr, "=== Getting remaining")
	remaining := true
	for remaining {
		select {
		case <-wechan:
			break
		default:
			remaining = false
		}
	}

	// Wait for all goroutines to complete.
	fmt.Fprintln(os.Stderr, "=== Waiting for done messages")
	for i := 0; i < len(widgets); i++ {
		select {
		case <-done:
			break
		}
	}

	fmt.Fprintln(os.Stderr, "Stopped all goroutines. Shutting down.")
}
