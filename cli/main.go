package main

import (
	"fmt"
	"context"
	"time"
	"flag"
	"strconv"
	"os"
	"os/signal"
	"syscall"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"

	"github.com/imacks/misense"
)

var (
	bleTimeout int
	scanDedup bool
	showInfo bool
	sensorAddr []string
	powerSaver bool
	readTime int
	setRTC bool
)

func init() {
	flag.IntVar(&bleTimeout, "t", 5, "Scan/connect timeout in seconds")
	flag.BoolVar(&scanDedup, "u", false, "Scan dedup")
	flag.BoolVar(&showInfo, "i", false, "Show info about sensor")
	flag.BoolVar(&setRTC, "T", false, "Sets the sensor RTC time to the current system time")
	flag.BoolVar(&powerSaver, "B", false, "Enable battery saving mode")
	flag.IntVar(&readTime, "r", 0, "Number of seconds to read. 0 or negative to run forever")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s %s () %s\n", "misense", "1.0.0", "CLI to interface with Mi2 LYWSD03MMC")
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintf(os.Stdout, "Usage: %s [-t 5] [-u]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s -i [-t 5] <mac1> [<mac2>...]\n", os.Args[0])
		fmt.Fprintf(os.Stdout, "       %s [-t 5] [-T] [-B] [-r 0] <mac1> [<mac2>...]\n", os.Args[0])
		fmt.Fprintln(os.Stdout, "")
		fmt.Fprintln(os.Stdout, "Parameters:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	sensorAddr = flag.Args()

	runmode := "scan"
	if len(sensorAddr) > 0 {
		if showInfo == true {
			runmode = "show"
		} else {
			runmode = "read"
		}
	}

	timeout, _ := time.ParseDuration(strconv.Itoa(bleTimeout) + "s")

	if runmode == "scan" {
		err := scan(timeout, !scanDedup)
		if err != nil {
			panic(err)
		}
		return
	} else if runmode == "show" {
		err := drawDevice(sensorAddr, timeout)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	} else if runmode == "read" {
		err := readDevice(sensorAddr, timeout)
		if err != nil {
			panic(err)
		}
		return
	}
}

func drawDevice(macaddr []string, timeout time.Duration) error {
	host, errDevice := linux.NewDevice()
	if errDevice != nil {
		return errDevice
	}
	defer func() {
		fmt.Printf("stop host\n")
		host.Stop()
	}()

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), timeout))

	for _, v := range macaddr {
		fmt.Printf("connecting to %s\n", v)
		client, err := host.Dial(ctx, ble.NewAddr(v))
		if err != nil {
			return err
		}
		defer func() {
			fmt.Printf("disconnect %s\n", v)
			client.CancelConnection()
		}()

		p, err := client.DiscoverProfile(true)
		if err != nil {
			return err
		}

		drawDeviceTree(client, p)
	}

	return nil
}

func readDevice(macaddr []string, timeout time.Duration) error {
	host, errDevice := linux.NewDevice()
	if errDevice != nil {
		return errDevice
	}
	defer func() {
		fmt.Printf("stop host\n")
		host.Stop()
	}()

	var sensors = make([]*misense.MiTDSensor, len(macaddr))
	defer func() {
		for _, v := range sensors {
			if v == nil {
				continue
			}
			fmt.Printf("disconnecting %s\n", v.MAC())
			err := v.Disconnect()
			if err != nil {
				fmt.Printf("ERROR! %v\n", err)
			}
		}
	}()

	for i, v := range macaddr {
		sensors[i] = misense.NewMiTDSensor(host, v)
		errConn := sensors[i].Connect(timeout)
		if errConn != nil {
			return errConn
		}

		fmt.Printf("sensor %s version %s\n", sensors[i].MAC(), sensors[i].Version())

		t, err := sensors[i].Time()
		if err != nil {
			return err
		}

		fmt.Printf("sensor %s rtc: %s\n", sensors[i].MAC(), t.String())
		if setRTC {
			fmt.Printf("sensor %s set_rtc\n", sensors[i].MAC())
			err = sensors[i].SetTime(time.Now())
			if err != nil {
				return err
			}
		}

		fmt.Printf("sensor %s enable notify\n", sensors[i].MAC())
		err = sensors[i].EnableNotify()
		if err != nil {
			return err
		}
		
		if powerSaver {
			fmt.Printf("sensor %s enable power_saving\n", sensors[i].MAC())
			err = sensors[i].SavePower()
			if err != nil {
				return err
			}
		}

		sensors[i].Subscribe(func(r *misense.THReading) {
			fmt.Printf("[%s] read %s\n", time.Now().Format("02-Jan-06 15:04:05 MST"), r.String())
		})
	}

	if readTime < 1 {
		fmt.Printf("press ctrl c to quit\n")
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Printf("stopping\n")
	} else {
		timeout, _ := time.ParseDuration(strconv.Itoa(readTime) + "s")
		time.Sleep(timeout)
	}

	return nil
}



	/*

	maxT, minT, maxH, minH, err2 := mi2.Comfortable()
	if err2 != nil {
		return err2
	}
	fmt.Printf("Comfortable maxt %f mint %f maxh %d minh %d\n", maxT, minT, maxH, minH)

	err2 = mi2.SetComfortable(27, 26, 60, 20)
	if err2 != nil {
		return err2
	}
	*/
