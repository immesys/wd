package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/immesys/wd"
	"github.com/mgutz/ansi"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "wd"
	app.Usage = "control watchdogs"
	app.Version = "1.3.0"
	app.Commands = []cli.Command{
		{
			Name:      "kick",
			Usage:     "create or kick a watchdog",
			Action:    cli.ActionFunc(actionKick),
			ArgsUsage: "name [timeout]",
		},
		{
			Name:      "fault",
			Usage:     "fault a watchdog",
			Action:    cli.ActionFunc(actionFault),
			ArgsUsage: "name reason",
		},
		{
			Name:      "retire",
			Usage:     "retire a set of watchdogs by prefix",
			Action:    cli.ActionFunc(actionRetire),
			ArgsUsage: "prefix",
		},
		{
			Name:      "auth",
			Usage:     "create a new prefix auth key",
			Action:    cli.ActionFunc(actionAuth),
			ArgsUsage: "prefix",
		},
		{
			Name:      "clear",
			Usage:     "clear the cumulative downtime on a prefix",
			Action:    cli.ActionFunc(actionClear),
			ArgsUsage: "prefix",
		},
		{
			Name:      "status",
			Usage:     "list watchdog status",
			Action:    cli.ActionFunc(actionStatus),
			ArgsUsage: "prefix",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "noheader",
				},
				cli.BoolFlag{
					Name: "nocolor",
				},
				cli.BoolFlag{
					Name: "tabsep",
				},
				cli.BoolFlag{
					Name: "nobadfirst",
				},
			},
		}}
	app.Run(os.Args)
}

func actionKick(c *cli.Context) error {
	if len(c.Args()) < 1 || len(c.Args()) > 2 {
		fmt.Println("Usage: wd kick name [timeout]")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	var timeout int64 = 300
	if len(c.Args()) == 2 {
		var err error
		timeout, err = strconv.ParseInt(c.Args()[1], 10, 32)
		if err != nil {
			fmt.Println("Invalid timeout:", err)
			os.Exit(1)
		}
	}
	err := wd.Kick(c.Args()[0], int(timeout))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}
func actionClear(c *cli.Context) error {
	if len(c.Args()) < 1 || len(c.Args()) > 2 {
		fmt.Println("Usage: wd clear prefix")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	wd.Clear(c.Args()[0])
	return nil
}
func actionRetire(c *cli.Context) error {
	if len(c.Args()) != 1 {
		fmt.Println("Usage: wd retire prefix")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	wd.Retire(c.Args()[0])
	return nil
}
func actionFault(c *cli.Context) error {
	if len(c.Args()) < 2 {
		fmt.Println("Usage: wd fault name reason")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	reason := strings.Join(c.Args()[1:], " ")
	wd.Fault(c.Args()[0], reason)
	return nil
}
func actionAuth(c *cli.Context) error {
	if len(c.Args()) != 1 {
		fmt.Println("Usage: wd auth prefix")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	key, err := wd.Auth(c.Args()[0])
	if err == nil {
		fmt.Println(key)
		os.Exit(0)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
	return nil
}
func actionStatus(c *cli.Context) error {
	if len(c.Args()) != 1 {
		fmt.Println("Usage: wd status prefix")
		os.Exit(1)
	}
	if !wd.ValidPrefix(c.Args()[0]) {
		fmt.Println("watchdog names must match [a-z0-9\\._]")
		os.Exit(1)
	}
	st, err := wd.Status(c.Args()[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
	namemax := 4
	for _, s := range st {
		if len(s.Name) > namemax {
			namemax = len(s.Name)
		}
	}
	var fline string
	color := !c.Bool("nocolor")
	noheader := c.Bool("noheader")
	if c.Bool("tabsep") {
		color = false
		noheader = true
		fline = "%s\t%s\t%s\t%s\t%s\n"
	} else {
		fline = "%4s %-" + strconv.Itoa(namemax) + "s %-32s %-8s %s\n"
	}
	if !noheader {
		fmt.Printf(fline, "STATE", "NAME", "EXPIRE", "CUMD", "REASON")
	}
	do := func(s wd.WDStatus) {
		if color {
			if s.Status != "KGOOD" {
				fmt.Print(ansi.ColorCode("red+b"))
			} else {
				fmt.Print(ansi.ColorCode("green+b"))
			}
		}
		s.CumDTime -= time.Duration(int64(s.CumDTime) % 1000000000)
		fmt.Printf(fline, s.Status, s.Name, s.Expires, s.CumDTime, strings.TrimSpace(s.Reason))
		if color {
			fmt.Print(ansi.Reset)
		}
	}
	if !c.Bool("nobadfirst") {
		for _, s := range st {
			if s.Status != "KGOOD" {
				do(s)
			}
		}
		for _, s := range st {
			if s.Status == "KGOOD" {
				do(s)
			}
		}
	} else {
		for _, s := range st {
			do(s)
		}
	}

	return nil
}
