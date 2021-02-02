package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"sync"
)

const saveQueueLength = 1000

type Store interface {
	GetUrl(key, url *string) error
	PutUrl(url, key *string) error
}

type ProxyStore struct {
	urlstore *Urlstore
	client   *rpc.Client
}
type Urlstore struct {
	url      map[string]string
	m        sync.RWMutex
	f        *os.File
	saveChan chan record
	methods  string
}
type record struct {
	Key, Url string
}

func NewUrlstore(fileName string) *Urlstore {
	store := &Urlstore{url: make(map[string]string), saveChan: make(chan record, saveQueueLength)}
	if fileName != "" {
		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal("Error opening URLStore: ", err)
		}
		store.f = f

		if err := store.Load(fileName); err != nil {
			log.Println("Error loading URLStore:", err)
		}
		go store.saveChanToFile(fileName)
	}
	return store
}

func (store *Urlstore) saveChanToFile(filename string) error {
	defer store.f.Close()
	e := json.NewEncoder(store.f)
	for {
		fmt.Println("ch for...")
		recordC := <-store.saveChan
		if err := e.Encode(recordC); err != nil {
			log.Println("Error saving to URLStore: ", err)
		}
	}
}

func (store *Urlstore) save(key, url string) error {
	e := json.NewEncoder(store.f)
	err := e.Encode(record{key, url})
	return err
}

func (store *Urlstore) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening URLStore:", err)
		return err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			fmt.Println(r, err)
			store.Set(&r.Key, &r.Url)
		}
	}
	if err == io.EOF {
		return nil
	}
	fmt.Println("err:", err)
	return err
}

func (store *Urlstore) Count() int {
	return len(store.url)
}

func (store *Urlstore) Set(key, url *string) error {
	store.m.Lock()
	defer store.m.Unlock()
	if _, present := store.url[*key]; present {
		return errors.New("key already exists")
	}
	store.url[*key] = *url
	return nil
}

func (store *Urlstore) GetUrl(key, url *string) error {
	store.m.RLock()
	defer store.m.RUnlock()
	if u, ok := store.url[*key]; ok {
		*url = u
		return nil
	}
	return errors.New("key not found")
}

func (store *Urlstore) PutUrl(url, key *string) error {
	for {
		*key = genKey(store.Count())
		if err := store.Set(key, url); err == nil {
			store.saveChan <- record{*key, *url}
			break
		}
	}
	return nil
}

func Newporxy(address string) *ProxyStore {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Println("Error construction pro")
	}
	return &ProxyStore{client: client, urlstore: NewUrlstore("")}
}

func (store *ProxyStore) GetUrl(key, url *string) error {
	err := store.urlstore.GetUrl(key, url)
	if err == nil {
		return nil
	}

	err = store.client.Call("Store.GetUrl", key, url)
	if err != nil {
		return err
	}
	store.urlstore.Set(key, url)
	return nil
}

func (store *ProxyStore) PutUrl(url, key *string) error {
	err := store.client.Call("Store.PutUrl", url, key)
	if err != nil {
		return err
	}
	store.urlstore.Set(key, url)
	return nil
}
