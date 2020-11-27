package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.sr.ht/~kisom/goutils/die"
)

const dbVersion = "1"

var dbFile = filepath.Join(os.Getenv("HOME"), ".parts.json")
var partsDB = &database{Version: dbVersion}

type part struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Class       string `json:"class,omitempty"`
}

func (p part) String() string {
	return fmt.Sprintf("%s: %s", p.Name, p.Description)
}

type database struct {
	Version    string          `json:"version"`
	LastUpdate int64           `json:"json"`
	Parts      map[string]part `json:"parts"`
}

func help(w io.Writer) {
	fmt.Fprintf(w, `Usage:  parts [id] -- query the database for a part
	parts [-c class] [id] [description] -- store a part in the database

	Options:
		-f path		Path to parts database (default is
				%s).
          
`, dbFile)
}

func loadDatabase() {
	data, err := ioutil.ReadFile(dbFile)
	if err != nil && os.IsNotExist(err) {
		partsDB = &database{
			Version: dbVersion,
			Parts:   map[string]part{},
		}
		return
	}
	die.If(err)

	err = json.Unmarshal(data, partsDB)
	die.If(err)
}

func findPart(partName string) {
	partName = strings.ToLower(partName)
	for name, part := range partsDB.Parts {
		if strings.Contains(strings.ToLower(name), partName) {
			fmt.Println(part.String())
		}
	}
}

func writeDB() {
	data, err := json.Marshal(partsDB)
	die.If(err)

	err = ioutil.WriteFile(dbFile, data, 0644)
	die.If(err)
}

func storePart(name, class, description string) {
	p, exists := partsDB.Parts[name]
	if exists {
		fmt.Printf("warning: replacing part %s\n", name)
		fmt.Printf("\t%s\n", p.String())
	}

	partsDB.Parts[name] = part{
		Name:        name,
		Class:       class,
		Description: description,
	}

	writeDB()
}

func listParts() {
	parts := make([]string, 0, len(partsDB.Parts))
	for partName := range partsDB.Parts {
		parts = append(parts, partName)
	}

	sort.Strings(parts)
	for _, partName := range parts {
		fmt.Println(partsDB.Parts[partName].String())
	}
}

func main() {
	var class string
	var helpFlag bool

	flag.StringVar(&class, "c", "", "device class")
	flag.StringVar(&dbFile, "f", dbFile, "`path` to database")
	flag.BoolVar(&helpFlag, "h", false, "Print a help message.")
	flag.Parse()

	if helpFlag {
		help(os.Stdout)
		return
	}

	loadDatabase()

	switch flag.NArg() {
	case 0:
		help(os.Stdout)
		return
	case 1:
		partName := flag.Arg(0)
		if partName == "list" {
			listParts()
		} else {
			findPart(flag.Arg(0))
		}
		return
	default:
		description := strings.Join(flag.Args()[1:], " ")
		storePart(flag.Arg(0), class, description)
		return
	}
}
