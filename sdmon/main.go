package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/immesys/wd"
	"github.com/urfave/cli"
)

var timeout int

func main() {
	app := cli.NewApp()
	app.Name = "sdmon"
	app.Usage = "Maintain systemd watchdogs"
	app.Version = "1.6.0"
	app.Flags = []cli.Flag{
		cli.DurationFlag{
			Name:  "interval",
			Value: 2 * time.Minute,
		},
		cli.DurationFlag{
			Name:  "holdoff",
			Value: 5 * time.Minute,
		},
		cli.StringSliceFlag{
			Name: "unit",
		},
	}
	app.Action = runApp
	app.Run(os.Args)

}

func (e *Engine) fault(unit, reason string) {
	la, ok := e.lastAction[unit]
	if !ok || la.msg != reason || time.Now().Sub(la.tm) > e.interval {
		wd.Fault(e.prefix+"sd."+e.displayName[unit], reason)
		e.lastAction[unit] = Action{tm: time.Now(), msg: reason}
	}
}
func (e *Engine) healthy(unit string) {
	la, ok := e.lastAction[unit]
	if !ok || la.msg != "K" || time.Now().Sub(la.tm) > e.interval {
		wd.Kick(e.prefix+"sd."+e.displayName[unit], e.timeout)
		e.lastAction[unit] = Action{tm: time.Now(), msg: "K"}
	}
}

type Action struct {
	tm  time.Time
	msg string
}
type Engine struct {
	//WD params
	prefix          string
	timeout         int
	interval        time.Duration
	healthyInterval time.Duration

	displayName map[string]string
	lastAction  map[string]Action
	conn        *dbus.Conn
}

func (e *Engine) Scan() {
	units, err := e.conn.ListUnits()
	if err != nil {
		panic(err)
	}
	nw := time.Now()
	addressed := make(map[string]bool)
	for _, u := range units {
		_, ok := e.displayName[u.Name]
		if ok {
			addressed[u.Name] = true
			//This is a unit we care about
			props, err := e.conn.GetUnitProperties(u.Name)
			if err != nil {
				panic(err)
			}
			aet := props["ActiveEnterTimestamp"]
			if u.SubState == "running" {
				//We might be ok, but there might be a holdoff
				if aet != nil && nw.Sub(time.Unix(0, int64(aet.(uint64))*1000)) > e.healthyInterval {
					//We are properly healthy
					e.healthy(u.Name)
				} else {
					//We faulting because it has not been running for long enough
					if aet != nil {
						e.fault(u.Name, fmt.Sprintf("only up since %s", time.Unix(0, int64(aet.(uint64))*1000)))
					} else {
						e.fault(u.Name, "running, but uptime unknown")
					}
				}
			} else {
				//we are not running
				axt := props["ActiveExitTimestamp"]
				if axt == nil {
					e.fault(u.Name, fmt.Sprintf("%s.%s since %s", u.ActiveState, u.SubState, time.Unix(0, int64(axt.(uint64))*1000)))
				} else {
					e.fault(u.Name, fmt.Sprintf("%s.%s", u.ActiveState, u.SubState))
				}
			}
		}
	}
	//Pick up all units not discovered
	for unit, _ := range e.displayName {
		if !addressed[unit] {
			e.fault(unit, "not observed")
		}
	}
}
func runApp(c *cli.Context) error {
	e := Engine{
		displayName: make(map[string]string),
		lastAction:  make(map[string]Action),
	}
	hn, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	e.prefix = hn
	e.prefix = strings.Replace(e.prefix, "-", "_", -1)
	e.prefix = strings.ToLower(e.prefix)
	e.prefix = "410.br." + e.prefix + "."
	e.interval = c.Duration("interval")
	e.timeout = int((c.Duration("interval") / time.Second)) * 2
	e.healthyInterval = c.Duration("holdoff")
	for _, u := range c.StringSlice("unit") {
		parts := strings.Split(u, ":")
		if len(parts) == 1 {
			e.displayName[parts[0]+".service"] = parts[0]
		} else {
			e.displayName[parts[0]+".service"] = parts[1]
		}
	}
	e.conn, err = dbus.NewSystemConnection()
	if err != nil {
		panic(err)
	}
	for {
		e.Scan()
		time.Sleep(e.interval)
	}
}
