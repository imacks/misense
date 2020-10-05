package main

import (
	"fmt"

	"github.com/currantlabs/ble"
)

func drawDeviceTree(client ble.Client, p *ble.Profile) {
	fmt.Printf("[ble_device]\n")
	fmt.Printf("  - mac: %s\n", client.Address().String())
	fmt.Printf("  - rssi: %d\n", client.ReadRSSI())
	fmt.Printf("  - name: %s\n", client.Name())

	for _, s := range p.Services {
		fmt.Printf("  - [service]\n")
		drawService(s, "    -")

		for _, c := range s.Characteristics {
			fmt.Printf("%s [characteristic]\n", "    -")
			drawCharacteristic(client, c, "      -")

			for _, d := range c.Descriptors {
				fmt.Printf("%s [descriptor]\n", "      -")
				drawDescriptor(client, d, "        -")
			}
		}
	}
}

func drawDescriptor(client ble.Client, d *ble.Descriptor, treePrefix string) {
	fmt.Printf("%s uuid: %s\n", treePrefix, d.UUID.String())
	fmt.Printf("%s name: %s\n", treePrefix, ble.Name(d.UUID))
	fmt.Printf("%s handle: 0x%02x\n", treePrefix, d.Handle)

	b, err := client.ReadDescriptor(d)
	if err != nil {
		fmt.Printf("ERROR! %v\n", err)
	} else {
		fmt.Printf("%s value: %x | %q\n", treePrefix, b, b)
	}
}

func drawCharacteristic(client ble.Client, c *ble.Characteristic, treePrefix string) {
	fmt.Printf("%s uuid: %s\n", treePrefix, c.UUID.String())
	fmt.Printf("%s name: %s\n", treePrefix, ble.Name(c.UUID))
	fmt.Printf("%s property: 0x%02X | %s\n", treePrefix, c.Property, propString(c.Property))
	fmt.Printf("%s handle: 0x%02X\n", treePrefix, c.Handle)
	fmt.Printf("%s vhandle: 0x%02X\n", treePrefix, c.ValueHandle)

	if (c.Property & ble.CharRead) != 0 {
		b, err := client.ReadCharacteristic(c)
		if err != nil {
			fmt.Printf("ERROR! %v\n", err)
			return
		}
		
		fmt.Printf("%s value: %x | %q\n", treePrefix, b, b)
	} else {
		fmt.Printf("%s value: [no read permission]\n", treePrefix)
	}
}

func drawService(s *ble.Service, treePrefix string) {
	fmt.Printf("%s uuid: %s\n", treePrefix, s.UUID)
	fmt.Printf("%s name: %s\n", treePrefix, ble.Name(s.UUID))
	fmt.Printf("%s handle: 0x%02X\n", treePrefix, s.Handle)
}

func propString(p ble.Property) string {
	var s string
	for k, v := range map[ble.Property]string{
		ble.CharBroadcast:   "B",
		ble.CharRead:        "R",
		ble.CharWriteNR:     "w",
		ble.CharWrite:       "W",
		ble.CharNotify:      "N",
		ble.CharIndicate:    "I",
		ble.CharSignedWrite: "S",
		ble.CharExtended:    "E",
	} {
		if p&k != 0 {
			s += v
		}
	}
	return s
}
