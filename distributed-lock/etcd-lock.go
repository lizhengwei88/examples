package main

import (
	"log"
	"zieckey/etcdsync"
)

func main() {
	m, err := etcdsync.New("/lock", 10, []string{"http://127.0.0.1:2379"})
	if err != nil || m == nil {
		log.Printf("etcdsync.New failed")
		return
	}
	err = m.Lock()
	if err != nil {
		log.Printf("etcdsync.lock failed")
		return
	}
	//do something
	err = m.Unlock()
	if err != nil {
		log.Printf("etcdsync.unlock faild")
		return
	}
	log.Printf("etcdsync.unlock ok")
}
