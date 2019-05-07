// TODO: Code cleanup
package main

import (
	"bufio"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// NOTE: time.Tick(...) is used instead of time.NewTicker(...) because
	// the program will be shutting down when the Widget is shutting down.
	widgets := []Widget{
		Widget{
			Generator: CommandGenerator{Instance: "spotify",
				C:      time.Tick(time.Second * 10),
				IsJSON: true,
				CmdCreator: func() *exec.Cmd {
					return exec.Command("spotifystatus", "--json")
				}}},

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

	sigtermch := make(chan os.Signal)
	signal.Notify(sigtermch, os.Interrupt, syscall.SIGTERM)
	status.SetTermSignal(sigtermch)

	sigstopch := make(chan os.Signal)
	signal.Notify(sigstopch, os.Interrupt, syscall.SIGSTOP)
	status.SetStopSignal(sigstopch)

	sigcontch := make(chan os.Signal)
	signal.Notify(sigcontch, os.Interrupt, syscall.SIGCONT)
	status.SetContSignal(sigcontch)

	status.Start()
}
