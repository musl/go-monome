package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/goburrow/serial"
	"github.com/musl/go-monome"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix(` `)

	name := flag.String(`name`, `/dev/tty.usbserial`, `The path to the serial port.`)
	baud := flag.Uint(`baud`, 9600, `Serial Port Baud Rate`)
	timeout := flag.Uint(`timeout`, 1, `Serial Port Read Timeout in milliseconds`)
	parity := flag.String(`parity`, "E", `Serial Port Parity: 'N', 'O', 'E', 'M', or 'S'`)
	dataBits := flag.Uint(`dataBits`, 8, `Serial Data Bits: 5, 6, 7, or 8 (default 8)`)
	stopBits := flag.Uint(`stopBits`, 1, `Serial Port Stop Bits: '1' or '2' (default 1)`)
	debug := flag.Bool(`debug`, false, `Show debugging logs`)
	flag.Parse()

	serialConfig := &serial.Config{
		Address:  *name,
		BaudRate: int(*baud),
		DataBits: int(*dataBits),
		StopBits: int(*stopBits),
		Parity:   *parity,
		Timeout:  time.Duration(time.Duration(*timeout) * time.Millisecond),
	}

	m := monome.NewMonomeConfig(serialConfig, *debug)

	m.ButtonChanged(func(m *Monome, x, y, s uint) {
		err := m.LED(x, y).Toggle()
		if err != nil {
			log.Println(err)
		}
	})

	log.Fatal(m.Loop())
}
