package logging

var log = Init()
var olog = New("subsystem #42", LevelNotice)

func Example() {
	log.Notice("Hello, world.")
	log.Warning("this program is about to end")

	olog.Print("now online")
	Suppress("olog")
	olog.Print("extraneous information")

	Enable("olog")
	olog.Print("relevant now")

	SuppressAll()
	log.Alert("screaming into the void")
	olog.Critical("can anyone hear me?")

	log.Enable()
	log.Notice("i'm baaack")
	log.Suppress()
	log.Warning("but not for long")

	EnableAll()
	log.Notice("fare thee well")
	olog.Print("all good journeys must come to an end")
}
