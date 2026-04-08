# Poltergeist
[![Ask DeepWiki](https://devin.ai/assets/askdeepwiki.png)](https://deepwiki.com/TANICE-GAWD/Poltergeist)

Poltergeist is a evil spirit from German Folk lore in which it had invisible hands and it used it to control/destroy objects

Poltergeist is a specialized Linux utility designed to repurpose the touch-sensitive surfaces of OnePlus earbuds (specifically Nord Buds 3) as a remote control for a Linux desktop. By intercepting raw Bluetooth RFCOMM packets, the system translates physical earbud gestures—such as double taps, triple taps, and long presses—into desktop actions via xdotool.



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
