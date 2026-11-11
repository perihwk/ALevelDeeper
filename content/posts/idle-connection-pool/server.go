package main

import (
	"fmt"
	"net/http"
	"syscall"
	"time"
	"uuid"
)

func tv(t syscall.Timeval) time.Duration { return time.Duration(t.Nano()) }

func sampleCPU(interval time.Duration) {
	var prev syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &prev)
	last := time.Now()
	for range time.Tick(interval) {
		var cur syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &cur)
		now := time.Now()
		el := now.Sub(last).Seconds()
		user := (tv(cur.Utime) - tv(prev.Utime)).Seconds() / el * 100
		sys := (tv(cur.Stime) - tv(prev.Stime)).Seconds() / el * 100
		fmt.Printf("%d,%.1f,%.1f\n", now.Unix(), user, sys) // time,user%,sys%
		prev, last = cur, now
	}
}

func uuid_gen(w http.ResponseWriter, req *http.Request) {
	time.Sleep(1 * time.Second) // simulate a slow response
	fmt.Fprintf(w, "%s\n", uuid.New().String())
}

func main() {
	//go sampleCPU(time.Millisecond)
	http.HandleFunc("/uuid", uuid_gen)
	http.ListenAndServe(":8090", nil)
}
