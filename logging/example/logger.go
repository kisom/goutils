package main

import "github.com/kisom/goutils/logging"

var log = logging.Init()
var olog = logging.New("subsystem #42", logging.LevelNotice)

func main() {
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
