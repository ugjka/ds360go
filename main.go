package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const DSDEVVID = "054c"
const DSDEVPID = "0ce6"
const SEARCHDEV = "Vendor=" + DSDEVVID + " Product=" + DSDEVPID
const DEVICES = "/proc/bus/input/devices"
const EVENTREGEX = `event\d+`
const EVDEVPATH = "evdev=/dev/input/%s"
const PIDFILE = "/tmp/ds360go.pid"

const XBOXDRVCONF = `
	xboxdrv 
   	-o %s 
    --mimic-xpad
	--silent
    --quiet
    --axismap
    -y1=y1,-y2=y2
    --evdev-absmap
    ABS_HAT0X=dpad_x,ABS_HAT0Y=dpad_y,ABS_X=X1,ABS_Y=Y1,ABS_RX=X2,ABS_RY=Y2,ABS_Z=LT,ABS_RZ=RT
    --evdev-keymap 
    BTN_SOUTH=A,BTN_EAST=B,BTN_NORTH=Y,BTN_WEST=X,BTN_START=start,BTN_MODE=guide,BTN_SELECT=back,BTN_TL=LB,BTN_TR=RB,BTN_TL2=LT,BTN_TR2=RT,BTN_THUMBL=TL,BTN_THUMBR=TR
`

func main() {
	// check for dependencies
	exes := []string{
		"xboxdrv",
		"pgrep",
	}
	for _, exe := range exes {
		if _, err := exec.LookPath(exe); err != nil {
			fmt.Fprintln(os.Stderr, "Fatal! dependency:", err)
			os.Exit(1)
		}
	}

	debug := flag.Bool("debug", false, "debug log")
	flag.Parse()

	if !*debug {
		log.SetOutput(&DummyWriter{})
	}

	present, err := checkPresent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal! checkPresent: %v\n", err)
		os.Exit(1)
	}
	if !present {
		fmt.Fprintf(os.Stderr, "Fatal! VID '%s' not found in /proc/bus/input/devices, please make sure your Dualsense is connected\n", DSDEVVID)
		os.Exit(1)
	}

	evpath, err := findEvdevPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal! findEvdevPath: %v", err)
		os.Exit(1)
	}
	log.Println(evpath)

	command := parseXboxdrv(evpath)

	err = command.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal! start xboxdrv: %v", err)
		os.Exit(1)
	}

	log.Println("xboxdrv started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		log.Println(<-sig, "caught, exiting")
		if command.Process != nil {
			command.Process.Kill()
			command.Wait()
		}
		os.Remove(PIDFILE)
		os.Exit(0)
	}()

	err = os.WriteFile(PIDFILE, []byte(fmt.Sprint(os.Getpid())), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal! could not write pid file: %v", err)
		os.Exit(1)
	}

	// Guard against steam, because it runs its own gamepad manager that interferes with
	// xboxdrv, games like Gta 4 running via wine get confused
	tick := time.NewTicker(time.Second * 5)
	for {
		<-tick.C

		if steamGameoverlayuiRunning() {
			if command.Process != nil {
				log.Println("steam overlay detected, killing xboxdrv")
				command.Process.Kill()
				command.Wait()
				command = parseXboxdrv(evpath)
			}
			continue
		}

		if command.Process == nil {
			log.Println("steam overlay no more, ressurecting xboxdrv")
			command.Start()
		}
	}
}

func checkPresent() (present bool, err error) {
	data, err := os.ReadFile(DEVICES)
	if err != nil {
		return present, err
	}
	return strings.Contains(string(data), SEARCHDEV), nil
}

func findEvdevPath() (path string, err error) {
	data, err := os.ReadFile(DEVICES)
	if err != nil {
		return path, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var found bool

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), SEARCHDEV) {
			found = true
			break
		}
	}

	if !found {
		return path, fmt.Errorf("VID '%s' not found in /proc/bus/input/devices, please make sure your Dualsense is connected", DSDEVVID)
	}

	evreg := regexp.MustCompile(EVENTREGEX)
	for scanner.Scan() {
		if evreg.MatchString(scanner.Text()) {
			eventNum := evreg.FindStringSubmatch(scanner.Text())[0]
			return fmt.Sprintf(EVDEVPATH, eventNum), nil
		}
	}

	return path, fmt.Errorf("event number not found")
}

func parseXboxdrv(evpath string) *exec.Cmd {
	arr := strings.Fields(fmt.Sprintf(XBOXDRVCONF, evpath))
	var cleanup []string

	for _, v := range arr {
		v = strings.TrimSpace(v)
		if v != "" {
			cleanup = append(cleanup, v)
		}
	}

	cmd := exec.Command(cleanup[0], cleanup[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func steamGameoverlayuiRunning() bool {
	_, err := exec.Command("pgrep", "-fl", "gameoverlayui").Output()
	return err == nil
}

type DummyWriter struct{}

func (d *DummyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
