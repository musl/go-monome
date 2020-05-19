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
	m := monomes[0]

	m.ButtonChanged(func(m *monome.Monome, x, y, s uint) error {
		return m.LED(x, y).Toggle()
	})

	m.Open()
	defer m.Close()

	log.Fatal(m.Loop())
}
