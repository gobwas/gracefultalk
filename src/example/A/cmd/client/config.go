package main

import (
	"flag"
	"strings"
	"time"
)

func (c *Client) ExportFlags(f *flag.FlagSet) {
	f.StringVar(&c.Addr,
		"addr", "127.0.0.1:3000",
		"addr to dial to",
	)
	f.DurationVar(&c.RetryDelay,
		"retry_delay", time.Second,
		"timeout for reestablish the connection",
	)
	f.IntVar(&c.RetryLimit,
		"retry_limit", 5,
		"number of attempts to reestablish timeout",
	)
}

func isConnectionRefused(err error) bool {
	return strings.Contains(err.Error(), "connection refused")
}
