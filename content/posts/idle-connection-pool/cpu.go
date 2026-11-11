package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func tv(t syscall.Timeval) time.Duration { return time.Duration(t.Nano()) }

// sampleCPU prints "elapsed_ms,user_pct,sys_pct" every interval until stop is closed.
// Percentages are of one core, so they can exceed 100 on multiple cores.
func sampleCPU(interval time.Duration, start time.Time, stop <-chan struct{}) {
	var prev syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &prev)
	last := time.Now()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
		}
		var cur syscall.Rusage
		syscall.Getrusage(syscall.RUSAGE_SELF, &cur)
		now := time.Now()
		el := now.Sub(last).Seconds()
		user := (tv(cur.Utime) - tv(prev.Utime)).Seconds() / el * 100
		sys := (tv(cur.Stime) - tv(prev.Stime)).Seconds() / el * 100
		fmt.Printf("%d,%.1f,%.1f\n", now.Sub(start).Milliseconds(), user, sys)
		prev, last = cur, now
	}
}

func newClient(mode string, conc int) *http.Client {
	switch mode {
	case "default":
		// stock transport: MaxIdleConnsPerHost = 2
		return http.DefaultClient
	case "nopool":
		// every connection is closed by the client as soon as the request finishes
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConnsPerHost = -1
		return &http.Client{Transport: t}
	case "sized":
		// pool large enough to hold every connection from one round
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = conc
		t.MaxIdleConnsPerHost = conc
		return &http.Client{Transport: t}
	}
	fmt.Fprintf(os.Stderr, "unknown mode %q (want default, nopool or sized)\n", mode)
	os.Exit(2)
	return nil
}

func main() {
	mode := flag.String("mode", "default", "default | nopool | sized")
	conc := flag.Int("conc", 200, "concurrent requests per round")
	dur := flag.Duration("duration", 20*time.Second, "how long to keep running rounds")
	interval := flag.Duration("interval", 200*time.Millisecond, "CPU sample interval")
	url := flag.String("url", "http://127.0.0.1:8090/uuid", "target URL")
	flag.Parse()

	client := newClient(*mode, *conc)
	var errs atomic.Int64

	get := func() {
		resp, err := client.Get(*url)
		if err != nil {
			errs.Add(1)
			return
		}
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
	}

	start := time.Now()
	stop := make(chan struct{})
	done := make(chan struct{})
	fmt.Println("elapsed_ms,user_pct,sys_pct")
	go func() {
		sampleCPU(*interval, start, stop)
		close(done)
	}()

	rounds := 0
	for time.Since(start) < *dur {
		var wg sync.WaitGroup
		for range *conc {
			wg.Go(get)
		}
		wg.Wait()
		rounds++
	}

	close(stop)
	<-done
	fmt.Fprintf(os.Stderr, "mode=%s rounds=%d requests=%d errors=%d elapsed=%s\n",
		*mode, rounds, rounds**conc, errs.Load(), time.Since(start).Round(time.Millisecond))
}
