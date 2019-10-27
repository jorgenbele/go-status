package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"github.com/jorgenbele/go-status/status"
)

func main() {
	// NOTE: time.Tick(...) is used instead of time.NewTicker(...) because
	// the program will be shutting down when the Widget is shutting down.
	widgets := []status.Widget{
		status.Widget{
			Gen: CmdGen{Instance: "spotify",
				C:      time.Tick(time.Second * 10),
				IsJSON: true,
				CmdCreator: func() *exec.Cmd {
					return exec.Command("spotifystatus", "--json")
				}}},

		status.Widget{Gen: StreamingCmdGen{
			Instance: "nmcliwatcher",
			CmdCreator: func() *exec.Cmd {
				return exec.Command("nm_watcher", "wlp3s0")
			}}},

		status.Widget{Gen: StreamingCmdGen{
			Instance: "mullvadwatcher",
			CmdCreator: func() *exec.Cmd {
				return exec.Command("mullvad_watcher")
			}}},

		status.Widget{
			Gen: CmdGen{Instance: "mullvadvpn",
				C:      time.Tick(time.Second * 10),
				IsJSON: true,
				CmdCreator: func() *exec.Cmd {
					return exec.Command("mullvad_jsonblock")
				}}},

		status.Widget{Gen: BatteryGen{
			Alignment: status.AlignRight,
			Every:     time.Second * 10,
		}},

		status.Widget{Gen: CPUGen{
			Alignment: status.AlignRight,
			Every:     time.Second * 10,
		}},

		status.Widget{Gen: ClockGen{
			Format:    "Mon Jan 2 15:04:05",
			Alignment: status.AlignRight,
			Every:     time.Second,
		}},
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [--format (i3bar | lemonbar | dzen2)]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nIf no --format <bar> is specified then i3bar is used.\n")
		os.Exit(1)
	}

	var b status.Bar
	if len(os.Args) == 3 && os.Args[1] == "--format" {
		switch os.Args[2] {
		case "lemonbar":
			b = status.NewLemonbar(out)
			break

		case "dzen2":
			b = status.NewDzen2Bar(out)
			break

		case "i3bar":
			b = status.NewI3Bar(status.I3BarHeader{Version: 1}, out)
			break

		default:
			usage()
		}
	} else if len(os.Args) == 1 {
		b = status.NewI3Bar(status.I3BarHeader{Version: 1}, out)
	} else {
		usage()
	}

	s := status.NewStatus(b)

	for _, w := range widgets {
		s.AddWidget(w)
	}

	sigtermch := make(chan os.Signal)
	signal.Notify(sigtermch, os.Interrupt, syscall.SIGTERM)
	s.SetTermSignal(sigtermch)

	//sigstopch := make(chan os.Signal)
	//signal.Notify(sigstopch, os.Interrupt, syscall.SIGTSTP)
	//status.SetStopSignal(sigstopch)

	//sigcontch := make(chan os.Signal)
	//signal.Notify(sigcontch, os.Interrupt, syscall.SIGCONT)
	//status.SetContSignal(sigcontch)

	s.Start()
}
