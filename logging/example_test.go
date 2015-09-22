package logging_test

import "github.com/kisom/goutils/logging"

var log = logging.Init()
var olog, _ = logging.New("subsystem #42", logging.LevelNotice)

func Example() {
	log.Notice("Hello, world.")
	log.Warning("this program is about to end")

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
}

func ExampleNewFromFile() {
	log, err := logging.NewFromFile("file logger", logging.LevelNotice,
		"example.log", "example.err", true)
	if err != nil {
		log.Fatalf("failed to open logger: %v", err)
	}

	log.Notice("hello, world")
	log.Notice("some more things happening")
	log.Warning("something suspicious has happened")
	log.Alert("pick up that can, Citizen!")
}
