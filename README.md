# Poltergeist

Poltergeist is an evil spirit from German folklore, known for its invisible hands used to control or destroy objects.

Inspired by this concept, **Poltergeist** is a specialized Linux utility that repurposes the touch-sensitive surfaces of OnePlus earbuds (specifically Nord Buds 3) as a remote control for a Linux desktop.

By intercepting raw Bluetooth RFCOMM packets, the system translates physical earbud gestures >> such as double taps, triple taps, and long presses >> into desktop actions via `xdotool`.

---

##  Features

-  Control your laptop with touch gestures from your earbuds  
-  Map gestures to actions like mouse clicks, scrolling, and media controls  
-  Control Active Noise Cancellation (ANC) modes  
-  Query battery status of earbuds and charging case  
-  Single binary written in Go  

---

##  Command List

### `listen`
The primary mode for Poltergeist.

- Performs protocol handshake
- Enters infinite loop waiting for gesture packets

**Behavior:**
- Calls `doHandshake()` >> sends `HELLO` and `REGISTER` packets  
- Uses `readLoop` to capture notifications  

**Expected Output:**
* Expected Output:
- [*] Performing handshake...
- [TX] HELLO: aa070000000123000012
- [TX] REGISTER: aa0c0000008541050000b550a069
- [*] Waiting for gestures... (Press Ctrl+C to stop)



### `on / off / trans`
Controls ANC state.

**Implementation:**  
Wraps `cmdANC(fd, mode, label)`

**Modes:**
- `on` → `0x01` (ANC Enabled)  
- `off` → `0x00` (ANC Disabled)  
- `trans` → `0x02` (Transparency Mode)  

**Data Flow:**
- Sends packet with `CAT=0x04` and `SUB=0x04`

---

### `battery`
Queries battery levels.

**Implementation:**
- Calls `cmdBattery(fd)`

**Data Flow:**
- Sends `CAT=0x06`, `SUB=0x01`
- Parsed when `cat == 0x06 && sub == 0x81`

**Expected Output:**
- [BATTERY] L:85% R:85% C:100%

## Connection and Handshake Flow :
<img width="960" height="380" alt="image" src="https://github.com/user-attachments/assets/82e1c3e1-6bb6-4d1e-8dfd-8f8789c4074d" />

## Gesture to Desktop Action Pipeline
<img width="674" height="159" alt="image" src="https://github.com/user-attachments/assets/32484fa3-e4a6-4c87-82d1-0bda1e5a41b3" />


## Packet Dispatch Logic
<img width="952" height="389" alt="image" src="https://github.com/user-attachments/assets/c0de2aed-ff7f-4b66-83ef-f47806bb497a" />



## Features

*   Control your laptop with touch gestures from your earbuds.
*   Map gestures to actions like mouse clicks, scrolling, and media key presses.
*   Control Active Noise Cancellation (ANC) modes.
*   Query the battery status of the earbuds and case.
*   Single binary, written in Go.

## Prerequisites

*   A Linux-based operating system.
*   Go (version 1.25.7 or later).
*   `xdotool`: A command-line tool for simulating keyboard input and mouse activity.
    ```sh
    # On Debian/Ubuntu
    sudo apt-get install xdotool
    ```
*   OnePlus Nord Buds (tested with model `9C:DE:F0:33:F5:DD`). You may need to change the `DEVICE_MAC` constant in `main.go` for your device.

## Installation & Setup

1.  Clone the repository:
    ```sh
    git clone https://github.com/tanice-gawd/poltergeist.git
    cd poltergeist
    ```
2.  Pair and connect your OnePlus Nord Buds to your computer using your system's Bluetooth manager. You can verify the connection with `bluetoothctl`.

3.  Update the `DEVICE_MAC` constant in `main.go` with your earbud's MAC address if it's different.

## Usage

Run the program from the command line, specifying the desired command.

```sh
go run main.go <command>
```

### Commands

| Command      | Description                                          |
|--------------|------------------------------------------------------|
| `listen`     | **(Main mode)** Connect and listen for touch gestures. |
| `on`         | Enable ANC (Active Noise Cancellation).              |
| `off`        | Disable ANC.                                         |
| `trans`      | Enable Transparency mode.                            |
| `battery`    | Query and display battery levels.                    |
| `connect`    | Connect the earbuds using `bluetoothctl`.            |
| `disconnect` | Disconnect the earbuds using `bluetoothctl`.         |

### Listening for Gestures

To start the main gesture listening service, run:
```sh
go run main.go listen
```

The program will connect and display the following table, indicating it's ready to receive commands.

```
╔══════════════════════════════════════╗
║     Nord Buds Gesture Remote READY   ║
╠══════════════════════════════════════╣
║  L double tap  →  Left click         ║
║  R double tap  →  Right click        ║
║  L triple tap  →  Scroll up          ║
║  R triple tap  →  Scroll down        ║
║  L long press  →  Play / Pause       ║
║  R long press  →  Next track         ║
╚══════════════════════════════════════╝
