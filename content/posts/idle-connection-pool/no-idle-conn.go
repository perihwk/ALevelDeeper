package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"uuid"
)

func get_uuid() {
	resp, err := http.Get("http://127.0.0.1:8090/uuid")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	fmt.Printf("%s", string(body))
}

func server() {
	http.HandleFunc("/uuid", uuid_gen)
	http.ListenAndServe(":8090", nil)
}

func uuid_gen(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s\n", uuid.New().String())
}

func main() {
	go server()
	time.Sleep(1 * time.Second) // give server a chance to come alive
	get_uuid()
	get_uuid()
}
