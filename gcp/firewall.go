package gcp

import (
	"io/ioutil"
	"net/http"
	"time"
)

func getAddress() string {
	var c = &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := c.Get("https://getaddress.vthor.workers.dev")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func GetFWRUles() {
	computeService.Firewalls.Get(computeService.project, "client")
}
