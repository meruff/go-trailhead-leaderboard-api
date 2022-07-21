package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/meruff/go-trailhead-leaderboard-api/trailhead"
)

const (
	trailblazerUrl      = "https://trailblazer.me/"
	supabaseUrl         = "https://nvmegutajwfwssbzpgdb.functions.supabase.co/"
	meIdUrl             = trailblazerUrl + "id/"
	apexExecUrl         = trailblazerUrl + "aura?r=0&aura.ApexAction.execute=2"
	profileAppConfigUrl = trailblazerUrl + "c/ProfileApp.app?aura.format=JSON&aura.formatAdapter=LIGHTNING_OUT"
	rankUrl             = supabaseUrl + "rank"
	badgesUrl           = supabaseUrl + "badges"
	skillsUrl           = supabaseUrl + "skills"
)

var auraContext = ""

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/trailblazer/{id}", profileHandler)
	r.HandleFunc("/trailblazer/{id}/profile", profileHandler)
	r.HandleFunc("/trailblazer/{id}/rank", rankHandler)
	r.HandleFunc("/trailblazer/{id}/skills", skillsHandler)
	r.HandleFunc("/trailblazer/{id}/certifications", certificationsHandler)
	r.HandleFunc("/trailblazer/{id}/badges", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}/{count}", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}/{count}/{after}", badgesHandler)
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

// profileHandler gets profile information of the Trailblazer i.e. Name, Location, Company, Title etc.
// Uses a Trailblazer handle only, not an ID.
func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAlias := vars["id"]

	if strings.HasPrefix(userAlias, "005") {
		writeErrorToBrowser(w, `{"error":"/profile requires a Trailblazer handle, not an ID as a parameter."}`, 503)
		return
	}

	res, err := http.Get(meIdUrl + userAlias)

	if res != nil {
		defer res.Body.Close()
	}

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
	writeJSONToBrowser(w, out)
}

// skillsHandler returns information about a Trailblazer's skills
func skillsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	if userID == "" {
		writeErrorToBrowser(w, fmt.Sprintf(`{"error":"Could not retrieve trailhead Id with provided alias %s"}`, userID), 503)
		return
	}

	responseBody, err := doSupabaseCallout(skillsUrl, fmt.Sprintf(`{
		"queryProfile": true,
		"trailblazerId": "%s"
	}`, userID))

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	} else {
		writeJSONToBrowser(w, responseBody)
	}
}

// rankHandler returns information about a Trailblazer's rank and overall points
func rankHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	if userID == "" {
		writeErrorToBrowser(w, fmt.Sprintf(`{"error":"Could not retrieve trailhead Id with provided alias %s"}`, userID), 503)
		return
	}

	responseBody, err := doSupabaseCallout(rankUrl, fmt.Sprintf(`{
		"queryProfile": true,
		"trailblazerId": "%s"
	}`, userID))

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	} else {
		writeJSONToBrowser(w, responseBody)
	}
}

// badgeshandler gets badges the Trailblazer has earned. Returns first 30. Optionally can
// provide filter criteria, or offset i.e. "event" type badges, offset by 30.
func badgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	if userID == "" {
		writeErrorToBrowser(w, fmt.Sprintf(`{"error":"Could not retrieve trailhead Id with provided alias %s"}`, userID), 503)
		return
	}

	filter, after, count := vars["filter"], vars["after"], vars["count"]

	if filter == "" {
		filter = "MODULE"
	} else {
		filter = strings.ToUpper(filter)
	}

	countConvert, err := strconv.Atoi(count)

	if countConvert == 0 {
		countConvert = 8
	}

	badgeRequestBody, err := json.Marshal(trailhead.BadgeRequest{
		QueryProfile:  true,
		TrailblazerId: userID,
		Filter:        filter,
		After:         after,
		Count:         countConvert})

	log.Println(string(badgeRequestBody))
	responseBody, err := doSupabaseCallout(badgesUrl, string(badgeRequestBody))

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No data returned from Trailhead."}`, 503)
	} else {
		writeJSONToBrowser(w, responseBody)
	}
}

// certificationsHandler gets Salesforce certifications the Trailblazer has earned.
func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	if userID == "" {
		writeErrorToBrowser(w, fmt.Sprintf(`{"error":"Could not retrieve trailhead Id with provided alias %s"}`, userID), 503)
		return
	}

	trailheadData := doTrailheadAuraCallout(trailhead.GetApexAction("AchievementService", "fetchAchievements", userID, "", ""), "")

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
		res, err := http.Get(meIdUrl + userAlias)

		if res != nil {
			defer res.Body.Close()
		}

		if err != nil {
			log.Println(err)
			writeErrorToBrowser(w, `{"error":"Problem retrieving Trailblazer ID."}`, 503)
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Println(err)
			writeErrorToBrowser(w, `{"error":"Problem retrieving Trailblazer ID."}`, 503)
		}

		userID, strBody := "", string(body)

		// Try finding userID using TDIDUserId__c if present in response.
		var index = strings.Index(strBody, `"TBIDUserId__c":"005`)

		if -1 != index {
			userID = string(strBody[index+17 : index+35])
		}

		// Try parsing userID from profileData.
		if !strings.HasPrefix(userID, "005") {
			index = strings.Index(strBody, `\"Id\":\"`)

			if -1 != index {
				userID = string(strBody[index+9 : index+27])
			}
		}

		// Fall back to trying uid.
		if !strings.HasPrefix(userID, "005") {
			index = strings.Index(strBody, "uid: '005")

			if -1 != index {
				userID = string(strBody[index+6 : index+24])
			}
		}

		// If no ID found, write to browser and return empty string.
		if !strings.HasPrefix(userID, "005") {
			writeErrorToBrowser(w, fmt.Sprintf(`{"error":"Could not find Trailhead ID for user: %s(%s)'. Does this profile exist? Is it set to public?"}%s`, userAlias, userID, strBody), 404)
			return ""
		}

		return userID
	}

	return userAlias
}

// doTrailheadAuraCallout wraps doTrailheadCallout specifically for calls to the Profile App for Aura which needs the FwUID.
// It will retreive the FwUID if unknown or if the initial call fails and retry the call so that the calling method does not
// need to know about the FwUID
func doTrailheadAuraCallout(apexAction string, pageURI string) trailhead.Data {
	// If config has been retrieved, try aura call
	if 0 != len(auraContext) {
		var trailheadData = doTrailheadCallout(
			`message={"actions":[` + apexAction + `]}` +
				`&aura.context=` + auraContext + `&aura.pageURI=` + pageURI + `&aura.token="`)

		// If the response is not nil, call was successful
		if trailheadData.Actions != nil {
			return trailheadData
		}

		// Else  the response is nil, try getting the new fwuid and retry call before failing
	}

	// Get fwuid from profile app config
	updateAuraProfileAppConfig()

	// Make aura call
	if 0 != len(auraContext) {
		return doTrailheadCallout(
			`message={"actions":[` + apexAction + `]}` +
				`&aura.context=` + auraContext + `&aura.pageURI=` + pageURI + `&aura.token="`)
	}

	return trailhead.Data{Actions: nil}
}

// updateAuraProfileAppConfig retrives the profile app config to extract the aura context
func updateAuraProfileAppConfig() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", profileAppConfigUrl, nil)

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Referer", "https://trailblazer.me/id")
	req.Header.Add("Origin", "https://trailblazer.me")
	req.Header.Add("DNT", "1")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)

	// Deserialize the entire app config
	var profileAppConfig trailhead.ProfileAppConfig
	json.Unmarshal(body, &profileAppConfig)

	if 0 != len(profileAppConfig.AuraConfig.Context.FwUID) {
		bytes, err := json.Marshal(profileAppConfig.AuraConfig.Context.Loaded)

		if err != nil {
			log.Println(err)
		}

		auraContext = trailhead.GetAuraContext(profileAppConfig.AuraConfig.Context.FwUID, string(bytes))
	} else {
		auraContext = ""
	}
}

// doTrailheadCallout does the callout and returns the Apex REST response from Trailhead.
func doTrailheadCallout(messagePayload string) trailhead.Data {
	client := &http.Client{}
	req, err := http.NewRequest("POST", apexExecUrl, strings.NewReader(messagePayload))

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

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var trailheadData trailhead.Data
	json.Unmarshal(body, &trailheadData)

	return trailheadData
}

// doSupabaseCallout make a callout to the given URL using the given payload.
func doSupabaseCallout(url string, payload string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(payload))

	if err != nil {
		log.Println(err)
	}

	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	return string(body), err
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
