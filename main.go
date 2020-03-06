package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

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

// Page refers to the literal Page we're going to display
type Page struct {
	Response string
}

func trailblazerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	if !strings.HasPrefix(userID, "005") {
		userID = getTrailheadID(userID)
	}
	url := "https://trailblazer.me/aura?r=0&aura.ApexAction.execute=1"
	method := "POST"
	//0051I000004UgTlQAK
	payload := strings.NewReader("message=%7B%22actions%22%3A%5B%7B%22id%22%3A%22105%3Ba%22%2C%22descriptor%22%3A%22aura%3A%2F%2FApexActionController%2FACTION%24execute%22%2C%22callingDescriptor%22%3A%22UNKNOWN%22%2C%22params%22%3A%7B%22namespace%22%3A%22%22%2C%22classname%22%3A%22TrailheadProfileService%22%2C%22method%22%3A%22fetchTrailheadData%22%2C%22params%22%3A%7B%22userId%22%3A%22" + userID + "%22%2C%22language%22%3A%22en-US%22%7D%2C%22cacheable%22%3Afalse%2C%22isContinuation%22%3Afalse%7D%7D%5D%7D&aura.context=%7B%22mode%22%3A%22PROD%22%2C%22fwuid%22%3A%22kHqYrsGCjDhXliyGcYtIfA%22%2C%22app%22%3A%22c%3AProfileApp%22%2C%22loaded%22%3A%7B%22APPLICATION%40markup%3A%2F%2Fc%3AProfileApp%22%3A%22ZoNFIdcxHaEP9RDPdsobUQ%22%7D%2C%22dn%22%3A%5B%5D%2C%22globals%22%3A%7B%22srcdoc%22%3Atrue%7D%2C%22uad%22%3Atrue%7D&aura.pageURI=%2Fid&aura.token=")

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
	// fmt.Println(trailheadData.Actions[0].ReturnValue.ReturnValue.Body)

	defer res.Body.Close()

	p := Page{trailheadData.Actions[0].ReturnValue.ReturnValue.Body}
	pageData, err := json.Marshal(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pageData)
}

func getTrailheadID(userAlias string) string {
	res, err := http.Get("https://trailblazer.me/id/" + userAlias)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	return string(string(body)[strings.Index(string(body), "uid: ")+6 : strings.Index(string(body), "uid: ")+24])
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{"Please provide a valid Trialhead User Id or Name at /trailblazer/{id}"}
	pageData, err := json.Marshal(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(pageData)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/trailblazer/{id}", trailblazerHandler)
	r.PathPrefix("/").HandlerFunc(catchAllHandler)
	http.Handle("/", r)

	port := os.Getenv("PORT")
	if port == "" {
		http.ListenAndServe(":8000", nil)
	} else {
		http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	}
}
