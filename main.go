// TODO: Code cleanup
package main

import (
	"bufio"
	"os"
	"os/exec"
	"time"
)

func main() {
	widgets := []Widget{
		Widget{Generator: StreamingCommandGenerator{
			Instance: "nmcliwatcher",
			CmdCreator: func() *exec.Cmd {
				return exec.Command("nm_watcher", "wlp3s0")
			}}},

		Widget{Generator: StreamingCommandGenerator{
			Instance: "mullvadwatcher",
			CmdCreator: func() *exec.Cmd {
				return exec.Command("mullvad_watcher")
			}}},

		Widget{
			Generator: CommandGenerator{Instance: "mullvadvpn",
				C:      time.Tick(time.Second * 10),
				IsJSON: true,
				CmdCreator: func() *exec.Cmd {
					return exec.Command("mullvad_jsonblock")
				}}},

		Widget{Generator: BatteryGenerator{
			Alignment: AlignRight,
			Every:     time.Second * 10,
		}},

		Widget{Generator: CPUGenerator{
			Alignment: AlignRight,
			Every:     time.Second * 10,
		}},

		Widget{Generator: ClockGenerator{
			Format:    "Mon Jan 2 15:04:05",
			Alignment: AlignRight,
			Every:     time.Second,
		}},
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	i3writer := NewI3BarWriter(I3BarHeader{Version: 1}, *out)
	status := NewStatus(i3writer)

	for _, w := range widgets {
		status.AddWidget(w)
	}

	status.Start()
}
