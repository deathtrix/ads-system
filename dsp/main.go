package main

import (
	"fmt"
	"log"
	"net/http"
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

// func loadPage(filename string) ([]byte, error) {
// 	body, err := ioutil.ReadFile("./static/" + filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return body, nil
// }

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from DSP")
}

func cookieSyncHandler(w http.ResponseWriter, r *http.Request) {
	// get data from redis

	// build and set the cookie
	id := guuid.New()

	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: dspCookieName, Value: id.String(), Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)

	// cookie, _ := r.Cookie("username")
	// redirect to ssp
	// ssp := "http://www.google.com"
	// http.Redirect(w, r, ssp, 301)
}

func scriptHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/js/"):]
	template, err := template.ParseFiles("./static/" + title)
	if err != nil {
		panic(err)
	}

	// load ssps from redis
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
		ssp := SSP{"url": sspMap["url"], "cookie": sspMap["cookie"], "resync": sspMap["resync"]}
		ssps = append(ssps, ssp)
	}

	check(template.Execute(w, ScriptData{DSPName: dspName, SSPs: ssps}))
}

func addSSP(mapSSP map[string]interface{}) {
	pipe := client.Pipeline()
	audID := fmt.Sprintf("%v", checkVal(pipe.Incr("audience_counter").Result()))
	check(pipe.HMSet("ssp:"+audID,
		mapSSP).Err())
	pipe.SAdd("ssps:index", "ssp:"+audID)
	checkVal(pipe.Exec())
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
	ssp := map[string]interface{}{
		"url":    "http://localhost:6060/sync.gif",
		"cookie": "ssp1_cookie",
		"resync": "1",
	}
	addSSP(ssp)
}

func main() {
	client = newRedisClient()

	// seed database
	seedDatabase()

	// handle HTTP requests
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/pixelSync.gif", cookieSyncHandler)
	http.HandleFunc("/js/", scriptHandler)

	// bring HTTP server up
	s := &http.Server{
		Addr:           ":5050",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
