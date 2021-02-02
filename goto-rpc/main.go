package main

import (
	"cs/practice/goto-rpc/common"
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
)

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	hostname   = flag.String("host", "localhost:8080", "http host name")
	dataFile   = flag.String("file", "store.json", "data store file name")
	rpcEnabled = flag.Bool("rpc", false, "start rpc server")
	masterAddr = flag.String("master", "", "rpc master address")
)

var store common.Store

func main() {
	flag.Parse()
	if *masterAddr != "" {
		store = common.Newporxy(*masterAddr)
	} else {
		store = common.NewUrlstore(*dataFile)
	}
	if *rpcEnabled {
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}
	http.HandleFunc("/add", AddUrl)
	http.HandleFunc("/", Redirect)
	http.ListenAndServe(*listenAddr, nil)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	var url string
	err := store.GetUrl(&key, &url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func AddUrl(w http.ResponseWriter, r *http.Request) {
	urlname := r.FormValue("url")
	if urlname == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	var key string
	err := store.PutUrl(&urlname, &key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}
