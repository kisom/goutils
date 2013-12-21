package modhex

import "bytes"
import "fmt"
import "testing"

func TestInvalidEncoder(t *testing.T) {
	s := ""
	for i := 0; i < 16; i++ {
		if NewEncoding(s) != nil {
			fmt.Println("modhex: NewEncoding accepted bad encoding")
			t.FailNow()
		}
	}
}

var encodeTests = []struct {
	In  []byte
	Out []byte
}{
	{[]byte{0x47}, []byte("fi")},
	{[]byte{0xba, 0xad, 0xf0, 0x0d}, []byte("nlltvcct")},
}

var decodeFail = [][]byte{
	[]byte{0x47},
	[]byte("abcdef"),
}

func TestStdEncodingEncode(t *testing.T) {
	enc := StdEncoding
	for _, et := range encodeTests {
		out := make([]byte, EncodedLen(len(et.In)))
		enc.Encode(out, et.In)
		if !bytes.Equal(out, et.Out) {
			fmt.Println("modhex: StdEncoding: bad encoding")
			fmt.Printf("\texpected: %x\n", et.Out)
			fmt.Printf("\t  actual: %x\n", out)
			t.FailNow()
		}
	}
}

func TestStdDecoding(t *testing.T) {
	enc := StdEncoding
	for _, et := range encodeTests {
		out := make([]byte, DecodedLen(len(et.Out)))
		n, err := enc.Decode(out, et.Out)
		if err != nil {
			fmt.Printf("%v\n", err)
			t.FailNow()
		} else if n != len(et.In) {
			fmt.Println("modhex: bad decoded length")
			t.FailNow()
		} else if !bytes.Equal(out, et.In) {
			fmt.Println("modhex: StdEncoding: bad decoding")
			fmt.Printf("\texpected: %x\n", et.In)
			fmt.Printf("\t  actual: %x\n", out)
			t.FailNow()
		}
	}
}

func TestStdDecodingFail(t *testing.T) {
	enc := StdEncoding
	for _, et := range decodeFail {
		dst := make([]byte, DecodedLen(len(et)))
		_, err := enc.Decode(dst, et)
		if err == nil {
			fmt.Println("modhex: decode should fail")
			t.FailNow()
		}
	}
}

func TestStdEncodingToString(t *testing.T) {
	enc := StdEncoding
	for _, et := range encodeTests {
		out := enc.EncodeToString(et.In)
		if out != string(et.Out) {
			fmt.Println("modhex: StdEncoding: bad encoding")
			fmt.Printf("\texpected: %x\n", et.Out)
			fmt.Printf("\t  actual: %x\n", out)
			t.FailNow()
		}
	}
}

func TestStdEncodingString(t *testing.T) {
	enc := StdEncoding
	for _, et := range encodeTests {
		out, err := enc.DecodeString(string(et.Out))
		if err != nil {
			fmt.Printf("%v\n", err)
			t.FailNow()
		} else if !bytes.Equal(out, et.In) {
			fmt.Println("modhex: StdEncoding: bad encoding")
			fmt.Printf("\texpected: %x\n", et.In)
			fmt.Printf("\t  actual: %x\n", out)
			t.FailNow()
		}
	}
}

var corruptTests = []struct {
	In      []byte
	Written int64
	Error   string
}{
	{[]byte("aa"), 0, "modhex: corrupt input at byte 0"},
	{[]byte("ca"), 0, "modhex: corrupt input at byte 0"},
	{[]byte("ccac"), 1, "modhex: corrupt input at byte 1"},
	{[]byte("ccca"), 1, "modhex: corrupt input at byte 1"},
}

func TestCorruptInputError(t *testing.T) {
	enc := StdEncoding
	for _, ct := range corruptTests {
		dst := make([]byte, DecodedLen(len(ct.In)))
		n, err := enc.Decode(dst, ct.In)
		if err == nil {
			fmt.Println("modhex: decode should fail")
			t.FailNow()
		} else if (err.(CorruptInputError)).Written() != ct.Written {
			fmt.Printf("modhex: decode should fail at byte %d, failed at byte %d\n",
				ct.Written, (err.(CorruptInputError)).Written())
			t.FailNow()
		} else if err.Error() != ct.Error {
			fmt.Printf("modhex: invalid error '%s' returned\n", err.Error())
			fmt.Printf(" (expected '%s')\n", ct.Error)
			t.FailNow()
		} else if int64(n) != ct.Written {
			fmt.Printf("modhex: decode should fail at byte %d, failed at byte %d\n",
				ct.Written, n)
			t.FailNow()
		}

	}
}

func TestCorruptInputErrorString(t *testing.T) {
	enc := StdEncoding
	for _, ct := range corruptTests {
		_, err := enc.DecodeString(string(ct.In))
		if err == nil {
			fmt.Println("modhex: decode should fail")
			t.FailNow()
		} else if (err.(CorruptInputError)).Written() != ct.Written {
			fmt.Printf("modhex: decode should fail at byte %d, failed at byte %d\n",
				ct.Written, (err.(CorruptInputError)).Written())
			t.FailNow()
		} else if err.Error() != ct.Error {
			fmt.Printf("modhex: invalid error '%s' returned\n", err.Error())
			fmt.Printf(" (expected '%s')\n", ct.Error)
			t.FailNow()
		}
	}
}

func TestFoo(t *testing.T) {
	fmt.Println("Hello, world!->", StdEncoding.EncodeToString([]byte("Hello, world!")))
	t.FailNow()
}
