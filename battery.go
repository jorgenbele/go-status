package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	PowerSupplyPath = "/sys/class/power_supply"
	BatPrefix       = "BAT"
)

// BatStatus is used to represent the battery status
type BatStatus int

// BatCharge represents the charge level as a percentage of maximum charge.
type BatCharge int

const (
	BatUnknown BatStatus = iota
	BatDischarging
	BatCharging
	BatFull
)

var batBar [5]string
var batColors [5]Color
var batPrefix map[BatStatus]string
var batStatus map[string]BatStatus

func init() {
	batBar = [...]string{
		"",
		"",
		"",
		"",
		"",
	}

	batColors = [...]Color{
		ColorFromHex("#B82E34"), // very low
		ColorFromHex("#B82E34"), // low
		ColorFromHex("#8A8B8C"), // medium
		ColorFromHex("#8A8B8C"), // near full
		ColorFromHex("#8A8B8C"), // full
	}

	batPrefix = map[BatStatus]string{
		BatUnknown:     "",
		BatCharging:    " ",
		BatDischarging: "",
		BatFull:        "",
	}

	batStatus = map[string]BatStatus{
		"Unknown\n":     BatUnknown,
		"Charging\n":    BatCharging,
		"Discharging\n": BatDischarging,
		"Full\n":        BatFull,
	}
}

func batteryStatus(name string) (status BatStatus, charge BatCharge, err error) {
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
	status, ok := batStatus[string(data)]
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

// Color returns a suitable color for the given battery capacity/state.
func (b Battery) Color() Color {
	return batColors[int(float64(b.Charge)/100.0*float64(len(batColors)-1))]
}

// Symbol returns a suitable symbol for the given battery capacity/state.
func (b Battery) Symbol() string {
	symb := batBar[int(float64(b.Charge)/100.0*float64(len(batBar)-1))]
	return fmt.Sprintf("%s%s", batPrefix[b.Status], symb)
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

type BatteryGenerator struct {
	Alignment AlignStr
	Every     time.Duration
}

func (b BatteryGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {

	gen := func() (e []Element, err error) {
		bats, err := BatteryInfo()
		if err != nil {
			return
		}
		for _, bat := range bats {
			color := bat.Color()
			e = append(e, Element{Name: "Battery", Instance: bat.Path,
				Alignment: b.Alignment, Color: &color,
				FullText: fmt.Sprintf("%d%% %s", bat.Charge, bat.Symbol())})
		}
		return
	}

	ticker := time.NewTicker(b.Every)
	generator(w, index, ctx, ticker.C, gen)
	ticker.Stop()
	return
}
