package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/antonholmquist/jason"
)

func GetReq(host string, uri string, user string, pass string) (*jason.Object, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, err := http.NewRequest("GET", "https://"+host+uri, nil)
	if err != nil {
		log.Println(err)
	}
	req.SetBasicAuth(user, pass)
	client := &http.Client{Transport: tr}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, errors.New("Got '" + rsp.Status + "' from ltm.")
	}
	body, err := ioutil.ReadAll(rsp.Body)
	json, err := jason.NewObjectFromBytes(body)
	if err != nil {
		log.Println("Problem parsing json resp: ", err)
	}
	return json, nil
}

func PutReq(host string, uri string, payload []byte, user string, pass string) (*jason.Object, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, err := http.NewRequest("PUT", "https://"+host+uri, bytes.NewBuffer(payload))
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(pass, pass)
	client := &http.Client{Transport: tr}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return nil, errors.New("Got '" + rsp.Status + "' from ltm.")
	}
	body, err := ioutil.ReadAll(rsp.Body)
	json, err := jason.NewObjectFromBytes(body)
	if err != nil {
		log.Println("Problem parsing json resp: ", err)
	}
	return json, nil
}

func GetActive(lbs []string, user string, pass string) string {
	active := ""
	var wg sync.WaitGroup
	for _, lb := range lbs {
		// Increment the WaitGroup counter.
		wg.Add(1)
		go func(lb string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			json, err := GetReq(lb, "/mgmt/tm/cm/failover-status", user, pass)
			if err != nil {
				log.Println(err)
				return
			}
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
