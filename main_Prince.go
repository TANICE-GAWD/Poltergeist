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


// OPO PROTOCOL — HANDSHAKE


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

