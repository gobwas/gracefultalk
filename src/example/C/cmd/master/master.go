package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Master struct {
	Addr    string
	Command string
}

func main() {
	m := new(Master)
	m.ExportFlags(flag.CommandLine)
	flag.Parse()

	cmdArgs := strings.Split(m.Command, " ")
	if len(cmdArgs) < 1 {
		log.Fatalf("insufficient command arguments")
	}
	if len(cmdArgs[0]) == 0 {
		log.Fatalf("empty command binary")
	}

	ln, err := net.Listen("tcp", m.Addr)
	if err != nil {
		log.Fatal(err)
	}
	lnFile, err := ln.(*net.TCPListener).File()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("master #%d is exporting %s", os.Getpid(), ln.Addr())

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGUSR1)

	var (
		mu      sync.Mutex
		running *exec.Cmd
	)
	for sig := range ch {
		log.Printf("signal received: %v", sig)
		switch sig {
		case syscall.SIGTERM:
			mu.Lock() // Do not unlock the mutex.
			cmd := running
			if cmd != nil {
				terminate(cmd)
			}
			os.Exit(0)

		case syscall.SIGUSR1:
			mu.Lock()
			log.Printf("starging with args %v", cmdArgs)
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			cmd.ExtraFiles = []*os.File{lnFile}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				log.Fatalf("start command error: %v", err)
			}
			log.Printf("started command with pid %d", cmd.Process.Pid)
			if running != nil {
				terminate(running)
			}
			running = cmd
			mu.Unlock()
		}
	}
}

func terminate(cmd *exec.Cmd) {
	log.Printf(
		"sending SIGTERM to #%d instance",
		cmd.Process.Pid,
	)
	cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Second):
		kill(cmd)
	case err := <-done:
		log.Printf(
			"instance #%d exited (err is %v)",
			cmd.Process.Pid, err,
		)
	}
}

func kill(cmd *exec.Cmd) {
	log.Printf(
		"sending SIGKILL to #%d instance",
		cmd.Process.Pid,
	)
	cmd.Process.Signal(syscall.SIGKILL)
}
