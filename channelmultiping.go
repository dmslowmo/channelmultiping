package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type HostStatus struct {
	hostName string
	isUp     bool
}

func ping(c chan HostStatus, host string) {
	var err error
	var hs HostStatus
	cmdName := "ping"
	cmdArgs := []string{"-c3", "-w3", host}

	hs.hostName = host
	if _, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		hs.isUp = false
	} else {
		hs.isUp = true
	}
	c <- hs
}

/*
Goroutines are created as many as there are hosts to ping.
The results are stored in channels and are to be popped up later for
printing to the console.
The number of channels may be as little as 1.
*/
func main() {
	// get the starting time
	startTime := time.Now().UTC()

	// read the input file
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inFile.Close() // close the file when main exits
	scanner := bufio.NewScanner(inFile)

	//use channels
	numOfGr := 1 //hardcoded for now
	chans := make([]chan HostStatus, numOfGr)
	for i := range chans {
		chans[i] = make(chan HostStatus)
		fmt.Println(i+1, "channels created")
	}

	go func() {
		i := 0
		for scanner.Scan() {
			host := scanner.Text()
			go ping(chans[i], host)
			i = (i+1)%numOfGr
		}
	}()
	for i := 0; i < 254; i++ { //hardcoded for now
		res := <- chans[i%numOfGr]
		fmt.Println(res.hostName, "is", res.isUp)
	}
	// get the ending time and calculate the total duration
	endTime := time.Now().UTC()
	fmt.Println(endTime.Sub(startTime))
}
