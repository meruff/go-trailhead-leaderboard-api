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
const fwuid = "axnV2upVY_ZFzdo18txAEw"

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
						CertificationStatus   string `json:"certificationStatus"`
						CertificationURL      string `json:"certificationUrl"`
						DateCompleted         string `json:"dateCompleted"`
						DateExpired           string `json:"dateExpired"`
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

// Gets a basic overview of the Trailblazer i.e. profile counts, recent badges, and skills.
func trailblazerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	var trailheadData = getApexExecResponse(
		`message={"actions":[` + getAction("TrailheadProfileService", "fetchTrailheadData", userID, "", "") + `]}` +
			`&aura.context=` + getAuraContext() + `&aura.pageURI=/id&aura.token="`)

	writeJSONToBrowser(w, trailheadData.Actions[0].ReturnValue.ReturnValue.Body)
}

// Gets profile information of the Trailblazer i.e. Name, Location, Company, Title etc.
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

	writeJSONToBrowser(w, out)
}

// Gets badges the Trailblazer has earned. Returns first 30.
func badgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	var trailheadData = getApexExecResponse(
		`message={"actions":[` + getAction("TrailheadProfileService", "fetchTrailheadBadges", userID, "0", "All") + `]}` +
			`&aura.context=` + getAuraContext() + `&aura.pageURI=&aura.token="`)

	writeJSONToBrowser(w, trailheadData.Actions[0].ReturnValue.ReturnValue.Body)
}

// Gets badges the Trailblazer has earned based on filter criteria, or offset i.e. "event" type badges, offset by 30.
func badgesFilterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])
	badgesFilter := vars["filter"]
	skip := vars["offset"]

	if skip == "" {
		skip = "0"
	}

	var trailheadData = getApexExecResponse(
		`message={"actions":[` + getAction("TrailheadProfileService", "fetchTrailheadBadges", userID, skip, badgesFilter) + `]}` +
			`&aura.context=` + getAuraContext() + `&aura.pageURI=&aura.token="`)

	writeJSONToBrowser(w, trailheadData.Actions[0].ReturnValue.ReturnValue.Body)
}

// Gets Salesforce certifications the Trailblazer has earned.
func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := getTrailheadID(w, vars["id"])

	var trailheadData = getApexExecResponse(
		`message={"actions":[` + getAction("AchievementService", "fetchAchievements", userID, "", "") + `]}` +
			`&aura.context=` + getAuraContext() + `&aura.pageURI=&aura.token="`)

	jsonOutput, err := json.Marshal(trailheadData.Actions[0].ReturnValue.ReturnValue.CertificationsResult)

	if err != nil {
		fmt.Println(err)
	}

	writeJSONToBrowser(w, string(jsonOutput))
}

// Gets the Trailblazer's user Id from Trailhead, if provided with a custom user handle i.e. "matruff" => "0051I000004XSMrQAO"
func getTrailheadID(w http.ResponseWriter, userAlias string) string {
	if !strings.HasPrefix(userAlias, "005") {
		res, err := http.Get(trailblazerMe + userAlias)
		if err != nil {
			fmt.Println(err)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}

		defer res.Body.Close()

		userID := string(string(body)[strings.Index(string(body), "uid: ")+6 : strings.Index(string(body), "uid: ")+24])

		if !strings.HasPrefix(userID, "005") {
			writeJSONToBrowser(w, `{"error":"Could not find Trailhead ID for user: '`+userAlias+`'. Does this profile exist? Is it set to public?"}`)
			return ""
		}

		return userID
	}

	return userAlias
}

// Does the callout and returns the Apex REST response from Trailhead.
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

// The default message if no Trailblazer Id or handle is provided, or if the user has navigated to an unsupported page.
func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	writeJSONToBrowser(w, `{"error":"Please provide a valid Trialhead user Id/handle or visit a valid URL. Example: /trailblazer/{id}"}`)
}

// Returns a JSON string representing an Apex action to be used in the callout to Trailhead.
func getAction(className string, methodName string, userID string, skip string, filter string) string {
	actionString :=
		`{
            "id":"212;a",
            "descriptor":"aura://ApexActionController/ACTION$execute",
            "callingDescriptor":"UNKNOWN",
            "params":{
                "namespace":"",
                "classname":"` + className + `",
                "method":"` + methodName + `",
                "params":{
                    "userId":"` + userID + `",
                    "language":"en-US"`

	if skip != "" {
		actionString += `,
                    "skip":` + skip + `,
                    "perPage":30`
	}

	if filter != "" {
		actionString += `,
                    "filter":"` + strings.Title(filter) + `"`
	}

	actionString += `
                },
                    "cacheable":false,
                    "isContinuation":false
                }
            }`

	return actionString
}

// Returns a JSON string containing the Aura "context" to use in the callout to Trailhead.
func getAuraContext() string {
	return `{
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
    }`
}

// Simply writes a provided string to the browser in JSON format.
func writeJSONToBrowser(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
}
