package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	// "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"

	guuid "github.com/google/uuid"
)

const dspName = "dsp3354"
const dspCookieName = dspName + "_cookie_id"

// SSP structure with SSP properties
type SSP map[string]interface{}

// ScriptData structure for templating
type ScriptData struct {
	DSPName string
	SSPs    []SSP
}

var client *redis.Client

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkVal(val interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return val
}

func loadStaticPage(filename string) ([]byte, error) {
	body, err := ioutil.ReadFile("./static/" + filename)
	if err != nil {
		panic(err)
	}
	return body, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from DSP")
}

func addSSPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		ssp := make(map[string]interface{})
		err := decoder.Decode(&ssp)
		if err != nil {
			panic(err)
		}
		addSSP(ssp)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "SSP saved"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func cookieSyncHandler(w http.ResponseWriter, r *http.Request) {
	audIDCookie, _ := r.Cookie(dspCookieName)
	audID := audIDCookie.Value
	var id string
	if len(audID) > 0 {
		id = audID
	} else {
		id = guuid.New().String()
	}
	aud := map[string]interface{}{
		"timestamp": time.Now().Unix(),
	}
	addAudience(id, aud)
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: dspCookieName, Value: id, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)

	// for ssp - if resync - redirect back to dsp
	// dsp := "http://www.google.com"
	// http.Redirect(w, r, dsp, 301)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	audID := r.URL.Path[len("/status/"):]
	audMap, err := client.HGetAll("aud:" + audID).Result()
	if err != nil {
		panic(err)
	}
	if len(audMap) > 0 {
		for key, value := range audMap {
			if key == "timestamp" {
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					panic(err)
				}
				ts := time.Unix(i, 0)
				fmt.Fprintf(w, key+": "+ts.String())
			} else {
				fmt.Fprintf(w, key+": "+value)
			}
		}
	} else {
		fmt.Fprintf(w, "Audience cookie not found")
	}
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/js/"):]
	template, err := template.ParseFiles("./static/" + title)
	if err != nil {
		panic(err)
	}

	// load SSPs from redis
	var ssps []SSP
	sspsList, err := client.SMembers("ssps:index").Result()
	if err != nil {
		panic(err)
	}
	for _, sspName := range sspsList {
		sspMap, err := client.HGetAll(sspName).Result()
		if err != nil {
			panic(err)
		}
		ssp := SSP{"url": sspMap["sync-url"], "cookie": sspMap["cookie-name"], "resync": sspMap["resync"]}
		ssps = append(ssps, ssp)
	}

	check(template.Execute(w, ScriptData{DSPName: dspName, SSPs: ssps}))
}

func addSSP(mapSSP map[string]interface{}) {
	sspID := fmt.Sprintf("%v", checkVal(client.Incr("ssp_counter").Result()))
	check(client.HMSet("ssp:"+sspID, mapSSP).Err())
	client.SAdd("ssps:index", "ssp:"+sspID)
}

func addAudience(audID string, mapSSP map[string]interface{}) {
	check(client.HMSet("aud:"+audID, mapSSP).Err())
	client.SAdd("audience:index", "ssp:"+audID)
}

func newRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redisdsp:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	return client
}

func seedDatabase() {
	// add new SSP
	ssp := map[string]interface{}{
		"name":             "ssp1",
		"sync-url":         "http://localhost:6060/sync.gif",
		"sync-details-url": "http://localhost:6060/usersync",
		"cookie-name":      "ssp1_cookie",
		"resync":           "1",
	}
	addSSP(ssp)
}

func main() {
	client = newRedisClient()
	seedDatabase()

	// handle HTTP requests
	http.HandleFunc("/pixelSync.gif", cookieSyncHandler)
	http.HandleFunc("/js/", scriptHandler)
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/add-ssp", addSSPHandler)
	http.HandleFunc("/", rootHandler)

	// bring HTTP server up
	s := &http.Server{
		Addr:           ":5050",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
