package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Server contains options for handling requests.
type Server struct {
	Addr  string
	Delay time.Duration
}

func (s *Server) ListenAndServe() error {
	echo := func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		r.Body.Close()
		reverse(body)
		time.Sleep(s.Delay)
		w.Write(body)
	}
	srv := http.Server{
		Handler: http.HandlerFunc(echo),
		Addr:    s.Addr,
	}
	var (
		exit = make(chan struct{})
		term = make(chan os.Signal, 1)
	)
	signal.Notify(term, syscall.SIGINT)
	go func() {
		var once sync.Once
		for sig := range term {
			log.Printf("signal received: %v", sig)
			once.Do(func() { // Ignore double SIGTERM.
				go func() {
					srv.Shutdown(context.TODO())
					close(exit)
				}()
			})
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	<-exit
	return nil
}

func main() {
	s := new(Server)
	s.ExportFlags(flag.CommandLine)
	flag.Parse()
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	log.Println("shutdown OK")
}
