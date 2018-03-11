package main

import "flag"

func (m *Master) ExportFlags(f *flag.FlagSet) {
	f.StringVar(&m.Addr,
		"addr", "127.0.0.1:3000",
		"addr to listen to",
	)
	f.StringVar(&m.Command,
		"c", "",
		"command to run for worker",
	)
}
