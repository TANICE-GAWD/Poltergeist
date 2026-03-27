package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

const DEVICE_MAC = "9C:DE:F0:33:F5:DD"

var DEVICE_NAME_KEYWORDS = []string{"Nord Buds", "OnePlus"}

var CMD_CHAR bluetooth.UUID
var NOTIFY_CHAR bluetooth.UUID

func init() {
	var err error

	CMD_CHAR, err = bluetooth.ParseUUID("0100079A-D102-11E1-9B23-00025B00A5A5")
	if err != nil {
		panic(err)
	}

	NOTIFY_CHAR, err = bluetooth.ParseUUID("0200079A-D102-11E1-9B23-00025B00A5A5")
	if err != nil {
		panic(err)
	}
}



func disconnectDevice() {
	fmt.Println("[SYS] Disconnecting audio...")
	exec.Command("bluetoothctl", "disconnect", DEVICE_MAC).Run()
	time.Sleep(2 * time.Second)
}

func reconnectDevice() {
	fmt.Println("[SYS] Reconnecting audio...")
	exec.Command("bluetoothctl", "connect", DEVICE_MAC).Run()
	time.Sleep(3 * time.Second)
}

func findDevice() (*bluetooth.ScanResult, error) {
	fmt.Println("[*] Scanning for Nord Buds...")

	var found *bluetooth.ScanResult

	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		name := result.LocalName()

		if result.Address.String() == DEVICE_MAC {
			fmt.Println("[FOUND]", name, DEVICE_MAC)
			found = &result
			adapter.StopScan()
			return
		}

		for _, k := range DEVICE_NAME_KEYWORDS {
			if strings.Contains(name, k) {
				fmt.Println("[FOUND]", name)
				found = &result
				adapter.StopScan()
				return
			}
		}
	})

	if err != nil {
		return nil, err
	}

	if found == nil {
		return nil, fmt.Errorf("device not found")
	}

	return found, nil
}

func handleNotify(data []byte) {
	fmt.Println("[RX RAW]", hexStr(data))

	
	if len(data) > 5 && data[0] == 0xAA {
		fmt.Printf("[PROTO] CAT=0x%02X SUB=0x%02X LEN=%d\n",
			data[4], data[5], len(data))
	}

	
	if len(data) >= 16 && data[4] == 0x06 && data[5] == 0x81 {
		left := data[12]
		right := data[14]
		caseB := data[15]

		fmt.Println("\n========== BATTERY INFO ==========")
		fmt.Printf("Left Bud:  %d%%\n", left)
		fmt.Printf("Right Bud: %d%%\n", right)
		fmt.Printf("Case:      %d%%\n", caseB)
		fmt.Println("==================================\n")
	}
}

func sendPacket(char bluetooth.DeviceCharacteristic, data []byte, name string) {
	fmt.Println("[TX]", name+":", hexStr(data))


	n, err := char.WriteWithoutResponse(data)
	if char.UUID().String() == "" {
		fmt.Println("[ERROR] Characteristic not initialized")
		return
	}
	if err != nil {
		fmt.Println("[ERROR] Write failed:", err)
		return
	}

	if n != len(data) {
		fmt.Println("[WARN] Partial write:", n, "/", len(data))
	}
}

func doHandshake(cmdChar bluetooth.DeviceCharacteristic) {
	sendPacket(cmdChar, []byte{
		0xAA, 0x07, 0x00, 0x00, 0x00, 0x01, 0x23, 0x00, 0x00, 0x12,
	}, "HELLO")

	time.Sleep(2 * time.Second)

	sendPacket(cmdChar, []byte{
		0xAA, 0x0C, 0x00, 0x00, 0x00, 0x85, 0x41, 0x05,
		0x00, 0x00, 0xB5, 0x50, 0xA0, 0x69,
	}, "REGISTER")

	time.Sleep(2 * time.Second)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: nordbuds <listen|on|off|trans|battery>")
		return
	}

	cmd := strings.ToLower(os.Args[1])

	adapter.Enable()

	
	
	fmt.Println("[INFO] Keeping existing connection (no forced disconnect)")

	result, err := findDevice()
	if err != nil {
		fmt.Println("[ERROR]", err)
		return
	}

	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		fmt.Println("[ERROR] Connection failed:", err)
		return
	}

	fmt.Println("[OK] Connected")
	fmt.Println("[WAIT] Stabilizing connection...")
	time.Sleep(5 * time.Second)

	time.Sleep(2 * time.Second)

	services, err := device.DiscoverServices(nil)
	if err != nil {
		fmt.Println("[ERROR] Service discovery failed:", err)
		return
	}

	var cmdChar bluetooth.DeviceCharacteristic
	var notifyChar bluetooth.DeviceCharacteristic

	for _, s := range services {
		chars, _ := s.DiscoverCharacteristics(nil)
		for _, c := range chars {
			if c.UUID() == CMD_CHAR {
				cmdChar = c
			}
			if c.UUID() == NOTIFY_CHAR {
				notifyChar = c
			}
		}
	}

	if cmdChar.UUID().String() == "" || notifyChar.UUID().String() == "" {
		fmt.Println("[ERROR] Required characteristics not found")
		return
	}

	notifyChar.EnableNotifications(func(buf []byte) {
		handleNotify(buf)
	})

	time.Sleep(1 * time.Second)

	switch cmd {

	case "listen":
		fmt.Println("[MODE] Listening for gestures...")

		doHandshake(cmdChar)

		fmt.Println("[READY] Perform gestures now...\n")

		select {} 

	case "on":
		doHandshake(cmdChar)
		sendPacket(cmdChar, []byte{
			0xAA, 0x0A, 0x00, 0x00, 0x04, 0x04,
			0x42, 0x03, 0x00, 0x01, 0x01, 0x01,
		}, "ANC ON")

	case "off":
		doHandshake(cmdChar)
		sendPacket(cmdChar, []byte{
			0xAA, 0x0A, 0x00, 0x00, 0x04, 0x04,
			0x42, 0x03, 0x00, 0x01, 0x01, 0x04,
		}, "ANC OFF")

	case "trans":
		doHandshake(cmdChar)
		sendPacket(cmdChar, []byte{
			0xAA, 0x0A, 0x00, 0x00, 0x04, 0x04,
			0x42, 0x03, 0x00, 0x01, 0x01, 0x02,
		}, "TRANSPARENCY")

	case "battery":
		doHandshake(cmdChar)
		sendPacket(cmdChar, []byte{
			0xAA, 0x07, 0x00, 0x00, 0x06, 0x01, 0x25, 0x00, 0x00,
		}, "BATTERY QUERY")

	default:
		fmt.Println("Unknown command")
	}

	
	reconnectDevice()
}