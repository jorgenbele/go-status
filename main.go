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
		Widget{Generator: StreamingCommandGenerator{Instance: "nmcliwatcher", CmdCreator: func() *exec.Cmd { return exec.Command("nm_watcher", "wlp3s0") }}},
		Widget{Generator: StreamingCommandGenerator{Instance: "mullvadwatcher", CmdCreator: func() *exec.Cmd { return exec.Command("mullvad_watcher") }}},
		Widget{Generator: CommandGenerator{Instance: "mullvadvpn", C: time.Tick(time.Second * 10), IsJSON: true, CmdCreator: func() *exec.Cmd { return exec.Command("mullvad_jsonblock") }}},
		Widget{Generator: BatteryGenerator{}},
		Widget{Generator: CPUGenerator{}},
		Widget{Generator: ClockGenerator{}},
		Widget{Generator: CommandGenerator{Instance: "File watcher", C: NewFsNotifyTicker([]string{"/home/jbr/.not_kv"}).C, CmdCreator: func() *exec.Cmd { return exec.Command("notification") }}},

		//Widget{Generator: CommandGenerator{Instance: "sleeptest", C: time.Tick(time.Second), CmdCreator: func() *exec.Cmd { return exec.Command("sleeptest") }}},
		//Widget{Generator: CommandGenerator{Instance: "errortest", C: time.Tick(time.Second), CmdCreator: func() *exec.Cmd { return exec.Command("errortest") }}},
		//Widget{Generator: CommandGenerator{Instance: "uptime", Tick: time.Second * 10, TrimSpace: true, CmdCreator: func() *exec.Cmd { return exec.Command("uptime", "-p") }}},
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
