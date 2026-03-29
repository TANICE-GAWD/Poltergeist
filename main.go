package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
	"net"
)

const (
	DEVICE_MAC    = "9C:DE:F0:33:F5:DD" 
	RFCOMM_CHAN   = 12      // discovered via Wireshark T_T ...finallyyy
	AF_BLUETOOTH  = 31
	BTPROTO_RFCOMM = 3
)
// BLUETOOTH / RFCOMM LAYER
// took this from C >> <bluetooth/rfcomm.h>

type sockaddrRFCOMM struct {
	Family  uint16
	Dev     [6]byte
	Channel uint8
	_       [1]byte
}


func parseMac(mac string) [6]byte{
	v, _ := net.ParseMAC(mac)
	var b [6]byte
	for i := 0; i<6 ; i++{
		b[i] = v[5-i]
	}
	return b
}

func connectRFCOMM() (int, error) {
	fd, err := syscall.Socket(AF_BLUETOOTH, syscall.SOCK_STREAM, BTPROTO_RFCOMM)
	if err != nil {
		return 0, fmt.Errorf("socket: %w", err)
	}

	addr := sockaddrRFCOMM{
		Family:  AF_BLUETOOTH,
		Dev:     parseMac(DEVICE_MAC),
		Channel: RFCOMM_CHAN,
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_CONNECT,
		uintptr(fd),
		uintptr(unsafe.Pointer(&addr)),
		unsafe.Sizeof(addr),
	)
	if errno != 0 {
		syscall.Close(fd)
		return 0, fmt.Errorf("connect errno: %w", errno)
	}

	return fd, nil
}

// PACKET HELPERS

func hexStr(data []byte) string {
    return hex.EncodeToString(data)
}

func sendRaw(fd int, data []byte, label string) {
	fmt.Printf("[TX] %s: %s\n", label, hexStr(data))
	_, err := syscall.Write(fd, data)
	if err != nil {
		fmt.Println("[ERROR] Write:", err)
		return
	}
}

func doHandshake(fd int) {
	sendRaw(fd, []byte{
		0xAA, 0x07, 0x00, 0x00, 0x00, 0x01, 0x23, 0x00, 0x00, 0x12,
	}, "HELLO")
	time.Sleep(2 * time.Second)

	sendRaw(fd, []byte{
		0xAA, 0x0C, 0x00, 0x00, 0x00, 0x85, 0x41, 0x05,
		0x00, 0x00, 0xB5, 0x50, 0xA0, 0x69,
	}, "REGISTER")
	time.Sleep(2 * time.Second)
}

// Gesture packet structure (CAT=0x04 SUB=0x02):
//
//  Byte  0    = 0xAA  (start of frame)
//  Byte  1    = LEN
//  Byte  2-3  = 0x00 0x00 (padding)
//  Byte  4    = 0x04  (CAT: ANC/gesture subsystem)
//  Byte  5    = 0x02  (SUB: gesture event)
//  Byte  6    = SEQ   (sequence number)
//  Byte  7    = payload length
//  Byte  8    = 0x00
//  Byte  9    = 0xF1  (gesture marker)
//  Byte  10   = BUD   (0x01=Left, 0x02=Right)
//  Byte  11   = 0x01
//  Byte  12   = GESTURE (0x02=double tap, 0x03=triple tap, 0x04=long press)

func handleNotify(data []byte) {
	fmt.Println("[RX RAW]", hexStr(data))

	for i := 0; i < len(data)-12; i++ {
		if data[i] != 0xAA {
			continue
		}

		frameLen := int(data[i+1])
		if i+frameLen+2 > len(data) {
			break
		}

		cat, sub := data[i+4], data[i+5]

		switch {
		// Gesture
		case cat == 0x04 && sub == 0x02 && data[i+9] == 0xF1:
			side := "LEFT"
			if data[i+10] == 0x02 {
				side = "RIGHT"
			}
			names := map[byte]string{
				0x02: "DOUBLE TAP",
				0x03: "TRIPLE TAP",
				0x04: "LONG PRESS",
			}
			gesture := data[i+12]
			name, ok := names[gesture]
			if !ok {
				name = fmt.Sprintf("UNKNOWN(0x%02X)", gesture)
			}

			fmt.Printf("\n╔══════════════════════════════╗\n")
			fmt.Printf("║  GESTURE: %-8s %s\n", side, name)
			fmt.Printf("╚══════════════════════════════╝\n")

			handleGesture(side, gesture)

		// Battery
		case cat == 0x06 && sub == 0x81:
			if i+15 >= len(data) {
				break
			}
			fmt.Printf("[PROTO] CAT=0x%02X SUB=0x%02X LEN=%d\n", cat, sub, len(data))
			fmt.Printf("\n[BATTERY] L:%d%% R:%d%% C:%d%%\n",
				data[i+12], data[i+14], data[i+15])
		}

		i += frameLen + 3
	}
}

func handleGesture(side string, gesture byte) {
	switch {
	case side == "LEFT" && gesture == 0x02:
		run("xdotool", "click", "1")
		fmt.Println("[ACTION] Left mouse click")

	case side == "RIGHT" && gesture == 0x02:
		run("xdotool", "click", "3")
		fmt.Println("[ACTION] Right mouse click")

	case side == "LEFT" && gesture == 0x03:
		run("xdotool", "click", "--repeat", "5", "--delay", "20", "4")
		fmt.Println("[ACTION] Scroll up ×5")

	case side == "RIGHT" && gesture == 0x03:
		run("xdotool", "click", "--repeat", "5", "--delay", "20", "5")
		fmt.Println("[ACTION] Scroll down ×5")

	case side == "LEFT" && gesture == 0x04:
		run("xdotool", "key", "XF86AudioPlay")
		fmt.Println("[ACTION] Play / Pause")

	case side == "RIGHT" && gesture == 0x04:
		run("xdotool", "key", "XF86AudioNext")
		fmt.Println("[ACTION] Next track")
	}
}

func run(name string, args ...string) {
	_ = exec.Command(name, args...).Run()
}


func readLoop(fd int) {
	buf := make([]byte, 1024)
	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			fmt.Println("\n[ERROR] Read:", err)
			return
		}
		if n > 0 {
			handleNotify(buf[:n])
		}
	}
}


func cmdANC(fd int, mode byte, label string) {
	doHandshake(fd)
	sendRaw(fd, []byte{
		0xAA, 0x0A, 0x00, 0x00, 0x04, 0x04,
		0x42, 0x03, 0x00, 0x01, 0x01, mode,
	}, label)
	time.Sleep(2 * time.Second)
}

func cmdBattery(fd int) {
	doHandshake(fd)
	sendRaw(fd, []byte{
		0xAA, 0x07, 0x00, 0x00, 0x06, 0x01, 0x25, 0x00, 0x00,
	}, "BATTERY QUERY")
	time.Sleep(3 * time.Second)
}


func cmdConnect() {
	fmt.Printf("[*] Linking %s... ", DEVICE_MAC)
	exec.Command("bluetoothctl", "connect", DEVICE_MAC).Run()

	fmt.Println("\n[OK] Device Connected")
}

func cmdDisconnect() {
	fmt.Printf("[*] Dropping %s... ", DEVICE_MAC)
	exec.Command("bluetoothctl", "disconnect", DEVICE_MAC).Run()
	fmt.Println("\n[OK] Device disconnected")
}




func printHelp() {
	fmt.Println(`
OnePlus Nord Buds 3 — Gesture Remote Controller

Usage:
  go run main.go <command>

Commands:
  listen    	Connect and listen for tap gestures (main mode)
  on        	Enable ANC (Active Noise Cancellation)
  off       	Disable ANC
  trans     	Enable Transparency mode
  battery   	Query battery levels
  connect	Connect the earbuds
  disconnect	Disconnect the earbuds

Gesture map:
  Left  double tap   →  Left mouse click
  Right double tap   →  Right mouse click
  Left  triple tap   →  Scroll up
  Right triple tap   →  Scroll down
  Left  long press   →  Play / Pause
  Right long press   →  Next track
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}
	cmd := strings.ToLower(os.Args[1])

	if cmd == "help" || cmd == "--help" || cmd == "-h" {
		printHelp()
		return
	}

	fmt.Printf("[*] Connecting to %s via RFCOMM channel %d...\n", DEVICE_MAC, RFCOMM_CHAN)
	fd, err := connectRFCOMM()
	if err != nil {
		fmt.Println("[ERROR]", err)
		fmt.Println("[HINT]  Make sure the buds are connected: bluetoothctl connect", DEVICE_MAC)
		return
	}
	defer syscall.Close(fd)
	fmt.Println("[OK] RFCOMM connected")

	go readLoop(fd)
	time.Sleep(300 * time.Millisecond)

	switch cmd {

	case "listen":
		fmt.Println("[*] Performing handshake...")
		doHandshake(fd)
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════╗")
		fmt.Println("║     Nord Buds Gesture Remote READY   ║")
		fmt.Println("╠══════════════════════════════════════╣")
		fmt.Println("║  L double tap  →  Left click         ║")
		fmt.Println("║  R double tap  →  Right click        ║")
		fmt.Println("║  L triple tap  →  Scroll up          ║")
		fmt.Println("║  R triple tap  →  Scroll down        ║")
		fmt.Println("║  L long press  →  Play / Pause       ║")
		fmt.Println("║  R long press  →  Next track         ║")
		fmt.Println("╚══════════════════════════════════════╝")
		fmt.Println()
		select {} 

	case "on":
		cmdANC(fd, 0x01, "ANC ON")

	case "off":
		cmdANC(fd, 0x04, "ANC OFF")

	case "trans":
		cmdANC(fd, 0x02, "TRANSPARENCY")

	case "battery":
		cmdBattery(fd)
	case "connect":
		cmdConnect()
		return

	case "disconnect":
		cmdDisconnect()
		return

	default:
		fmt.Printf("[ERROR] Unknown command: %s\n", cmd)
		printHelp()
	}
}