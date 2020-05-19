package main

import (
	"flag"
	"log"
	"os"

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

	log.Fatal(m.Loop())
}
