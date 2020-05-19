package monome

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/goburrow/serial"
)

// Monome is the type that looks after the configuration, serial port,
// and LED state of a given monome 40h.
type Monome struct {
	serialConfig *serial.Config
	serialPort   serial.Port
	ledState     [8]byte

	adcHandlers    []ADCHandler
	buttonHandlers []ButtonHandler

	Debug bool
}

// ADCHandler is a function that can be run when an ADC message is received.
type ADCHandler func(*Monome, uint, uint) error

// ButtonHandler is a function that can be run when a Button message is received.
type ButtonHandler func(*Monome, uint, uint, uint) error

// Find returns a new monome for each matching device found under /dev/*
func Find(debug bool) (monomes []*Monome) {
	/*
	 * Right now, this only works on OSX where the device name is
	 * distinct.
	 *
	 * If there are other file globs that return reliable devices
	 * then we can add them here.
	 *
	 * Also, this needs support for Windows.
	 */
	globs := []string{
		`/dev/tty.usbserial-m40h*`,
	}

	for _, glob := range globs {
		paths, err := filepath.Glob(glob)
		if err != nil {
			if debug {
				log.Println(err.Error())
			}
			continue
		}

		for _, path := range paths {
			if debug {
				log.Println("Found monome at:", path)
			}
			monomes = append(monomes, NewMonome(path, debug))
		}
	}

	return
}

// NewMonome returns a new Monome with the given serial port path and
// default serial configuration.
func NewMonome(portPath string, debug bool) *Monome {
	return NewMonomeConfig(
		&serial.Config{
			Address:  portPath,
			BaudRate: 9600,
			DataBits: 8,
			StopBits: 1,
			Parity:   "E",
			Timeout:  1 * time.Millisecond,
		},
		debug,
	)
}

// NewMonomeConfig returns a new Monome with a given `*serial.Config`
func NewMonomeConfig(serialConfig *serial.Config, debug bool) *Monome {
	return &Monome{serialConfig: serialConfig, Debug: debug}
}

// Open attepmts to open the device.
func (m *Monome) Open() error {
	port, err := serial.Open(m.serialConfig)
	if err != nil {
		return err
	}
	m.serialPort = port
	return nil
}

// Loop reads continuously from the monome, calling a handler when
// messages are recieved.
func (m *Monome) Loop() error {
	err := m.WriteState()
	if err != nil {
		return err
	}

	b := [2]byte{}
	for {
		n, err := m.serialPort.Read(b[:])
		if err == serial.ErrTimeout {
			continue
		}
		if err != nil {
			return err
		}
		if n != 2 {
			if m.Debug {
				log.Printf("Invalid message received: %v", b[:n])
			}
			continue
		}

		err = m.handle(b)
		if err != nil && m.Debug {
			log.Println("Error in handler:", err.Error())
		}
	}
}

// parse and dispatch messages
func (m *Monome) handle(message [2]byte) error {
	switch message[0] >> 4 {
	case 0:
		x := (message[1] >> 4) & 0x0f
		y := message[1] & 0x0f
		s := message[0]
		for _, h := range m.buttonHandlers {
			return h(m, uint(x), uint(y), uint(s))
		}
	case 1:
		p := (message[0] >> 2) & 0x0c
		v := (uint(message[0]&0x3) << 8) | uint(message[1])
		for _, h := range m.adcHandlers {
			return h(m, uint(p), v)
		}
	}

	if m.Debug {
		log.Printf("Unknown message received: %v", message)
	}
	return nil
}

// ButtonChanged adds a Button handler
func (m *Monome) ButtonChanged(bh ButtonHandler) {
	m.buttonHandlers = append(m.buttonHandlers, bh)
}

// ADCChanged adds a ADC handler
func (m *Monome) ADCChanged(ah ADCHandler) {
	m.adcHandlers = append(m.adcHandlers, ah)
}

// Close attepmts to open the device.
func (m *Monome) Close() {
	m.serialPort.Close()
}

// Write writes a message to the given monome.
func (m *Monome) Write(message [2]byte) error {
	n, err := m.serialPort.Write(message[:])
	if err != nil {
		return err
	}
	if n != len(message) {
		return fmt.Errorf("Incomplete write")
	}

	if m.Debug {
		log.Printf("Wrote: %v", message)
	}

	return nil
}

// Brightness sets the brightness of all LEDs given by a float between 0
// and 1 inclusive. 0 is off, 1.0 is full brightness.
func (m *Monome) Brightness(v float64) error {
	if v > 1.0 {
		v = 1.0
	}
	if v < 0.0 {
		v = 0.0
	}
	return m.Write([2]byte{0x30, byte(0xff * v)})
}

// Clear discards the LED state and turns off all LEDs.
func (m *Monome) Clear() error {
	m.ledState = [8]byte{}
	return m.WriteState()
}

// ADC represents a ADC input.
type ADC struct {
	parent *Monome
	n      byte
}

// ADC returns a new ADC
func (m *Monome) ADC(n uint) *ADC {
	return &ADC{parent: m, n: byte(n) & 0x0f}
}

// Enable turns a given ADC on.
func (a *ADC) Enable() error {
	return a.parent.Write([2]byte{0x50, ((a.n & 0x7) << 4) | 1})
}

// Disable turns a given ADC off.
func (a *ADC) Disable() error {
	return a.parent.Write([2]byte{0x50, (a.n & 0x7) << 4})
}

// LED returns a given LED on the given Monome, clamping the values of x
// and y to 0-7.
func (m *Monome) LED(x, y uint) *LED {
	return &LED{x: byte(x) & 7, y: byte(y) & 7}
}

// LEDTest tells the device to turn on or off all of the LEDs without
// updating internal state.
func (m *Monome) LEDTest(on bool) error {
	if on {
		return m.Write([2]byte{0x40, 1})
	}

	return m.Write([2]byte{0x40, 0})
}

// Row returns a given Row of LEDs on the monome.
func (m *Monome) Row(y uint) *Row {
	return &Row{parent: m, y: byte(y) & 0x07}
}

// Shutdown turns the monome off.
func (m *Monome) Shutdown() error {
	return m.Write([2]byte{0x60, 0})
}

// WriteState updates all LEDs on the device to match the stored state.
func (m *Monome) WriteState() error {
	for i := byte(0); i < byte(len(m.ledState)); i++ {
		err := m.Write([2]byte{0x80 | i, m.ledState[i]})
		if err != nil {
			return err
		}
	}
	return nil
}

// LED represents a single LED on the monome.
type LED struct {
	parent *Monome
	x, y   byte
}

// On turns a given LED on
func (l *LED) On() error {
	return l.parent.Write([2]byte{0x21, (l.x << 4) | (l.y & 0x0f)})
}

// Off turns a given LED off
func (l *LED) Off() error {
	return l.parent.Write([2]byte{0x20, (l.x << 4) | (l.y & 0x0f)})
}

// Toggle turns a given LED on if it's off and vice-versa.
func (l *LED) Toggle() error {
	l.parent.ledState[l.x] = l.parent.ledState[l.x] ^ (1 << l.y)
	return l.parent.Write([2]byte{
		0x20 | ((l.parent.ledState[l.x] >> l.y) & 1),
		(l.x << 4) | (l.y & 0x0f),
	})
}

// Row reperesents a single row of LEDS on the monome.
type Row struct {
	parent *Monome
	y      byte
}

// On turns on all of the given LEDs in a row.
func (r *Row) On(b byte) error {
	return r.parent.Write([2]byte{0x80, 0xff})
}

// Off turns off all of the given LEDs in a row.
func (r *Row) Off(b byte) error {
	return r.parent.Write([2]byte{0x80, 0})
}

// Set sets all of the given LEDs in a row.
func (r *Row) Set(b byte) error {
	return r.parent.Write([2]byte{0x70, b})
}

// Column reperesents a single column of LEDS on the monome.
type Column struct {
	parent *Monome
	x      byte
}

// On turns on all of the given LEDs in a column.
func (c *Column) On(b byte) error {
	return c.parent.Write([2]byte{0x80, 0xff})
}

// Off turns off all of the given LEDs in a column.
func (c *Column) Off(b byte) error {
	return c.parent.Write([2]byte{0x80, 0})
}

// Set sets all of the given LEDs in a column.
func (c *Column) Set(b byte) error {
	return c.parent.Write([2]byte{0x80, b})
}
