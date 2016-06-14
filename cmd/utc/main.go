package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	format  = "2006-01-02 15:04" // Format that will be used for times.
	tz      = "Local"            // String descriptor for timezone.
	fromLoc *time.Location       // Go time.Location for the named timezone.
	toLoc   *time.Location       // Go time.Location for output timezone.
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: utc [-f format] [-u] [-h] [-z zone] [time(s)...]

utc converts times to UTC. If no arguments are provided, prints the
current time in UTC.

Flags:

	-f format	Go timezone format. See the Go documentation
			(e.g. https://golang.org/pkg/time/#pkg-constants)
			for an explanation of this format.

			Default value: %s

	-h		Print this help message.

	-u		Timestamps are in UTC format and should be converted
			to the timezone specified by the -z argument (which
			defaults to '%s'). Note that this isn't particularly
			useful with no arguments.

	-z zone		Text form of the time zone; this can be in short
			time zone abbreviation (e.g. MST) or a location
			(e.g. America/Los_Angeles). This has no effect when
			printing the current time.

			Default value: %s

Examples (note that the examples are done in the America/Los_Angeles /
PST8PDT time zone):

	+ Getting the current time in UTC:
	  $ utc
	  2016-06-14 14:30 = 2016-06-14 21:30
	+ Converting a local timestamp to UTC:
	  $ utc '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-15 04:30
	+ Converting a local EST timestamp to UTC (on a machine set to
  	  PST8PDT):
	  $ utc -z EST '2016-06-14 21:30'  
	  2016-06-14 21:30 = 2016-06-15 02:30
	+ Converting timestamps in the form '14-06-2016 3:04PM':
	  $ utc -f '02-01-2006 3:04PM' '14-06-2016 9:30PM'
	  14-06-2016 9:30PM = 15-06-2016 4:30AM
	+ Converting timestamps from standard input:
	  $ printf "2016-06-14 14:42\n2016-06-13 11:01" | utc -
	  2016-06-14 14:42 = 2016-06-14 21:42
	  2016-06-13 11:01 = 2016-06-13 18:01
	+ Converting a UTC timestamp to the local time zone:
	  $ utc -u '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-14 14:30
	+ Converting a UTC timestamp to EST (on a machine set to
	  PST8PDT):
	  $ utc -u -z EST '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-14 16:30

`, format, tz, tz)
}

func init() {
	var help bool
	var utc bool

	flag.Usage = func() { usage(os.Stderr) }
	flag.StringVar(&format, "f", format, "time format")
	flag.BoolVar(&help, "h", false, "print usage information")
	flag.BoolVar(&utc, "u", false, "timestamps are in UTC format")
	flag.StringVar(&tz, "z", tz, "time zone to convert from; if blank, the local timezone is used")

	flag.Parse()

	if help {
		usage(os.Stdout)
		os.Exit(0)
	}

	if utc {
		var err error
		toLoc, err = time.LoadLocation(tz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Malformed timezone %s: %s\n", tz, err)
			os.Exit(1)
		}

		fromLoc = time.UTC
	} else {
		var err error
		fromLoc, err = time.LoadLocation(tz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Malformed timezone %s: %s\n", tz, err)
			os.Exit(1)
		}

		toLoc = time.UTC
	}
}

func showTime(t time.Time) {
	fmt.Printf("%s = %s\n", t.Format(format), t.In(toLoc).Format(format))
}

func dumpTimes(times []string) bool {
	var errored bool

	for _, t := range times {
		u, err := time.ParseInLocation(format, t, fromLoc)
		if err != nil {
			errored = true
			fmt.Fprintf(os.Stderr, "Malformed time %s: %s\n", t, err)
			continue
		}

		showTime(u)
	}

	return errored
}

func main() {
	var times []string
	n := flag.NArg()

	switch n {
	case 0:
		showTime(time.Now())
		os.Exit(0)
	case 1:
		if flag.Arg(0) == "-" {
			s := bufio.NewScanner(os.Stdin)

			for s.Scan() {
				times = append(times, s.Text())
			}
		} else {
			times = flag.Args()
		}
	default:
		times = flag.Args()
	}

	if dumpTimes(times) {
		os.Exit(1)
	}
}
