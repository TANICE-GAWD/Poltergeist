package main







import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)





const (
	DEVICE_MAC  = "9C:DE:F0:33:F5:DD"
	CMD_UUID    = "0100079a-d102-11e1-9b23-00025b00a5a5"
	NOTIFY_UUID = "0200079a-d102-11e1-9b23-00025b00a5a5"
)

const (
	bluezService          = "org.bluez"
	bluezAdapter          = "org.bluez.Adapter1"
	bluezDevice           = "org.bluez.Device1"
	bluezGattChar         = "org.bluez.GattCharacteristic1"
	dbusObjectManager     = "org.freedesktop.DBus.ObjectManager"
	dbusProperties        = "org.freedesktop.DBus.Properties"
)








const (
	LEFT_SINGLE_TAP  = byte(0x01) 
	LEFT_DOUBLE_TAP  = byte(0x02) 
	LEFT_TRIPLE_TAP  = byte(0x03) 
	LEFT_LONG_PRESS  = byte(0x04) 
	RIGHT_SINGLE_TAP = byte(0x11) 
	RIGHT_DOUBLE_TAP = byte(0x12) 
	RIGHT_TRIPLE_TAP = byte(0x13) 
	RIGHT_LONG_PRESS = byte(0x14) 
	TAP_CATEGORY     = byte(0x0C) 
)





func macToPath(mac string) string {
	return strings.ReplaceAll(mac, ":", "_")
}

func getAdapterPath(conn *dbus.Conn) (dbus.ObjectPath, error) {
	obj := conn.Object(bluezService, "/")
	result := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)
	if err := obj.Call(dbusObjectManager+".GetManagedObjects", 0).Store(&result); err != nil {
		return "", fmt.Errorf("GetManagedObjects: %w", err)
	}
	for path, ifaces := range result {
		if _, ok := ifaces[bluezAdapter]; ok {
			return path, nil
		}
	}
	return "", fmt.Errorf("no Bluetooth adapter found — is bluetoothd running?")
}

func devicePath(adapterPath dbus.ObjectPath, mac string) dbus.ObjectPath {
	return dbus.ObjectPath(string(adapterPath) + "/dev_" + macToPath(mac))
}

func connectDevice(conn *dbus.Conn, devPath dbus.ObjectPath) error {
	obj := conn.Object(bluezService, devPath)

	v, err := obj.GetProperty(bluezDevice + ".Connected")
	if err == nil {
		if connected, ok := v.Value().(bool); ok && connected {
			fmt.Println("[OK] Already connected")
			return nil
		}
	}

	fmt.Println("[*] Connecting to", DEVICE_MAC, "...")
	call := obj.Call(bluezDevice+".Connect", 0)
	if call.Err != nil {
		return fmt.Errorf("Connect: %w", call.Err)
	}

	for i := 0; i < 15; i++ {
		time.Sleep(500 * time.Millisecond)
		v, err := obj.GetProperty(bluezDevice + ".Connected")
		if err == nil {
			if connected, ok := v.Value().(bool); ok && connected {
				fmt.Println("[OK] Connected")
				return nil
			}
		}
	}
	return fmt.Errorf("device did not become connected within timeout")
}

func disconnectAudio() {
	fmt.Println("[SYS] Disconnecting audio profile...")
	exec.Command("bluetoothctl", "disconnect", DEVICE_MAC).Run()
	time.Sleep(2 * time.Second)
}

func reconnectAudio() {
	fmt.Println("[SYS] Reconnecting audio profile...")
	exec.Command("bluetoothctl", "connect", DEVICE_MAC).Run()
	time.Sleep(3 * time.Second)
}
