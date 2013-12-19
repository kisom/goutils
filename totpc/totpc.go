package main

import (
	"fmt"
	"github.com/gokyle/twofactor"
	"io/ioutil"
	"time"
)

func main() {
	otp := twofactor.GenerateGoogleTOTP()
	if otp == nil {
		fmt.Println("totpc: failed to generate token")
		return
	}

	qr, err := otp.QR("totpc-demo")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = ioutil.WriteFile("out.png", qr, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(otp.OTP())
	for {
		for {
			t := time.Now()
			if t.Second() == 0 {
				break
			} else if t.Second() == 30 {
				break
			}
			<-time.After(1 * time.Second)
		}
		fmt.Println(otp.OTP())
		<-time.After(30 * time.Second)
	}
}
