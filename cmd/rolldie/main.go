package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"regexp"
	"strconv"

	"git.wntrmute.dev/kyle/goutils/die"
)

var dieRollFormat = regexp.MustCompile(`^(\d+)[dD](\d+)$`)

func rollDie(count, sides int) []int {
	sum := 0
	var rolls []int

	for range count {
		roll := rand.IntN(sides) + 1
		sum += roll
		rolls = append(rolls, roll)
	}

	rolls = append(rolls, sum)
	return rolls
}

func main() {
	flag.Parse()

	for _, arg := range flag.Args() {
		if !dieRollFormat.MatchString(arg) {
			fmt.Fprintf(os.Stderr, "invalid die format %s: should be XdY\n", arg)
			os.Exit(1)
		}

		dieRoll := dieRollFormat.FindAllStringSubmatch(arg, -1)
		count, err := strconv.Atoi(dieRoll[0][1])
		die.If(err)

		sides, err := strconv.Atoi(dieRoll[0][2])
		die.If(err)

		fmt.Println(rollDie(count, sides))
	}
}
