package main

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

var dpy = C.XOpenDisplay(nil)

func getBatteryPercentage(path string) (perc int, err error) {
	energy_now, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_now", path))
	if err != nil {
		perc = -1
		return
	}
	energy_full, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_full", path))
	if err != nil {
		perc = -1
		return
	}
	var enow, efull int
	fmt.Sscanf(string(energy_now), "%d", &enow)
	fmt.Sscanf(string(energy_full), "%d", &efull)
	perc = enow * 100 / efull
	return
}

func getLoadAverage(file string) (lavg string, err error) {
	loadavg, err := ioutil.ReadFile(file)
	if err != nil {
		return "Couldn't read loadavg", err
	}
	lavg = strings.Join(strings.Fields(string(loadavg))[:3], " ")
	return
}

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func nowPlaying(addr string) (np string, err error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		np = "Couldn't connect to mpd."
		return
	}
	defer conn.Close()
	reply := make([]byte, 512)
	conn.Read(reply) // The mpd OK has to be read before we can actually do things.

	message := "status\n"
	conn.Write([]byte(message))
	conn.Read(reply)
	r := string(reply)
	arr := strings.Split(string(r), "\n")
	if arr[8] != "state: play" { //arr[8] is the state according to the mpd documentation
		status := strings.SplitN(arr[8], ": ", 2)
		np := fmt.Sprintf("mpd - [%s]", status[1]) //status[1] should now be stopped or paused
		return
	}

	message = "currentsong\n"
	conn.Write([]byte(message))
	conn.Read(reply)
	r = string(reply)
	arr = strings.Split(string(r), "\n")
	if len(arr) > 5 {
		var artist, title string
		for _, info := range arr {
			field := strings.SplitN(info, ":", 2)
			switch field[0] {
			case "Artist":
				artist = field[1]
			case "Title":
				title = field[1]
			}
		}
		np = artist + " - " + title
		return
	} else { //This is a nonfatal error.
		np = "Playlist is empty."
		return
	}
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
		t := time.Now().Format("Mon 02 15:04")
		b, err := getBatteryPercentage("/sys/class/power_supply/BAT0")
		if err != nil {
			log.Println(err)
		}
		l, err := getLoadAverage("/proc/loadavg")
		if err != nil {
			log.Println(err)
		}
		m, err := nowPlaying("localhost:6600")
		if err != nil {
			log.Println(err)
		}
		s := formatStatus("%s :: %s :: %s :: %d%%", m, l, t, b)
		setStatus(s)
		time.Sleep(time.Second)
	}
}
