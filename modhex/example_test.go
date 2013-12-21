package modhex

import (
	"fmt"
	"github.com/gokyle/twofactor/modhex"
)

var out = "fjhghrhrhvdrdciihvidhrhfdb"
var in = "Hello, world!"

func ExampleEncoding_EncodeToString() {
	data := []byte("Hello, world!")
	str := modhex.StdEncoding.EncodeToString(data)
	fmt.Println(str)
	// Output:
	// fjhghrhrhvdrdciihvidhrhfdb
}

func ExampleEncoding_DecodeString() {
	str := "fjhghrhrhvdrdciihvidhrhfdb"
	data, err := modhex.StdEncoding.DecodeString(str)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("%s", string(data))
	// Output:
	// Hello, world!
}
