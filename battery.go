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

// batBar is initialized in init()
var batBar []string

func init() {
	batBar = []string{
		"",
		"",
		"",
		"",
		""}
}

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
		//ColorFromHex("#FF0000"), // very low
		//ColorFromHex("#AA0000"), // low
		//ColorFromHex("#FFFFFF"), // medium
		//ColorFromHex("#00AA00"), // near full
		//ColorFromHex("#00FF00"), // full
	}
	return colors
}

// Color returns a suitable color for the given battery capacity/state.
func (b Battery) Color() Color {
	colors := GetColors()
	return colors[int(float64(b.Charge)/100.0*float64(len(colors)-1))]
}

// Symbol returns a suitable symbol for the given battery capacity/state.
func (b Battery) Symbol() string {
	prefix := map[BatStatus]string{
		BatUnknown:     "",
		BatCharging:    " ",
		BatDischarging: "",
		BatFull:        "",
	}
	symb := batBar[int(float64(b.Charge)/100.0*float64(len(batBar)-1))]
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

type BatteryGenerator struct{}

func (b BatteryGenerator) Generate(w *Widget, index int, ctx *GeneratorCtx) {

	gen := func() (e []Element, err error) {
		bats, err := BatteryInfo()
		if err != nil {
			return
		}
		for _, b := range bats {
			color := b.Color()
			e = append(e, Element{Name: "Battery", Instance: b.Path,
				Alignment: AlignRight, Color: &color,
				FullText: fmt.Sprintf("%d%% %s", b.Charge, b.Symbol())})
		}
		return
	}

	ticker := time.NewTicker(time.Second * 10)
	generator(w, index, ctx, ticker.C, gen)
	ticker.Stop()
	return
}
