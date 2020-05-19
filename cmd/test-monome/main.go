package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/musl/go-monome"
)

func main() {
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

	m.Clear()
	for i := uint(0); i < 7; i++ {
		m.LED(i, i).On()
	}
	time.Sleep(100 * time.Millisecond)
	m.Clear()

	for i := uint(0); i < 7; i++ {
		m.Row(i).On()
		time.Sleep(100 * time.Millisecond)
		m.Row(i).Off()
		time.Sleep(100 * time.Millisecond)

		m.Column(i).On()
		time.Sleep(100 * time.Millisecond)
		m.Column(i).Off()
		time.Sleep(100 * time.Millisecond)
	}

	log.Fatal(m.Loop())
}
