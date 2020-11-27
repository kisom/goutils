package main

import (
	"os"
	"time"

	"git.sr.ht/~kisom/goutils/logging"
)

var log = logging.NewConsole()
var olog = logging.NewConsole()

func main() {
	log.Info("example", "Hello, world.", nil)
	log.Warn("example", "this program is about to end", nil)

	log.Critical("example", "screaming into the void", nil)
	olog.Critical("other", "can anyone hear me?", nil)

	log.Warn("example", "but not for long", nil)

	log.Info("example", "fare thee well", nil)
	olog.Info("other", "all good journeys must come to an end",
		map[string]string{"when": time.Now().String()})

	log.Info("example", "filelog test", nil)
	exampleNewFromFile()
	os.Remove("example.log")
	os.Remove("example.err")
}

func exampleNewFromFile() {
	flog, err := logging.NewSplitFile("example.log", "example.err", true)
	if err != nil {
		log.Fatal("filelog", "failed to open logger",
			map[string]string{"error": err.Error()})
	}

	flog.Info("filelog", "hello, world", nil)
	flog.Info("filelog", "some more things happening", nil)
	flog.Warn("filelog", "something suspicious has happened", nil)
	flog.Critical("filelog", "pick up that can, Citizen!", nil)
}
