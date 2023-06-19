package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/brutella/hap/accessory"

	"github.com/ryanseddon/powerwall-homekit/grid"
	"github.com/ryanseddon/powerwall-homekit/powerwall"
	"github.com/brutella/hap"
)

var powerwallIP, pinCode string

const ipDefault = ""
const pinCodeDefault = "00102003"

func main() {
	flag.StringVar(&powerwallIP, "ip", ipDefault, "ip address of powerwall")
	flag.StringVar(&pinCode, "pin", pinCodeDefault, "homekit pin code to use for this accessory")

	flag.Parse()

	if powerwallIP == ipDefault {
		fmt.Printf("Usage of %s:\n", os.Args[0])

		flag.PrintDefaults()

		os.Exit(1)
	}

	ip := net.ParseIP(powerwallIP)

	bridgeInfo := accessory.Info{Name: "Tesla Bridge"}

	bridge := accessory.NewBridge(bridgeInfo)

	pw := powerwall.NewPowerwall(ip)

	sensor := grid.NewSensor(ip)

	// pwConfig := hap.Config{Pin: pinCode}
	pwStore := hap.NewMemStore()
	pwStore.Set("Pin", []byte(pinCode))

	// NOTE: the first accessory in the list acts as the bridge, while the rest will be linked to it
	t, err := hap.NewServer(pwStore, bridge.A, pw.A, sensor.A)
	if err != nil {
		log.Panic(err)
	}

	// Setup a listener for interrupts and SIGTERM signals
	// to stop the server.
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		// Stop delivering signals.
		signal.Stop(c)
		// Cancel the context to stop the server.
		cancel()
	}()

	// Run the server.
	t.ListenAndServe(ctx)
}
