package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/goburrow/serial"
)

type config struct {
	Serial *serial.Config
	Debug  bool
}

type ledState [8]byte

type message [2]byte

// make some nice api for interacting with the thing, managing the LED state
//
// monome.LED(x,y).On() -> p.Write([]byte{0x20|1, (x<<4)|(y&0x0f)})
// monome.LED(x,y).Off() -> p.Write([]byte{0x20|0, (x<<4)|(y&0x0f)})
// monome.LED(x,y).Toggle() -> p.Write([]byte{0x20|(state^bit), (x<<4)|(y&0x0f)})
//
// monome.LEDBrightness(b) -> p.Write([]byte{0x30, b})
//
// monome.Clear() -> p.Write([]byte{0x40, 0})
// monome.LEDTest() -> p.Write([]byte{0x40, 0})
//
// monome.EnableADC(n) -> p.Write([]byte{0x50, n})
//
// monome.Shutdown() -> p.Write([]byte{0x60, 0})
//
// monome.Row(y).Set(b) -> p.Write([]byte{0x70|y, b})
//
// monome.Column(y).Set(b) -> p.Write([]byte{0x80|x, b})
//
// make some nice animation bits that are decoupled from reading and
// writing from the port
//

type messageHandler func(message, serial.Port, *ledState)

func readSerial(c *config, h messageHandler) error {
	p, err := serial.Open(c.Serial)
	if err != nil {
		return err
	}

	l := &ledState{}
	m := message{}
	b := message{}

	writeState(p, l)

	for {
		n, err := p.Read([]byte(b[:]))

		if err == serial.ErrTimeout {
			continue
		}

		if err != nil {
			return err
		}

		if n < 2 || n%2 != 0 {
			if c.Debug {
				log.Printf("Invalid message received: %v", b[:n])
			}
			continue
		}

		copy(m[:], b[:])
		h(m, p, l)
	}
}

func writeState(p serial.Port, l *ledState) {
	for i := byte(0); i < 8; i++ {
		p.Write([]byte{0x80 | i, l[i]})
	}
}

func clear(p serial.Port) {
	for i := byte(0); i < 8; i++ {
		p.Write([]byte{0x80 | i, 0})
	}
}

func debugger(m message, p serial.Port, l *ledState) {
	log.Printf("Debug: %02x%02x", m[0], m[1])
}

func lighter(m message, p serial.Port, l *ledState) {
	log.Printf("lighter: %02x%02x", m[0], m[1])

	if m[0] == 0 || m[0] == 1 {
		b := []byte{0x20 | m[0], m[1]}
		log.Printf("lighter writing: %02x%02x", b[0], b[1])
		p.Write(b)
	}
}

func toggler(m message, p serial.Port, l *ledState) {
	log.Printf("toggler: %02x%02x", m[0], m[1])

	if m[0] == 0 {
		x := (m[1] >> 4) & 0x0F
		log.Println("toggler: ", x)
		y := m[1] & 0x0F
		l[x] = l[x] ^ (1 << y)
		b := []byte{0x80 | x, l[x]}
		log.Printf("toggler writing: %02x%02x", b[0], b[1])
		p.Write(b)
	}
}

func sparkler(m message, p serial.Port, l *ledState) {
	for i := byte(0); i < 8; i++ {
		b := []byte{0x80 | i, byte(rand.Uint32() & 0xFF)}
		log.Printf("sparkler writing: %02x%02x", b[0], b[1])
		p.Write(b)
	}
}

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

	c := &config{
		Serial: &serial.Config{
			Address:  *name,
			BaudRate: int(*baud),
			DataBits: int(*dataBits),
			StopBits: int(*stopBits),
			Parity:   *parity,
			Timeout:  time.Duration(time.Duration(*timeout) * time.Millisecond),
		},
		Debug: *debug,
	}

	log.Fatal(readSerial(c, toggler))
}
