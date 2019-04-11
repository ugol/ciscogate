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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
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

type Imdata struct {
	XMLName    xml.Name `xml:"imdata"`
	Text       string   `xml:",chardata"`
	TotalCount string   `xml:"totalCount,attr"`
	AaaLogin   struct {
		Text                   string `xml:",chardata"`
		Token                  string `xml:"token,attr"`
		SiteFingerprint        string `xml:"siteFingerprint,attr"`
		RefreshTimeoutSeconds  string `xml:"refreshTimeoutSeconds,attr"`
		MaximumLifetimeSeconds string `xml:"maximumLifetimeSeconds,attr"`
		GuiIdleTimeoutSeconds  string `xml:"guiIdleTimeoutSeconds,attr"`
		RestTimeoutSeconds     string `xml:"restTimeoutSeconds,attr"`
		CreationTime           string `xml:"creationTime,attr"`
		FirstLoginTime         string `xml:"firstLoginTime,attr"`
		UserName               string `xml:"userName,attr"`
		RemoteUser             string `xml:"remoteUser,attr"`
		UnixUserId             string `xml:"unixUserId,attr"`
		SessionId              string `xml:"sessionId,attr"`
		LastName               string `xml:"lastName,attr"`
		FirstName              string `xml:"firstName,attr"`
		ChangePassword         string `xml:"changePassword,attr"`
		Version                string `xml:"version,attr"`
		BuildTime              string `xml:"buildTime,attr"`
		Node                   string `xml:"node,attr"`
		AaaUserDomain          struct {
			Text          string `xml:",chardata"`
			Name          string `xml:"name,attr"`
			RolesR        string `xml:"rolesR,attr"`
			RolesW        string `xml:"rolesW,attr"`
			AaaReadRoles  string `xml:"aaaReadRoles"`
			AaaWriteRoles struct {
				Text string `xml:",chardata"`
				Role struct {
					Text string `xml:",chardata"`
					Name string `xml:"name,attr"`
				} `xml:"role"`
			} `xml:"aaaWriteRoles"`
		} `xml:"aaaUserDomain"`
		DnDomainMapEntry []struct {
			Text            string `xml:",chardata"`
			Dn              string `xml:"dn,attr"`
			ReadPrivileges  string `xml:"readPrivileges,attr"`
			WritePrivileges string `xml:"writePrivileges,attr"`
		} `xml:"DnDomainMapEntry"`
	} `xml:"aaaLogin"`
}

type KubeResource struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Request    struct {
		UID  string `json:"uid"`
		Kind struct {
			Group   string `json:"group"`
			Version string `json:"version"`
			Kind    string `json:"kind"`
		} `json:"kind"`
		Resource struct {
			Group    string `json:"group"`
			Version  string `json:"version"`
			Resource string `json:"resource"`
		} `json:"resource"`
		Operation string `json:"operation"`
		UserInfo  struct {
			Username string   `json:"username"`
			Groups   []string `json:"groups"`
		} `json:"userInfo"`
		Object struct {
			Metadata struct {
				Name              string    `json:"name"`
				UID               string    `json:"uid"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
				Annotations       struct {
					OpenshiftIoDescription string `json:"openshift.io/description"`
					OpenshiftIoDisplayName string `json:"openshift.io/display-name"`
					OpenshiftIoRequester   string `json:"openshift.io/requester"`
				} `json:"annotations"`
			} `json:"metadata"`
			Spec struct {
				Finalizers []string `json:"finalizers"`
			} `json:"spec"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"object"`
		OldObject interface{} `json:"oldObject"`
	} `json:"request"`
}

var (
	address         = "0.0.0.0:8080"
	apicURL         = "apic1.rmlab.local"
	apicUsername    = "admin"
	apicPassword    = "C1sco123"
	openshiftTenant = "openshift39"
	//epgToBeCreated  = "prova18e26"

	writeTimeout    = time.Second * 15
	readTimeout     = time.Second * 15
	idleTimeout     = time.Second * 60
	gracefulTimeout = time.Second * 15
	ciscoStub       = "false"
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
	//bindEnvToStringVar(&epgToBeCreated, "CISCO_EPGTOBECREATED")
	bindEnvToDurationVar(&writeTimeout, "CISCO_WRITETIMEOUT")
	bindEnvToDurationVar(&readTimeout, "CISCO_READTIMEOUT")
	bindEnvToDurationVar(&idleTimeout, "CISCO_IDLETIMEOUT")
	bindEnvToDurationVar(&gracefulTimeout, "CISCO_GRACEFULTIMEOUT")
        bindEnvToStringVar(&ciscoStub, "CISCO_STUB")

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
		"use CISCO_GRACEFULTIMEOUT to set GracefulTimeout. Default is '%s' \n "+
		"use CISCO_WRITETIMEOUT to set WriteTimeout. Default is '%s' \n "+
		"use CISCO_READTIMEOUT to set ReadTimeout. Default is '%s'\n "+
		"use CISCO_IDLETIMEOUT to set IdleTimeout. Default is '%s'\n\n",
		address, apicURL, apicUsername, apicPassword, openshiftTenant, gracefulTimeout, writeTimeout, readTimeout, idleTimeout)
}

func printEnvironment() {
	log.Printf(""+
		"\n "+
		"Address: %s \n "+
		"ApicUrl: %s \n "+
		"ApicUsername: %s \n "+
		"ApicPassword: %s \n "+
		"OpenshiftTenant: %s \n "+
		"GracefulTimeout: %s\n "+
		"WriteTimeout: %s\n "+
		"ReadTimeout: %s\n "+
		"IdleTimeout: %s\n "+
		"CiscoStub: %s\n",
		address, apicURL, apicUsername, apicPassword, openshiftTenant, gracefulTimeout, writeTimeout, readTimeout, idleTimeout, ciscoStub)
}

func CiscoGateHandler(w http.ResponseWriter, r *http.Request) {

	totalRequests.Inc()

	if r.Method == http.MethodGet {
		log.Printf("Received GET %v from %s\n", r.RequestURI, r.RemoteAddr)
		vars := mux.Vars(r)
		epgToBeCreated := vars["epg"]
		doThat(epgToBeCreated)
	} else if r.Method == http.MethodPost {
		log.Printf("Received POST %v from %s\n", r.RequestURI, r.RemoteAddr)

		decoder := json.NewDecoder(r.Body)
		var k KubeResource
		err := decoder.Decode(&k)
		if err != nil {
			panic(err)
		}
		fmt.Println(k)
		epgToBeCreated := k.Request.Object.Metadata.Name
                if ciscoStub != "true" {
                  log.Printf("CISCO STUB NOT TRUE\n")
                  doThat(epgToBeCreated)
                } else {
	                 log.Printf("CISCO STUB TRUE... SKIPPING BACKEND CALLS!\n")
                       }
		// [{"op":"add","path":"/metadata/labels/thisisanewlabel", "value":"hello"}]
		//patchTemplate := `[{"metadata": {"annotations": {"opflex.cisco.com/endpoint-group": "%v"}}}]`
                patchTemplate := `[{"op":"add","path":"/metadata/labels", "value":{"opflex.cisco.com/endpoint-group=%v"}}]`
		patch := fmt.Sprintf(patchTemplate, epgToBeCreated)
		patchB64 := base64.URLEncoding.EncodeToString([]byte(patch))

		admissionBytes, err := ioutil.ReadFile("admission.json")
		if err != nil {
			panic(err)
		}
		admissionTemplate := string(admissionBytes)
		admission := fmt.Sprintf(admissionTemplate, patchB64)
		log.Printf("Generated admission: %v\n", admission)
		_, err = fmt.Fprint(w, admission)

		if err != nil {
			log.Println(err)
		}
	}
}

func doThat(epgToBeCreated string) {
	log.Printf("EPG name to be created: %v\n", epgToBeCreated)
	tokenURL := fmt.Sprintf("https://%v/api/mo/aaaLogin.xml", apicURL)
	otherURL := fmt.Sprintf("https://%v/api/node/mo/uni/tn-%v/ap-kubernetes/epg-%v.json", apicURL, openshiftTenant, epgToBeCreated)
	xmlAuth := fmt.Sprintf("<aaaUser name='%v' pwd='%v'/>", apicUsername, apicPassword)
	xmlAuthBytes := []byte(xmlAuth)
	log.Printf("Getting auth token for user: %v\n", apicUsername)
	log.Printf("Sending a POST request to %v containing %v", tokenURL, xmlAuth)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
	imDataRaw, _ := ioutil.ReadAll(resp.Body)
	token, err := ExtractToken(imDataRaw)
	if err != nil {
		panic(err)
	}
	log.Printf("Received Auth token:\n %v", token)
	log.Println("Loading answer.json template")
	jsonAnswerBytes, err := ioutil.ReadFile("answer.json")
	if err != nil {
		panic(err)
	}
	jsonAnswerTemplate := string(jsonAnswerBytes)
	jsonAnswer := fmt.Sprintf(jsonAnswerTemplate, openshiftTenant, epgToBeCreated, epgToBeCreated, openshiftTenant, openshiftTenant)
	log.Println(jsonAnswerTemplate)
	log.Printf("Sending a POST request with the token to %v with this json:\n %v", otherURL, jsonAnswer)

	cookie := http.Cookie{
		Name:  "APIC-cookie",
		Value: token,
		Expires: time.Now().Add(600*time.Second),
	}
	req, err = http.NewRequest("POST", otherURL, bytes.NewBuffer([]byte(jsonAnswer)))
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

func ExtractToken(imDataRaw []byte) (string, error) {
	var imData Imdata
	err := xml.Unmarshal([]byte(imDataRaw), &imData)
	token := imData.AaaLogin.Token
	return token, err
}

func StartServer() {
	log.Println("Starting Ciscogate server")
	printEnvironment()
	r := mux.NewRouter()

	r.HandleFunc("/ciscogate/{epg}", CiscoGateHandler).Methods("GET")
	r.HandleFunc("/ciscogate", CiscoGateHandler).Methods("POST")

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
