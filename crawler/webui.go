package crawler

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"log"
	"net/http"
	"strconv"
)

func (c *Crawler) StartQueryServer() {
	http.HandleFunc("/storage", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			fmt.Fprintf(w, "Illegal Id  %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		thread, err := c.Storage.Get(id)
		if err != nil {
			fmt.Fprint(w, "Cannot find thread: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		marshaler := jsonpb.Marshaler{}
		marshaler.Marshal(w, thread)
		w.Header().Set("Content-Type", "application/json")
	})
	go http.ListenAndServe(":8080", nil)
	log.Println("Start query server at :8080")
}
