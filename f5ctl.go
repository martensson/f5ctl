/* f5ctl - Code by Benjamin Mårtensson <benjamin@martensson.io> */
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"gopkg.in/yaml.v1"
)

type Config struct {
	Apiuser string
	Apipass string
	Lbuser  string
	Lbpass  string
	Felles  []string
	Dmz     []string
}

var cfg Config

func main() {
	port := flag.String("p", "5000", "Listen on this port. (default 5000)")
	config := flag.String("f", "config.yml", "Path to config. (default config.yml)")
	flag.Parse()
	file, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal("Problem parsing config: ", err)
	}
	api := rest.NewApi()
	//api.Use(rest.DefaultProdStack...)
	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)
	api.Use(&rest.AccessLogApacheMiddleware{Format: rest.CombinedLogFormat})
	api.Use(&rest.TimerMiddleware{})
	api.Use(&rest.RecorderMiddleware{})
	api.Use(&rest.PoweredByMiddleware{XPoweredBy: "f5ctl"})
	api.Use(&rest.RecoverMiddleware{})
	api.Use(&rest.GzipMiddleware{})
	api.Use(&rest.JsonIndentMiddleware{})
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "f5ctl",
		Authenticator: func(userId string, password string) bool {
			if userId == cfg.Apiuser && password == cfg.Apipass {
				return true
			}
			return false
		},
	})
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(statusMw.GetStatus())
			},
		},
		&rest.Route{"GET", "/v1/nodes/:env/#search", GetNodes},
		&rest.Route{"GET", "/v1/nodes/:env/", GetNodes},
		&rest.Route{"PUT", "/v1/nodes/:env/#search", PutNodes},
	)
	api.SetApp(router)
	log.Println("Starting f5ctl on :" + *port)
	log.Fatal(http.ListenAndServe(":"+*port, api.MakeHandler()))
}
