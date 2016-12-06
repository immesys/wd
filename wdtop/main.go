package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/immesys/wd"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/urfave/cli"
)

var timeout int

func main() {
	app := cli.NewApp()
	app.Name = "wdtop"
	app.Usage = "Maintain top watchdogs"
	app.Version = "1.4.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "prefix",
		},
		cli.Float64Flag{
			Name:  "min-mem-mb",
			Value: 1000.0,
		},
		cli.Float64Flag{
			Name:  "max-cpu-percent",
			Value: 80.0,
		},
		cli.StringSliceFlag{
			Name: "df",
		},
		cli.DurationFlag{
			Name:  "interval",
			Value: 2 * time.Minute,
		},
		cli.StringSliceFlag{
			Name: "proc",
		},
	}
	app.Action = runApp
	app.Run(os.Args)

}

func runApp(c *cli.Context) error {
	prefix := c.String("prefix")
	if prefix == "" {
		fmt.Println("You need to specify --prefix")
		os.Exit(1)
	}
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	timeout = int((c.Duration("interval") / time.Second)) * 2
	for {
		doMemory(prefix, c.Float64("min-mem-mb"))
		doCPU(prefix, c.Float64("max-cpu-percent"))
		doDisk(prefix, c.StringSlice("df"))
		doProc(prefix, c.StringSlice("proc"))
		time.Sleep(c.Duration("interval"))
	}
}
func doMemory(prefix string, minMB float64) {
	t, err := mem.VirtualMemory()
	if err != nil {
		wd.Fault(prefix+"memory", "unable to obtain stats")
		return
	}
	avMB := float64(t.Available) / 1024 / 1024
	if avMB > minMB {
		wd.Kick(prefix+"memory", timeout)
	} else {
		wd.Fault(prefix+"memory", fmt.Sprintf("%.2f MB available", avMB))
	}
}

func doCPU(prefix string, maxPercent float64) {
	t, err := cpu.Percent(0, false)
	if err != nil {
		panic(err)
	}
	sum := 0.0
	for _, ti := range t {
		sum += ti
	}
	sum /= float64(len(t))
	if sum < maxPercent {
		wd.Kick(prefix+"cpu", timeout)
	} else {
		wd.Fault(prefix+"cpu", fmt.Sprintf("%.2f %% CPU usage", sum))
	}
}

func doDisk(prefix string, reservations []string) {
	for _, r := range reservations {
		parts := strings.SplitN(r, ":", 3)
		dir := parts[0]
		name := parts[1]
		minMB, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			panic(err)
		}
		var stat syscall.Statfs_t
		syscall.Statfs(dir, &stat)
		av_mb := float64(stat.Bavail*uint64(stat.Bsize)) / 1024.0 / 1024.0
		if av_mb > minMB {
			wd.Kick(prefix+"disk."+name, timeout)
		} else {
			wd.Fault(prefix+"disk."+name, fmt.Sprintf("%.2f MB available", av_mb))
		}
	}
}

func doProc(prefix string, processNames []string) {
	ok := make(map[string]bool)
	repnames := make(map[string]string)
	for _, pn := range processNames {
		parts := strings.SplitN(pn, ":", 2)
		ok[parts[0]] = false
		repnames[parts[0]] = parts[1]
	}
	procs, err := ps.Processes()
	if err != nil {
		return
	}
	for _, p := range procs {
		_, match := ok[p.Executable()]
		if match {
			ok[p.Executable()] = true
		}
	}
	for pname, running := range ok {
		if running {
			wd.Kick(prefix+"ps."+repnames[pname], timeout)
		} else {
			wd.Fault(prefix+"ps."+repnames[pname], "not running")
		}
	}
}
