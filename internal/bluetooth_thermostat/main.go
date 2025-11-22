package bluetooththermostat

import (
	"encoding/binary"

	"github.com/hoegaarden/go-bthome"
	"github.com/rs/zerolog/log"
	"tinygo.org/x/bluetooth"
)

const BleServiceTemperature = 0xFFE0

var bthomeUUID = bluetooth.New16BitUUID(binary.LittleEndian.Uint16(bthome.BTHomeUUID[:]))
var parser *bthome.Parser

type BluetoothThermostat struct {
	address     string
	isConnected bool
	// Add fields as necessary
}

func getParser() *bthome.Parser {
	if parser == nil {
		parser = bthome.NewParser()
	}
	return parser
}

func New(address string, callback func(temp, humidity, battery float64)) *BluetoothThermostat {
	thermostat := &BluetoothThermostat{
		isConnected: false,
		address:     address,
	}

	getParser().AddEncryptionKey(address, "key") // TODO: load key from config

	go start(thermostat.onScan)
	return thermostat
}

var BTCallbacks []func(*bluetooth.Adapter, bluetooth.ScanResult)

func start(onScan func(*bluetooth.Adapter, bluetooth.ScanResult)) {
	BTCallbacks = append(BTCallbacks, onScan)

	adapter := bluetooth.DefaultAdapter
	// Enable adapter
	err := adapter.Enable()
	if err != nil {
		log.Info().Err(err).Msg("Failed to enable BLE adapter")
	}

	// Start scanning and define callback for scan results
	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		for _, callback := range BTCallbacks {
			callback(adapter, device)
		}
	})

	if err != nil && err.Error() != "Operation already in progress" {
		log.Info().Err(err).Msg("Failed to register scan callback")
	}
}

func (b *BluetoothThermostat) IsConnected() bool {
	return b.isConnected
}

func (b *BluetoothThermostat) onScan(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	// log.Debug().Msgf("found device: %s, RSSI: %d, Name: %s, Payload: %v", device.Address.String(), device.RSSI, device.LocalName(), device.AdvertisementPayload)

	if device.ServiceData() != nil {
		if device.ServiceData()[0].UUID == bthomeUUID {
			log.Debug().Msgf("Found BTHome device: %s, RSSI: %d, Name: %s", device.Address.String(), device.RSSI, device.LocalName())
			if !device.HasServiceUUID(bthomeUUID) {
				return
			}

			addr := device.Address.String()

			for _, sd := range device.ServiceData() {
				packets, err := parser.Parse(addr, nil, sd.Data)
				if err != nil {
					log.Printf("[%s] error: %v\n", addr, err)
					continue
				}
				for _, p := range packets {
					log.Printf("[%s] %s\n", addr, p)
				}
			}
		}
	}
	// if device.Address.String() == b.address {
	// 	log.Info().Msgf("Found target device: %s", device.Address.String())
	// }
}
