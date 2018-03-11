package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client contains options for dealing with a server.
type Client struct {
	Addr       string
	RetryDelay time.Duration
	RetryLimit int
}

func (c *Client) Call(ctx context.Context, data string) (string, error) {
	req, err := http.NewRequest(
		"GET", fmt.Sprintf("http://%s", c.Addr),
		strings.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	var resp *http.Response
	for i := c.RetryLimit; i >= 0; i-- {
		resp, err = http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			if isConnectionRefused(err) {
				time.Sleep(c.RetryDelay)
				continue
			}
		}
		break
	}
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.Close {
		log.Println("server requested to close the connection")
	}

	return string(body), nil
}

func main() {
	c := new(Client)
	c.ExportFlags(flag.CommandLine)
	flag.Parse()

	log.SetFlags(0)
	fmt.Print("> ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		resp, err := c.Call(context.TODO(), s.Text())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("< %s\n> ", resp)
	}
}
