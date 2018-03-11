package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var delay = flag.Duration("delay", time.Second*5, "delay of the response")

func main() {
	flag.Parse()
	log.SetPrefix(fmt.Sprintf("[worker %d] ", os.Getpid()))

	file := os.NewFile(3, "listener")
	ln, err := net.FileListener(file)
	if err != nil {
		log.Fatal(err)
	}
	s := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(*delay)
			io.Copy(w, r.Body)
			r.Body.Close()
		}),
	}
	var (
		exit = make(chan struct{})
		term = make(chan os.Signal, 1)
	)
	signal.Notify(term, syscall.SIGTERM)
	go func() {
		var once sync.Once
		for sig := range term {
			log.Printf("signal received: %v", sig)
			once.Do(func() { // Ignore double SIGTERM.
				go func() {
					s.Shutdown(context.TODO())
					close(exit)
				}()
			})
		}
	}()
	log.Println("serving...")
	if err := s.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-exit
}
