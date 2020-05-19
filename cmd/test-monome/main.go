package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/musl/go-monome"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix(` `)

	debug := flag.Bool(`debug`, false, `Show debugging logs`)
	flag.Parse()

	monomes := monome.Find(*debug)
	if len(monomes) < 1 {
		log.Fatal("Couldn't find anything.")
	}

	m := monomes[0]

	m.ButtonChanged(func(m *monome.Monome, x, y, s uint) error {
		if s == 1 {
			return nil
		}
		return m.LED(x, y).Toggle()
	})

	m.Open()
	defer m.Close()

	for i := uint(0); i < 8; i++ {
		m.Row(i).On()
	}

	m.Brightness(0.1)

	for i := 0; i < 1000; i++ {
		x, y := uint(rand.Uint32())&0x0f, uint(rand.Uint32())&0x0f

		if rand.Float64() > 0.5 {
			m.LED(x, y).On()
			continue
		}

		m.LED(x, y).Off()
		time.Sleep(1 * time.Millisecond)
	}

	for i := uint(0); i < 8; i++ {
		m.Row(i).On()
		time.Sleep(10 * time.Millisecond)
		m.Column(i).On()
		time.Sleep(10 * time.Millisecond)

		m.Row(i).Off()
		time.Sleep(10 * time.Millisecond)
		m.Column(i).Off()
		time.Sleep(10 * time.Millisecond)
	}

	for i := uint(7); i > 0; i-- {
		m.Row(i).On()
		time.Sleep(10 * time.Millisecond)
		m.Column(i).On()
		time.Sleep(10 * time.Millisecond)

		m.Row(i).Off()
		time.Sleep(10 * time.Millisecond)
		m.Column(i).Off()
		time.Sleep(10 * time.Millisecond)
	}

	log.Fatal(m.Loop())
}
