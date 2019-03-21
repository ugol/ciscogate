// Copyright © 2018
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	address         = "localhost:8080"
	apicURL         = "apic1.rmlab.local"
	apicUsername    = "admin"
	apicPassword    = "C1sco123"
	openshiftTenant = "openshift39"
	epgToBeCreated  = "prova18e26"

	writeTimeout    = time.Second * 15
	readTimeout     = time.Second * 15
	idleTimeout     = time.Second * 60
	gracefulTimeout = time.Second * 15
)

var (
	totalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ciscogate_total_votes",
		Help: "The total number of processed requests",
	})
)

func init() {

	bindEnvToStringVar(&address, "CISCO_ADDRESS")
	bindEnvToStringVar(&apicURL, "CISCO_APICURL")
	bindEnvToStringVar(&apicUsername, "CISCO_APICUSERNAME")
	bindEnvToStringVar(&apicPassword, "CISCO_APICPASSWORD")
	bindEnvToStringVar(&openshiftTenant, "CISCO_OPENSHIFTTENANT")
	bindEnvToStringVar(&epgToBeCreated, "CISCO_EPGTOBECREATED")
	bindEnvToDurationVar(&writeTimeout, "CISCO_WRITETIMEOUT")
	bindEnvToDurationVar(&readTimeout, "CISCO_READTIMEOUT")
	bindEnvToDurationVar(&idleTimeout, "CISCO_IDLETIMEOUT")
	bindEnvToDurationVar(&gracefulTimeout, "CISCO_GRACEFULTIMEOUT")

}

func bindEnvToStringVar(v *string, key string) {
	env := os.Getenv(key)
	if env != "" {
		*v = env
	}
}

func bindEnvToDurationVar(v *time.Duration, key string) {
	env := os.Getenv(key)
	if env != "" {
		d := durationFrom(env)
		*v = d
	}
}

func durationFrom(d string) time.Duration {
	duration, e := time.ParseDuration(d)
	if e != nil {
		log.Printf("Parse error: %v\n", e)
	}
	return duration
}

func PrintUsage() {

	fmt.Printf("Usage:\n"+
		"ciscogate start\n"+
		"\n "+
		"use CISCO_ADDRESS to set listen interface/webhook. Default is '%s' \n "+
		"use CISCO_APICURL to set apic url. Default is '%s' \n "+
		"use CISCO_APICUSERNAME to set user name. Default is '%s' \n "+
		"use CISCO_APICPASSWORD to set password. Default is '%s' \n "+
		"use CISCO_OPENSHIFTTENANT to set OCP tenant. Default is '%s' \n "+
		"use CISCO_EPGTOBECREATED to set epg. Default is '%s' \n "+
		"use CISCO_GRACEFULTIMEOUT to set GracefulTimeout. Default is '%s' \n "+
		"use CISCO_WRITETIMEOUT to set WriteTimeout. Default is '%s' \n "+
		"use CISCO_READTIMEOUT to set ReadTimeout. Default is '%s'\n "+
		"use CISCO_IDLETIMEOUT to set IdleTimeout. Default is '%s'\n\n",
		address, apicURL, apicUsername, apicPassword, openshiftTenant, epgToBeCreated, gracefulTimeout, writeTimeout, readTimeout, idleTimeout)
}

func printEnvironment() {
	log.Printf(""+
		"\n "+
		"Address: %s \n "+
		"ApicUrl: %s \n "+
		"ApicUsername: %s \n "+
		"ApicPassword: %s \n "+
		"OpenshiftTenant: %s \n "+
		"EpgToBeCreated: %s \n "+
		"GracefulTimeout: %s\n "+
		"WriteTimeout: %s\n "+
		"ReadTimeout: %s\n "+
		"IdleTimeout: %s\n",
		address, apicURL, apicUsername, apicPassword, openshiftTenant, epgToBeCreated, gracefulTimeout, writeTimeout, readTimeout, idleTimeout)
}

func CiscoGateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		log.Printf("Serving GET %v to %s\n", r.RequestURI, r.RemoteAddr)
		//vars := mux.Vars(r)
		totalRequests.Inc()

		tokenURL := fmt.Sprintf("https://%v/api/mo/aaaLogin.xml", apicURL)
		otherURL := fmt.Sprintf("https://%v/api/node/mo/uni/tn-%v/ap-kubernetes/epg-%v.json", apicURL, openshiftTenant, epgToBeCreated)

		xmlAuth := fmt.Sprintf("<aaaUser name='%v' pwd='%v'/>", apicUsername, apicPassword)
		xmlAuthBytes := []byte(xmlAuth)



		log.Printf("Sending a POST request to %v containing %v", tokenURL, xmlAuth)
		req, err := http.NewRequest("POST", tokenURL, bytes.NewBuffer(xmlAuthBytes))
		req.Header.Set("Content-Type", "application/xml")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		log.Println("Status -->", resp.Status)
		log.Println("Headers -->", resp.Header)
		token, _ := ioutil.ReadAll(resp.Body)

		log.Printf("Received Auth token:\n %v", string(token))

		log.Println("Loading answer.json template")
		jsonAnswerBytes, err := ioutil.ReadFile("answer.json")
		if err != nil {
			panic(err)
		}

		jsonAnswerTemplate := string(jsonAnswerBytes)
		jsonAnswer := fmt.Sprintf(jsonAnswerTemplate, openshiftTenant, epgToBeCreated, epgToBeCreated, openshiftTenant, openshiftTenant)

		log.Println(jsonAnswerTemplate)

		log.Printf("Sending a POST request with the token to %v with this json:\n %v", otherURL, jsonAnswer)

		// TODO: add token and other metadata to the cookie
		cookie := http.Cookie{}
		req, err = http.NewRequest("POST", otherURL, bytes.NewBuffer(xmlAuthBytes))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&cookie)

		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		log.Println("Status -->", resp.Status)
		log.Println("Headers -->", resp.Header)
		answer, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Server answered with:\n %v", answer)

	}
}

func StartServer() {
	log.Println("Starting Ciscogate server")
	printEnvironment()
	r := mux.NewRouter()

	r.HandleFunc("/ciscogate", CiscoGateHandler).Methods("GET")

	r.Path("/metrics").Handler(promhttp.Handler())

	http.Handle("/", r)
	srv := &http.Server{
		Addr:         address,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	log.Printf("Waiting for connections at %s\n", address)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error %v during shutdown\n", err)
		os.Exit(1)
	} else {
		log.Println("Shutting down")
		os.Exit(0)
	}

}
