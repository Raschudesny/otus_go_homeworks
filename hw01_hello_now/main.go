package main

import (
	"fmt"
	"log"
	"time"

	"github.com/beevik/ntp"
)

func main() {
	systemTime := time.Now()
	ntpTime, err := ntp.Time("1.beevik-ntp.pool.ntp.org")
	if err != nil {
		log.Fatalf("Error during reading time from NTP server: %v", err)
	}

	fmt.Printf("current time: %v\n", systemTime.Round(time.Microsecond)) //nolint:forbidigo
	fmt.Printf("exact time: %v\n", ntpTime.Round(time.Microsecond))      //nolint:forbidigo
}
