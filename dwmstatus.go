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
		return "Couldn't read loadavg", err
	}
	return strings.Join(strings.Fields(string(loadavg))[:3], " "), nil
}

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func nowPlaying(addr string) (np string, err error) {
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		return "Couldn't connect to mpd.", err
	}
	reply := make([]byte, 512)
	conn.Read(reply) // The mpd OK has to be read before we can actually do things.
	message := "currentsong\n"
	conn.Write([]byte(message))
	conn.Read(reply)
	r := string(reply)
	arr := strings.Split(string(r), "\n")
	if len(arr) > 5 {
		title, artist := arr[3], arr[4] //This is very unpredictable and only works on the albums I own. It's probably better to iterate through the array and find the strings starting with "Artist: " and "Title: "
		title = strings.SplitAfterN(title, ":", 2)[1]
		artist = strings.SplitAfterN(artist, ":", 2)[1]
		np = artist + " - " + title
		return np, nil
	} else {
		return "Playlist is empty.", nil //This is a nonfatal error.
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
			log.Fatal(err)
		}
		l, err := getLoadAverage("/proc/loadavg")
		if err != nil {
			log.Fatal(err)
		}
		m, err := nowPlaying("localhost:6600")
		if err != nil {
			log.Fatal(err)
		}
		s := formatStatus("%s :: %s :: %s :: %d%%", m, l, t, b)
		setStatus(s)
		time.Sleep(time.Second)
	}
}
