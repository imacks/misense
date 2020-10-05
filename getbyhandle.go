package misense

import (
	"github.com/currantlabs/ble"
)

func getServiceByHandle(profile *ble.Profile, handle uint16) *ble.Service {
	for _, s := range profile.Services {
		if s.Handle == handle {
			return s
		}
	}

	return nil
}

func getCharacteristicByHandle(profile *ble.Profile, handle uint16) *ble.Characteristic {
	for _, s := range profile.Services {
		for _, c := range s.Characteristics {
			if c.Handle == handle {
				return c
			}
		}
	}

	return nil
}

func getCharacteristicByValueHandle(profile *ble.Profile, handle uint16) *ble.Characteristic {
	for _, s := range profile.Services {
		for _, c := range s.Characteristics {
			if c.ValueHandle == handle {
				return c
			}
		}
	}

	return nil
}

func getDescriptorByHandle(profile *ble.Profile, handle uint16) *ble.Descriptor {
	for _, s := range profile.Services {
		for _, c := range s.Characteristics {
			for _, d := range c.Descriptors {
				if d.Handle == handle {
					return d
				}
			}
		}
	}

	return nil
}