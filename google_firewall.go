package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func getAddress() string {
	var c = &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := c.Get("https://atto.run/ip")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func (c *ComputeClient) GetFirewallRules() error {
	a, err := c.Firewalls.Get(c.Project, "morning-mgmt-to-backend").Do()
	if err != nil {
		return err
	}
	b := getAddress()
	for _, r := range a.SourceRanges {
		if strings.HasPrefix(r, b) {
			fmt.Println(r)
		}
	}
	return nil
}
