package main

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var dpy = C.XOpenDisplay(nil)

func getBatteryPercentage(path string) (perc int, err error) {
	energy_now, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_now", path))
	if err != nil {
		return -1, err
	}
	energy_full, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_full", path))
	if err != nil {
		return -1, err
	}
	var enow, efull int
	fmt.Sscanf(string(energy_now), "%d", &enow)
	fmt.Sscanf(string(energy_full), "%d", &efull)
	return enow * 100 / efull, nil
}

func getLoadAverage(file string) (lavg string, err error) {
	loadavg, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		return "Couldn't read loadavg", err
	}
	return strings.Join(strings.Fields(string(loadavg))[:3], " "), nil
}

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func formatStatus(format string, args ...interface{}) *C.char {
	status := fmt.Sprintf(format, args...)
	return C.CString(status)
}

func main() {
	if dpy == nil {
		log.Fatal("Can't open display")
	}
	for {
		t := time.Now().Format("15:04")
		b, err := getBatteryPercentage("/sys/class/power_supply/BAT0")
		if err != nil {
			log.Fatal(err)
		}
		l, err := getLoadAverage("/proc/loadavg")
		if err != nil {
			log.Fatal(err)
		}
		s := formatStatus("%s :: %s :: %d%%", l, t, b)
		setStatus(s)
		time.Sleep(time.Second)
	}
}
