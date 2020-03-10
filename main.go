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

// TrailheadData represent a list of Users on trailhead.salesforce.com
type TrailheadData struct {
	Actions []struct {
		ID          string `json:"id"`
		State       string `json:"state"`
		ReturnValue struct {
			ReturnValue struct {
				Body              string `json:"body"`
				IsMyTrailheadUser bool   `json:"isMyTrailheadUser"`
			} `json:"returnValue"`
			Cacheable bool `json:"cacheable"`
		} `json:"returnValue"`
		Error []interface{} `json:"error"`
	} `json:"actions"`
}

func trailblazerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	callApexExecAndWriteToPage(w, r, "message=%7B%22actions%22%3A%5B%7B%22id%22%3A%22105%3Ba%22%2C%22descriptor%22%3A%22aura%3A%2F%2FApexActionController%2FACTION%24execute%22%2C%22callingDescriptor%22%3A%22UNKNOWN%22%2C%22params%22%3A%7B%22namespace%22%3A%22%22%2C%22classname%22%3A%22TrailheadProfileService%22%2C%22method%22%3A%22fetchTrailheadData%22%2C%22params%22%3A%7B%22userId%22%3A%22"+userID+"%22%2C%22language%22%3A%22en-US%22%7D%2C%22cacheable%22%3Afalse%2C%22isContinuation%22%3Afalse%7D%7D%5D%7D&aura.context=%7B%22mode%22%3A%22PROD%22%2C%22fwuid%22%3A%22kHqYrsGCjDhXliyGcYtIfA%22%2C%22app%22%3A%22c%3AProfileApp%22%2C%22loaded%22%3A%7B%22APPLICATION%40markup%3A%2F%2Fc%3AProfileApp%22%3A%22ZoNFIdcxHaEP9RDPdsobUQ%22%7D%2C%22dn%22%3A%5B%5D%2C%22globals%22%3A%7B%22srcdoc%22%3Atrue%7D%2C%22uad%22%3Atrue%7D&aura.pageURI=%2Fid&aura.token=")
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

	fmt.Println(string(body))

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

	callApexExecAndWriteToPage(w, r, "message=%7B%22actions%22%3A%5B%7B%22id%22%3A%22212%3Ba%22%2C%22descriptor%22%3A%22aura%3A%2F%2FApexActionController%2FACTION%24execute%22%2C%22callingDescriptor%22%3A%22UNKNOWN%22%2C%22params%22%3A%7B%22namespace%22%3A%22%22%2C%22classname%22%3A%22TrailheadProfileService%22%2C%22method%22%3A%22fetchTrailheadBadges%22%2C%22params%22%3A%7B%22userId%22%3A%22"+userID+"%22%2C%22language%22%3A%22en-US%22%2C%22skip%22%3A0%2C%22perPage%22%3A30%2C%22filter%22%3A%22All%22%7D%2C%22cacheable%22%3Afalse%2C%22isContinuation%22%3Afalse%7D%7D%5D%7D&aura.context=%7B%22mode%22%3A%22PROD%22%2C%22fwuid%22%3A%22kHqYrsGCjDhXliyGcYtIfA%22%2C%22app%22%3A%22c%3AProfileApp%22%2C%22loaded%22%3A%7B%22APPLICATION%40markup%3A%2F%2Fc%3AProfileApp%22%3A%22ek_TM7ZsKg1GOjZ-VKN7Pg%22%7D%2C%22dn%22%3A%5B%5D%2C%22globals%22%3A%7B%22srcdoc%22%3Atrue%7D%2C%22uad%22%3Atrue%7D&aura.pageURI=&aura.token=")
}

func badgesFilterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	badgesFilter := vars["filter"]

	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}

	callApexExecAndWriteToPage(w, r, "message=%7B%22actions%22%3A%5B%7B%22id%22%3A%22212%3Ba%22%2C%22descriptor%22%3A%22aura%3A%2F%2FApexActionController%2FACTION%24execute%22%2C%22callingDescriptor%22%3A%22UNKNOWN%22%2C%22params%22%3A%7B%22namespace%22%3A%22%22%2C%22classname%22%3A%22TrailheadProfileService%22%2C%22method%22%3A%22fetchTrailheadBadges%22%2C%22params%22%3A%7B%22userId%22%3A%22"+userID+"%22%2C%22language%22%3A%22en-US%22%2C%22skip%22%3A0%2C%22perPage%22%3A30%2C%22filter%22%3A%22"+badgesFilter+"%22%7D%2C%22cacheable%22%3Afalse%2C%22isContinuation%22%3Afalse%7D%7D%5D%7D&aura.context=%7B%22mode%22%3A%22PROD%22%2C%22fwuid%22%3A%22kHqYrsGCjDhXliyGcYtIfA%22%2C%22app%22%3A%22c%3AProfileApp%22%2C%22loaded%22%3A%7B%22APPLICATION%40markup%3A%2F%2Fc%3AProfileApp%22%3A%22ek_TM7ZsKg1GOjZ-VKN7Pg%22%7D%2C%22dn%22%3A%5B%5D%2C%22globals%22%3A%7B%22srcdoc%22%3Atrue%7D%2C%22uad%22%3Atrue%7D&aura.pageURI=&aura.token=")
}

func callApexExecAndWriteToPage(w http.ResponseWriter, r *http.Request, messagePayload string) {
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

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trailheadData.Actions[0].ReturnValue.ReturnValue.Body))
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
	r.PathPrefix("/").HandlerFunc(catchAllHandler)
	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		http.ListenAndServe(":8000", nil)
	} else {
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}
}
