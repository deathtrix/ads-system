package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	// "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"

	guuid "github.com/google/uuid"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from SSP")
}

const sspName = "ssp222"
const sspCookieName = sspName + "_cookie_id"
const sspCookieIn = "ssp1_cookie"
const dspNameIn = "dsp_name"
const resyncIn = "resync"

// DSP structure with DSP properties
type DSP map[string]interface{}

var client *redis.Client

func check(err error) {
	if err != nil {
		panic(err)
	}
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
				fmt.Fprintf(w, key+": "+ts.String()+"\n")
			} else {
				fmt.Fprintf(w, key+": "+value+"\n")
			}
		}
	} else {
		fmt.Fprintf(w, "Audience cookie not found")
	}
}

func cookieSyncHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dspName := query[dspNameIn][0]
	dspCookie := query[sspCookieIn][0]
	resync := query[resyncIn][0]

	audMap, err := client.HGetAll("aud:" + dspCookie).Result()
	if err != nil {
		panic(err)
	}

	var id string
	if len(audMap) > 0 {
		if len(audMap[sspCookieName]) > 0 {
			id = audMap[sspCookieName]
		} else {
			id = guuid.New().String()
		}
	}
	aud := map[string]interface{}{
		"timestamp":   time.Now().Unix(),
		"dspName":     dspName,
		sspCookieName: id,
	}
	addAudience(dspCookie, aud)

	// return cookie
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: sspCookieName, Value: id, Expires: expiration, Path: "/"}
	http.SetCookie(w, &cookie)

	w.Header().Add("Content-Type", "image/gif")
	if resync == "1" {
		// TODO: load DSP sync-url from DB
		dspSyncURL := "http://localhost:5050/resync.gif"
		dsp := dspSyncURL + "?ssp_name=" + sspName + "&ssp_cookie=" + id
		http.Redirect(w, r, dsp, 301)
	} else {
		w.Write([]byte(""))
	}
}

func addAudience(audID string, mapSSP map[string]interface{}) {
	check(client.HMSet("aud:"+audID, mapSSP).Err())
	client.SAdd("audience:index", "ssp:"+audID)
}

func newRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redisssp:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	return client
}

func main() {
	client = newRedisClient()

	http.HandleFunc("/sync.gif", cookieSyncHandler)
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/", rootHandler)

	s := &http.Server{
		Addr:           ":6060",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
