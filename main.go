package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"./trailhead"
)

const (
	trailblazerMe         = "https://trailblazer.me/id/"
	trailblazerMeApexExec = "https://trailblazer.me/aura?r=0&aura.ApexAction.execute=1"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/trailblazer/{id}", trailblazerHandler)
	r.HandleFunc("/trailblazer/{id}/profile", profileHandler)
	r.HandleFunc("/trailblazer/{id}/badges", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}/{offset}", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/certifications", certificationsHandler)
	r.PathPrefix("/").HandlerFunc(catchAllHandler)
	r.Use(loggingHandler)
	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		http.ListenAndServe(":8000", nil)
	} else {
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}
}

// trailblazerHandler gets a basic overview of the Trailblazer i.e. profile counts, recent badges, and skills.
func trailblazerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])
	if userID == "" {
		writeErrorToBrowser(w, `{"error":"Could not find valid User for this handle."}`, 503)
		return
	}

	var trailheadData = doTrailheadCallout(
		`message={"actions":[` + trailhead.GetApexAction("TrailheadProfileService", "fetchTrailheadData", userID, "", "") + `]}` +
			`&aura.context=` + trailhead.GetAuraContext() + `&aura.pageURI=/id&aura.token="`)

	if trailheadData.Actions != nil {
		writeJSONToBrowser(w, trailheadData.Actions[0].ReturnValue.ReturnValue.Body)
	} else {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	}
}

// profileHandler gets profile information of the Trailblazer i.e. Name, Location, Company, Title etc.
// Uses a Trailblazer handle only, not an ID.
func profileHandler(w http.ResponseWriter, r *http.Request) {
	calloutURL := trailblazerMe
	vars := mux.Vars(r)
	userAlias := vars["id"]
	if strings.HasPrefix(userAlias, "005") {
		writeErrorToBrowser(w, `{"error":"/profile requires a Trailblazer handle, not an ID as a parameter."}`, 503)
		return
	}

	res, err := http.Get(calloutURL + userAlias)
	if err != nil {
		log.Println(err)
		writeErrorToBrowser(w, `{"error":"Problem retrieving profile data."}`, 503)
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		writeErrorToBrowser(w, `{"error":"Problem retrieving profile data."}`, 503)
		return
	}

	jsonString := strings.Replace(string(body), "\\'", "\\\\'", -1)
	if !strings.Contains(jsonString, "var profileData = JSON.parse(") {
		writeErrorToBrowser(w, `{"error":"Problem retrieving profile data."}`, 503)
		return
	}

	jsonString = jsonString[strings.Index(jsonString, "var profileData = JSON.parse(")+29 : strings.Index(jsonString, "trailblazer.me\\\"}\");")+18]

	out, err := strconv.Unquote(jsonString)
	if err != nil {
		log.Println(err)
		writeErrorToBrowser(w, `{"error":"Problem retrieving profile data."}`, 503)
		return
	}
	out = strings.Replace(out, "\\'", "'", -1)

	defer res.Body.Close()

	writeJSONToBrowser(w, out)
}

// badgeshandler gets badges the Trailblazer has earned. Returns first 30. Optionally can
// provide filter criteria, or offset i.e. "event" type badges, offset by 30.
func badgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])
	if userID == "" {
		writeErrorToBrowser(w, `{"error":"Could not find valid User for this handle."}`, 503)
		return
	}

	badgesFilter := vars["filter"]
	skip := vars["offset"]

	if skip == "" {
		skip = "0"
	}

	var trailheadData = doTrailheadCallout(
		`message={"actions":[` + trailhead.GetApexAction("TrailheadProfileService", "fetchTrailheadBadges", userID, skip, badgesFilter) + `]}` +
			`&aura.context=` + trailhead.GetAuraContext() + `&aura.pageURI=&aura.token="`)

	if trailheadData.Actions != nil {
		writeJSONToBrowser(w, trailheadData.Actions[0].ReturnValue.ReturnValue.Body)
	} else {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	}
}

// certificationsHandler gets Salesforce certifications the Trailblazer has earned.
func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])
	if userID == "" {
		writeErrorToBrowser(w, `{"error":"Could not find valid User for this handle."}`, 503)
		return
	}

	var trailheadData = doTrailheadCallout(
		`message={"actions":[` + trailhead.GetApexAction("AchievementService", "fetchAchievements", userID, "", "") + `]}` +
			`&aura.context=` + trailhead.GetAuraContext() + `&aura.pageURI=&aura.token="`)

	if trailheadData.Actions != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trailheadData.Actions[0].ReturnValue.ReturnValue.CertificationsResult)
	} else {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	}
}

// loggingHandler logs time spent to access each request/what page was requested.
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	})
}

// catchAllHandler is the default message if no Trailblazer Id or handle is provided,
// or if the u	ser has navigated to an unsupported page.
func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	writeErrorToBrowser(w, `{"error":"Please provide a valid Trialhead user Id/handle or visit a valid URL. Example: /trailblazer/{id}"}`, 501)
}

// getTrailheadID gets the Trailblazer's user Id from Trailhead, if provided with a custom user handle i.e. "matruff" => "0051I000004UgTlQAK"
func getTrailheadID(w http.ResponseWriter, userAlias string) string {
	if !strings.HasPrefix(userAlias, "005") {
		res, err := http.Get(trailblazerMe + userAlias)
		if err != nil {
			log.Println(err)
			writeErrorToBrowser(w, `{"error":"Problem retrieving Trailblazer ID."}`, 503)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println(err)
			writeErrorToBrowser(w, `{"error":"Problem retrieving Trailblazer ID."}`, 503)
		}

		defer res.Body.Close()

		userID := string(string(body)[strings.Index(string(body), "uid: ")+6 : strings.Index(string(body), "uid: ")+24])

		if !strings.HasPrefix(userID, "005") {
			writeErrorToBrowser(w, `{"error":"Could not find Trailhead ID for user: '`+userAlias+`'. Does this profile exist? Is it set to public?"}`, 404)
			return ""
		}

		return userID
	}

	return userAlias
}

// doTrailheadCallout does the callout and returns the Apex REST response from Trailhead.
func doTrailheadCallout(messagePayload string) trailhead.Data {
	client := &http.Client{}
	req, err := http.NewRequest("POST", trailblazerMeApexExec, strings.NewReader(messagePayload))
	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Referer", "https://trailblazer.me/id")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	req.Header.Add("Origin", "https://trailblazer.me")
	req.Header.Add("DNT", "1")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	var trailheadData trailhead.Data
	json.Unmarshal(body, &trailheadData)

	defer res.Body.Close()

	return trailheadData
}

// writeJSONToBrowser simply writes a provided string to the browser in JSON format with optional HTTP code.
func writeJSONToBrowser(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
}

// writeErrorToBrowser writes an HTTP error to the broswer in JSON.
func writeErrorToBrowser(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(err))
}
