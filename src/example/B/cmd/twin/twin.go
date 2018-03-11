package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

var (
	grace = flag.String("sock", "/tmp/grace.sock", "path to a graceful socket")
	addr  = flag.String("addr", ":8811", "addr to bind to")
)

func main() {
	fmt.Print("attempt to receive listener...")
	ln, err := receive()
	if err != nil {
		fmt.Printf(" error: %v\n", err)
		fmt.Printf("attemtp to listen %q...", *addr)
		ln, err = net.Listen("tcp", *addr)
		if err != nil {
			panic(err.Error())
		}
	}
	fmt.Println(" OK")

	term := make(chan struct{})

	if _, err := os.Stat(*grace); err == nil {
		fmt.Printf("unlinking stale socket %q\n", *grace)
		os.Remove(*grace)
	}
	go sender(ln, term)

	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("processing request...")
			time.Sleep(10 * time.Second)
			fmt.Fprintf(w, "OK %d", os.Getpid())
		}),
	}

	go server.Serve(ln)

	<-term
	fmt.Println("shutting down...")
	server.Shutdown(context.Background())
}

func sender(ln net.Listener, done chan struct{}) {
	sockLn, err := net.Listen("unix", *grace)
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		sockLn.Close()
		close(done)
	}()

	worker, err := sockLn.Accept()
	if err != nil {
		panic(err.Error())
	}

	file, err := ln.(*net.TCPListener).File()
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	err = send(worker.(*net.UnixConn), int(file.Fd()))
	if err != nil {
		panic(err.Error())
	}
}

func send(conn *net.UnixConn, fd int) error {
	msg := []byte("my-listener")
	oob := syscall.UnixRights(fd)
	_, _, err := conn.WriteMsgUnix(msg, oob, nil)
	return err
}

func receive() (net.Listener, error) {
	conn, err := net.Dial("unix", *grace)
	if err != nil {
		return nil, err
	}

	var (
		msg = make([]byte, 4096)
		oob = make([]byte, 4096)
	)
	msgn, oobn, _, _, err := conn.(*net.UnixConn).ReadMsgUnix(msg, oob)
	if err != nil {
		return nil, err
	}
	cmsg, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, err
	}
	if len(cmsg) == 0 {
		return nil, fmt.Errorf("empty control message")
	}
	fds, err := syscall.ParseUnixRights(&cmsg[0])
	if err != nil {
		return nil, err
	}
	if len(fds) == 0 {
		return nil, fmt.Errorf("empty descriptors list")
	}

	file := os.NewFile(uintptr(fds[0]), string(msg[:msgn]))
	defer file.Close()

	return net.FileListener(file)
}
