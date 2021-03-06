package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"github.com/jorgenbele/go-status/status"
)

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

	// Parse a list of ranges: x-y,z-w,...
	s := strings.TrimSpace(string(data))
	ranges := strings.Split(s, ",")
	for _, r := range ranges {
		rsplit := strings.Split(r, "-")
		var left, right uint64

		if len(rsplit) != 2 {
			err = fmt.Errorf("invalid range: %v, expected 2 elements got %d",
				rsplit, len(rsplit))
			return
		}

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

func (c CPU) Color() status.Color {
	return status.ColorFromHex("#8A8B8C")
}

func (c CPU) Symbol() string {
	size := 5
	return status.HBar(int(c.Usage/float64(c.Cores)*float64(size)), size, '+', '-')
}

// CPUGen gets the CPU utilization by reading the /proc/loadavg file.
type CPUGen struct {
	Alignment status.AlignStr
	Every     time.Duration
}

// Generate ...
func (c CPUGen) Generate(w *status.Widget, index int, ctx *status.GeneratorCtx) {
	gen := func() (e []status.Element, err error) {
		cpu, err := CPUInfo()
		if err != nil {
			return
		}
		color := cpu.Color()
		e = append(e, status.Element{Name: "CPU", Alignment: c.Alignment, Color: &color,
			FullText: fmt.Sprintf("%d%% %s", cpu.UsagePerc(), cpu.Symbol())})
		return
	}
	ticker := time.NewTicker(c.Every)
	status.Generatorfunc(w, index, ctx, ticker.C, gen)
	ticker.Stop()
}
