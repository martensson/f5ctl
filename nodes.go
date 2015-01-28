package main

import (
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

func FindNodes(host string, search string) Nodes {
	ip := net.ParseIP(search)
	var json *jason.Object
	nodes := Nodes{}
	json = GetReq(host, "/mgmt/tm/ltm/node")
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
	return nodes
}

func GetNodes(w rest.ResponseWriter, r *rest.Request) {
	env := r.PathParam("env")
	search := r.PathParam("search")
	host := ""
	switch env {
	case "felles":
		host = GetActive(cfg.Felles)
	case "dmz":
		host = GetActive(cfg.Dmz)
	default:
		rest.Error(w, "Env not found", 404)
		return
	}
	nodes := FindNodes(host, search)
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
	switch env {
	case "felles":
		host = GetActive(cfg.Felles)
	case "dmz":
		host = GetActive(cfg.Dmz)
	default:
		rest.Error(w, "Env not found", 404)
		return
	}
	nodes := FindNodes(host, search)
	w.Header().Set("X-F5", host)
	if len(nodes) == 1 {
		var node Node
		for _, n := range nodes {
			node = n
		}
		if payload["State"] == "enabled" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-up", "session": "user-enabled"}`))
			nodes := FindNodes(host, search)
			w.WriteJson(nodes)
			return
		}
		if payload["State"] == "disabled" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-up", "session": "user-disabled"}`))
			nodes := FindNodes(host, search)
			w.WriteJson(nodes)
			return
		}
		if payload["State"] == "forced-offline" {
			PutReq(host, "/mgmt/tm/ltm/node/"+node.Name, []byte(`{"state": "user-down", "session": "user-disabled"}`))
			nodes := FindNodes(host, search)
			w.WriteJson(nodes)
			return
		}
	} else {
		rest.Error(w, "Node not found", 404)
		return
	}
}
