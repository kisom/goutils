package main

import (
	"fmt"

	"github.com/kisom/goutils/logging"
	"github.com/kisom/testio"
)

var log = logging.Init()
var olog, _ = logging.New("subsystem #42", logging.LevelNotice)

func main() {
	exampleNewWriters()
	log.Notice("Hello, world.")
	log.Warning("this program is about to end")

	log.SetLevel(logging.LevelDebug)
	log.Debug("hello world")
	log.SetLevel(logging.LevelNotice)

	olog.Print("now online")
	logging.Suppress("olog")
	olog.Print("extraneous information")

	logging.Enable("olog")
	olog.Print("relevant now")

	logging.SuppressAll()
	log.Alert("screaming into the void")
	olog.Critical("can anyone hear me?")

	log.Enable()
	log.Notice("i'm baaack")
	log.Suppress()
	log.Warning("but not for long")

	logging.EnableAll()
	log.Notice("fare thee well")
	olog.Print("all good journeys must come to an end")
	exampleNewFromFile()
}

func exampleNewWriters() {
	o := testio.NewBufCloser(nil)
	e := testio.NewBufCloser(nil)

	wlog, _ := logging.NewFromWriters("writers", logging.DefaultLevel, o, e)
	wlog.Notice("hello, world")
	wlog.Notice("some more things happening")
	wlog.Warning("something suspicious has happened")
	wlog.Alert("pick up that can, Citizen!")

	fmt.Println("--- BEGIN OUT ---")
	fmt.Printf("%s", o.Bytes())
	fmt.Println("--- END OUT ---")

	fmt.Println("--- BEGIN ERR ---")
	fmt.Printf("%s", e.Bytes())
	fmt.Println("--- END ERR ---")
}

func exampleNewFromFile() {
	flog, err := logging.NewFromFile("file logger", logging.LevelNotice,
		"example.log", "example.err", true)
	if err != nil {
		log.Fatalf("failed to open logger: %v", err)
	}
	defer flog.Close()

	flog.Notice("hello, world")
	flog.Notice("some more things happening")
	flog.Warning("something suspicious has happened")
	flog.Alert("pick up that can, Citizen!")
}
