package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const trailblazerMe = "https://trailblazer.me/id/"
const trailblazerMeUserID = "https://trailblazer.me/id?cmty=trailhead&uid="
const trailblazerMeApexExec = "https://trailblazer.me/aura?r=0&aura.ApexAction.execute=1"
const fwuid = "ReJ5V8Oa_EmHa1B_VZHK_g"

// TrailheadData represent a list of Users on trailhead.salesforce.com
type TrailheadData struct {
	Actions []struct {
		ID          string `json:"id"`
		State       string `json:"state"`
		ReturnValue struct {
			ReturnValue struct {
				Body                 string `json:"body"`
				SuperbadgesResult    string `json:"superbadgesResult"`
				CertificationsResult struct {
					CertificationsList []struct {
						CertificationImageURL string `json:"certificationImageUrl"`
						CertificationURL      string `json:"certificationUrl"`
						DateCompleted         string `json:"dateCompleted"`
						Description           string `json:"description"`
						Title                 string `json:"title"`
					} `json:"certificationsList"`
					StatusCode    string `json:"statusCode"`
					StatusMessage string `json:"statusMessage"`
				} `json:"certificationsResult"`
				IsMyTrailheadUser bool `json:"isMyTrailheadUser"`
			} `json:"returnValue"`
			Cacheable bool `json:"cacheable"`
		} `json:"returnValue"`
		Error []interface{} `json:"error"`
	} `json:"actions"`
	Context struct {
		Fwuid string `json:"fwuid"`
	} `json:"context"`
}

func trailblazerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	var trailheadData = getApexExecResponse(
		`message={
			"actions":[
				{
					"id":"105;a",
					"descriptor":"aura://ApexActionController/ACTION$execute",
					"callingDescriptor":"UNKNOWN",
					"params":{
						"namespace":"",
						"classname":"TrailheadProfileService",
						"method":"fetchTrailheadData",
						"params":{
							"userId":"` + userID + `",
							"language":"en-US"
						},
						"cacheable":false,
						"isContinuation":false
					}
				}
			]
		}
		&aura.context={
			"mode":"PROD",
			"fwuid":"` + fwuid + `",
			"app":"c:ProfileApp",
			"loaded":{
				"APPLICATION@markup://c:ProfileApp":"ZoNFIdcxHaEP9RDPdsobUQ"
			},
			"dn":[],
			"globals":{
				"srcdoc":true
			},
			"uad":true
		}
		&aura.pageURI=/id
		&aura.token=`)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trailheadData.Actions[0].ReturnValue.ReturnValue.Body))
}

func getTrailheadID(userAlias string) string {
	res, err := http.Get(trailblazerMe + userAlias)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	return string(string(body)[strings.Index(string(body), "uid: ")+6 : strings.Index(string(body), "uid: ")+24])
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	var calloutURL string
	vars := mux.Vars(r)
	userAlias := vars["id"]

	if strings.HasPrefix(userAlias, "005") {
		calloutURL = trailblazerMeUserID
	} else {
		calloutURL = trailblazerMe
	}

	res, err := http.Get(calloutURL + userAlias)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	jsonString := strings.Replace(string(body), "\\'", "\\\\'", -1)
	jsonString = jsonString[strings.Index(jsonString, "var profileData = JSON.parse(")+29 : strings.Index(jsonString, "trailblazer.me\\\"}\");")+18]

	out, err := strconv.Unquote(jsonString)
	if err != nil {
		fmt.Println(err)
	}
	out = strings.Replace(out, "\\'", "'", -1)

	defer res.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(out))
}

func badgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	var trailheadData = getApexExecResponse(
		`message={
			"actions":[
				{
					"id":"212;a",
					"descriptor":"aura://ApexActionController/ACTION$execute",
					"callingDescriptor":"UNKNOWN",
					"params":{
						"namespace":"",
						"classname":"TrailheadProfileService",
						"method":"fetchTrailheadBadges",
						"params":{
							"userId":"` + userID + `",
							"language":"en-US",
							"skip":0,"perPage":30,
							"filter":"All"
						},
						"cacheable":false,
						"isContinuation":false
					}
				}
			]
		}
		&aura.context={
			"mode":"PROD",
			"fwuid":"` + fwuid + `",
			"app":"c:ProfileApp",
			"loaded":{
				"APPLICATION@markup://c:ProfileApp":"ek_TM7ZsKg1GOjZ-VKN7Pg"
			},
			"dn":[],
			"globals":{
				"srcdoc":true
			},
			"uad":true
		}
		&aura.pageURI=
		&aura.token="`)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trailheadData.Actions[0].ReturnValue.ReturnValue.Body))
}

func badgesFilterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	badgesFilter := vars["filter"]
	skip := vars["offset"]

	if skip == "" {
		skip = "0"
	}

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	var trailheadData = getApexExecResponse(
		`message={
			"actions":[
				{
					"id":"212;a",
					"descriptor":"aura://ApexActionController/ACTION$execute",
					"callingDescriptor":"UNKNOWN",
					"params":{
						"namespace":"",
						"classname":"TrailheadProfileService",
						"method":"fetchTrailheadBadges",
						"params":{
							"userId":"` + userID + `",
							"language":"en-US",
							"skip":"` + skip + `",
							"perPage":30,
							"filter":"` + strings.Title(badgesFilter) + `"
						},
						"cacheable":false,
						"isContinuation":false
					}
				}
			]
		}
		&aura.context={
			"mode":"PROD",
			"fwuid":"` + fwuid + `",
			"app":"c:ProfileApp",
			"loaded":{
				"APPLICATION@markup://c:ProfileApp":"ek_TM7ZsKg1GOjZ-VKN7Pg"
			},
			"dn":[],
			"globals":{
				"srcdoc":true
			},
			"uad":true
		}
		&aura.pageURI=
		&aura.token=`)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trailheadData.Actions[0].ReturnValue.ReturnValue.Body))
}

func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	var trailheadData = getApexExecResponse(
		`message={
			"actions":[
				{
					"id":"105;a",
					"descriptor":"aura://ApexActionController/ACTION$execute",
					"callingDescriptor":"UNKNOWN",
					"params":{
						"namespace":"",
						"classname":"AchievementService",
						"method":"fetchAchievements",
						"params":{
							"userId":"` + userID + `",
							"language":"en-US"
						},
						"cacheable":false,
						"isContinuation":false
					}
				}
			]
		}
		&aura.context={
			"mode":"PROD",
			"fwuid":"` + fwuid + `",
			"app":"c:ProfileApp",
			"loaded":{
				"APPLICATION@markup://c:ProfileApp":"ZoNFIdcxHaEP9RDPdsobUQ"
			},
			"dn":[],
			"globals":{
				"srcdoc":true
			},
			"uad":true
		}
		&aura.pageURI=/id
		&aura.token=`)

	w.Header().Set("Content-Type", "application/json")
	jsonOutput, err := json.Marshal(trailheadData.Actions[0].ReturnValue.ReturnValue.CertificationsResult)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(jsonOutput)
}

func getApexExecResponse(messagePayload string) TrailheadData {
	url := trailblazerMeApexExec
	method := "POST"
	payload := strings.NewReader(messagePayload)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
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
	var trailheadData TrailheadData
	json.Unmarshal(body, &trailheadData)

	defer res.Body.Close()

	return trailheadData
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error":"Please provide a valid Trialhead User Id or Name at /trailblazer/{id}"}`))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/trailblazer/{id}", trailblazerHandler)
	r.HandleFunc("/trailblazer/{id}/profile", profileHandler)
	r.HandleFunc("/trailblazer/{id}/badges", badgesHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}", badgesFilterHandler)
	r.HandleFunc("/trailblazer/{id}/badges/{filter}/{offset}", badgesFilterHandler)
	r.HandleFunc("/trailblazer/{id}/certifications", certificationsHandler)
	r.PathPrefix("/").HandlerFunc(catchAllHandler)
	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		http.ListenAndServe(":8000", nil)
	} else {
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}
}
