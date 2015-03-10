package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/antonholmquist/jason"
)

type Node struct {
	Name        string
	Description string
	Session     string
	State       string
}
type Nodes map[string]Node

func FindNodes(host string, search string, user string, pass string) (Nodes, error) {
	ip := net.ParseIP(search)
	var json *jason.Object
	nodes := Nodes{}
	json, err := GetReq(host, "/mgmt/tm/ltm/node", user, pass)
	if err != nil {
		return nil, err
	}
	items, _ := json.GetObjectArray("items")
	for _, value := range items {
		if ip != nil {
			addr, _ := value.GetString("address")
			if addr != ip.String() {
				continue
			}
		} else if search != "" {
			name, _ := value.GetString("name")
			if search != name {
				continue
			}

		}
		var node Node
		address, _ := value.GetString("address")
		name, _ := value.GetString("name")
		description, _ := value.GetString("description")
		state, _ := value.GetString("state")
		session, _ := value.GetString("session")
		node.Name = name
		node.Description = description
		node.State = state
		node.Session = session
		nodes[address] = node
	}
	return nodes, nil
}

func GetNodes(w rest.ResponseWriter, r *rest.Request) {
	env := r.PathParam("env")
	search := r.PathParam("search")
	host := ""
	if bigip, ok := cfg.F5[env]; ok {
		host = GetActive(bigip.Ltm, bigip.User, bigip.Pass)
	} else {
		rest.Error(w, "Env not found", 404)
		return
	}
	if host == "" {
		rest.Error(w, "No active ltm", http.StatusInternalServerError)
		return
	}
	nodes, err := FindNodes(host, search, cfg.F5[env].User, cfg.F5[env].Pass)
	if err != nil {
		log.Panic(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-F5", host)
	if len(nodes) > 0 {
		w.WriteJson(nodes)
		return
	} else {
		rest.Error(w, "Node not found", 404)
		return
	}
}

func PutNodes(w rest.ResponseWriter, r *rest.Request) {
	env := r.PathParam("env")
	search := r.PathParam("search")
	var payload map[string]string
	err := r.DecodeJsonPayload(&payload)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if payload["State"] != "enabled" && payload["State"] != "disabled" && payload["State"] != "forced-offline" {
		rest.Error(w, "Please give a valid state", 400)
		return
	}
	host := ""
	if bigip, ok := cfg.F5[env]; ok {
		host = GetActive(bigip.Ltm, bigip.User, bigip.Pass)
	} else {
		rest.Error(w, "Env not found", 404)
		return
	}
	if host == "" {
		rest.Error(w, "No active ltm", http.StatusInternalServerError)
		return
	}
	nodes, err := FindNodes(host, search, cfg.F5[env].User, cfg.F5[env].Pass)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-F5", host)
	if len(nodes) == 1 {
		var node Node
		for _, n := range nodes {
			node = n
		}
		if payload["State"] == "enabled" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-up", "session": "user-enabled"}`), cfg.F5[env].User, cfg.F5[env].Pass)
			nodes, err := FindNodes(host, search, cfg.F5[env].User, cfg.F5[env].Pass)
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteJson(nodes)
			return
		}
		if payload["State"] == "disabled" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-up", "session": "user-disabled"}`), cfg.F5[env].User, cfg.F5[env].Pass)
			nodes, err := FindNodes(host, search, cfg.F5[env].User, cfg.F5[env].Pass)
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteJson(nodes)
			return
		}
		if payload["State"] == "forced-offline" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-down", "session": "user-disabled"}`), cfg.F5[env].User, cfg.F5[env].Pass)
			nodes, err := FindNodes(host, search, cfg.F5[env].User, cfg.F5[env].Pass)
			if err != nil {
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteJson(nodes)
			return
		}
	} else {
		rest.Error(w, "Node not found", 404)
		return
	}
}
