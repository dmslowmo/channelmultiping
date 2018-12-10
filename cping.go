package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type HostStatus struct {
	hostName string
	isUp     bool
}

func ping(result chan HostStatus, host string) {
	var err error
	var hs HostStatus
	cmdName := "ping"
	cmdArgs := []string{"-c3", "-t3", host}

	hs.hostName = host
	if _, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		hs.isUp = false
	} else {
		hs.isUp = true
	}
	result <- hs // store result in results channel
}

/**
Goroutines are created as many as there are hosts to ping by the Producer goroutine.
Each goroutine puts the ping result in a queue, implemented by buffered channel.

We limit the size of the results channel (from user input), with default size as 1.

The results are consumed by the Consumer goroutine.

WaitGroup is used such that the main thread will wait for the Producer and Consumer
goroutines to complete before exiting.

Command line usage: ./channelmultiping <hostlist.txt> [size_of_result_queue]
**/
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

	// create buffered channels
	numOfGr := 1 // default size
	if len(os.Args) >= 3 {
		buffSize, err := strconv.Atoi(os.Args[2])
		if err == nil {
			numOfGr = buffSize
		}
	}
	results := make(chan HostStatus, numOfGr)

	// producer goroutine
	var wg sync.WaitGroup

	hostCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("producer starts")
		for scanner.Scan() {
			host := scanner.Text()
			hostCount++
			go ping(results, host)
		}
		fmt.Println("producer ends")
	}()

	// consumer goroutine
	upCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("consumer starts")
		processed := 0
		for {
			res, valid := <-results
			if !valid {
				fmt.Println("results channel is closed")
				break
			}
			if (res.isUp) {
				upCount++
				fmt.Println(res.hostName, "is up")
			} else {
				fmt.Println(res.hostName, "is down")
			}
			processed++
			if processed == hostCount {
				close(results)
			}

		}
		fmt.Println("consumer closed results channel")
		fmt.Println("consumer ends")
	}()

	wg.Wait()

	// get the ending time and calculate the total duration
	endTime := time.Now().UTC()
	fmt.Println(endTime.Sub(startTime))
	fmt.Println(upCount, "hosts are up")
	fmt.Println(hostCount - upCount, "hosts are down")
}
