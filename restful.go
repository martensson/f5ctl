package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/antonholmquist/jason"
)

func GetReq(host string, uri string) *jason.Object {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, err := http.NewRequest("GET", "https://"+host+uri, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(cfg.Lbuser, cfg.Lbpass)
	client := &http.Client{Transport: tr}
	rsp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	json, err := jason.NewObjectFromBytes(body)
	if err != nil {
		log.Fatal("Problem parsing json resp: ", err)
	}
	return json
}

func PutReq(host string, uri string, payload []byte) *jason.Object {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, err := http.NewRequest("PUT", "https://"+host+uri, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(cfg.Lbuser, cfg.Lbpass)
	client := &http.Client{Transport: tr}
	rsp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	json, err := jason.NewObjectFromBytes(body)
	if err != nil {
		log.Fatal("Problem parsing json resp: ", err)
	}
	return json
}

func GetActive(lbs []string) string {
	active := ""
	var wg sync.WaitGroup
	for _, lb := range lbs {
		// Increment the WaitGroup counter.
		wg.Add(1)
		go func(lb string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			json := GetReq(lb, "/mgmt/tm/cm/failover-status")
			status, _ := json.GetObject("entries")
			for _, value := range status.Map() {
				entries, _ := value.Object()
				status, _ := entries.GetObject("nestedStats", "entries", "status")
				description, _ := status.GetString("description")
				if description == "ACTIVE" {
					active = lb
					break
				}
			}
		}(lb)
	}
	// Wait for all requests to complete.
	wg.Wait()
	return active
}
