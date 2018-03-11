package main

import (
	"flag"
	"time"
)

func (s *Server) ExportFlags(f *flag.FlagSet) {
	f.StringVar(&s.Addr,
		"addr", "127.0.0.1:3000",
		"addr to listen to",
	)
	f.DurationVar(&s.Delay,
		"delay", time.Second*5,
		"delay of the response",
	)
}

func reverse(p []byte) {
	var (
		i = 0
		j = len(p) - 1
	)
	for i < j {
		p[i], p[j] = p[j], p[i]
		i++
		j--
	}
}
