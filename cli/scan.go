package main

import (
	"fmt"
	"context"
	"time"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
)

func scan(timeout time.Duration, allowDup bool) error {
	host, errDevice := linux.NewDevice()
	if errDevice != nil {
		return errDevice
	}
	defer host.Stop()

	fmt.Printf("[VERBOSE] Scan for %s...\n", timeout)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), timeout))

	err := host.Scan(ctx, allowDup, func(a ble.Advertisement) {
		fmt.Printf("ADDR %s\n", a.Address())
		fmt.Printf("  rssi %3d\n", a.RSSI())
		if a.Connectable() {
			fmt.Printf("  allow_connect y\n")
		} else {
			fmt.Printf("  allow_connect n\n")
		}

		if len(a.LocalName()) > 0 {
			fmt.Printf("  name %s\n", a.LocalName())
		}

		if len(a.Services()) > 0 {
			fmt.Printf("  services %v\n", a.Services())
		}
		if len(a.ManufacturerData()) > 0 {
			fmt.Printf("  oem %X\n", a.ManufacturerData())
		}
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			return nil
		}
		return err
	}
	return nil
}