package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
	"sync"
	"time"
	"uuid"
)

func get_uuid() {
	resp, err := http.Get("http://127.0.0.1:8090/uuid")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%s", string(body))
}

func server() {
	http.HandleFunc("/uuid", uuid_gen)
	http.ListenAndServe(":8090", nil)
}

func uuid_gen(w http.ResponseWriter, req *http.Request) {
	time.Sleep(1 * time.Second) // simulate a slow response
	fmt.Fprintf(w, "%s\n", uuid.New().String())
}

func main() {
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()
	go server()
	time.Sleep(1 * time.Second) // give server a chance to come alive
	for {
		var wg sync.WaitGroup
		for range 2 {
			wg.Go(func() {
				get_uuid()
			})
		}
		wg.Wait()
	}
}
