package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/meruff/go-trailhead-leaderboard-api/trailhead"
)

const (
	trailheadApiUrl = "https://profile.api.trailhead.com/graphql"
	trailblazerUrl  = "https://www.salesforce.com/trailblazer/"
)

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

// profileHandler gets profile information of the Trailblazer i.e. Name, Company, Title etc. Uses a
// Trailblazer handle only, not an ID.
func profileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userAlias := vars["id"]

	if strings.HasPrefix(userAlias, "005") {
		writeErrorToBrowser(w, "/profile requires a trailblazer handle, not an ID as a parameter.", 503)
		return
	}

	res, err := http.Get(trailblazerUrl + userAlias)
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		log.Println(err)
		writeErrorToBrowser(w, "Problem retrieving profile data.", 503)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		writeErrorToBrowser(w, "Problem reading profile body.", 503)
		return
	}

	if !strings.Contains(string(body), "var profile = ") {
		writeErrorToBrowser(
			w,
			fmt.Sprintf("Cannot find profile data for %s. Does this trailblazer exist?", vars["id"]),
			503,
		)
		return
	}

	re := regexp.MustCompile(`var profile = (.*);`)
	match := re.FindStringSubmatch(string(body))

	if len(match) > 1 {
		var trailheadProfileData trailhead.Profile
		json.Unmarshal([]byte(match[1]), &trailheadProfileData)

		profileDataForUi := trailhead.ProfileReturn{}
		profileDataForUi.ProfilePhotoUrl = trailheadProfileData.PhotoURL
		profileDataForUi.ProfileUser.TBID_Role = trailheadProfileData.Role
		profileDataForUi.ProfileUser.CompanyName = trailheadProfileData.Company.Name
		profileDataForUi.ProfileUser.TrailblazerId = vars["id"]
		profileDataForUi.ProfileUser.Title = trailheadProfileData.Title
		profileDataForUi.ProfileUser.FirstName = trailheadProfileData.FirstName
		profileDataForUi.ProfileUser.LastName = trailheadProfileData.LastName
		profileDataForUi.ProfileUser.Id = trailheadProfileData.ID
		encodeAndWriteToBrowser(w, profileDataForUi)
	} else {
		writeErrorToBrowser(w, "No profile data found.", 503)
	}
}

// rankHandler returns information about a Trailblazer's rank and overall points
func rankHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	responseBody, err := doTrailheadCallout(
		trailhead.GetGraphqlPayload("GetTrailheadRank", vars["id"], "", trailhead.GetRankQuery()),
	)
	if err != nil {
		writeErrorToBrowser(w, "No rank data returned from Trailhead.", 503)
	}

	var trailheadRankData trailhead.Rank
	json.Unmarshal([]byte(responseBody), &trailheadRankData)
	encodeAndWriteToBrowser(w, trailheadRankData.Data)
}

// skillsHandler returns information about a Trailblazer's skills
func skillsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	responseBody, err := doTrailheadCallout(
		trailhead.GetGraphqlPayload(
			"GetEarnedSkills",
			vars["id"],
			"",
			trailhead.GetSkillsQuery(),
		),
	)
	if err != nil {
		writeErrorToBrowser(w, "No skills data returned from Trailhead.", 503)
	}

	var trailheadSkillsData trailhead.Skills
	json.Unmarshal([]byte(responseBody), &trailheadSkillsData)
	encodeAndWriteToBrowser(w, trailheadSkillsData.Data)
}

// certificationsHandler gets Salesforce certifications the Trailblazer has earned.
func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	responseBody, err := doTrailheadCallout(
		trailhead.GetGraphqlPayload(
			"GetUserCertifications",
			vars["id"],
			"",
			trailhead.GetCertificationsQuery(),
		),
	)
	if err != nil {
		writeErrorToBrowser(w, "No certification data returned from Trailhead.", 503)
	}

	var trailheadCertificationsData trailhead.Certifications
	json.Unmarshal([]byte(responseBody), &trailheadCertificationsData)
	certificationReturnData := trailhead.CertificationsReturn{}

	for _, certification := range trailheadCertificationsData.Data.Profile.Credential.Certifications {
		cReturn := trailhead.Certification{}
		cReturn.DateCompleted = certification.DateCompleted
		cReturn.CertificationUrl = certification.InfoURL
		cReturn.Description = certification.PublicDescription
		cReturn.CertificationStatus = certification.Status.Title
		cReturn.Title = certification.Title
		cReturn.CertificationImageUrl = certification.LogoURL

		if dateExpired, ok := certification.DateExpired.(string); ok {
			cReturn.DateExpired = dateExpired
		} else {
			cReturn.DateExpired = ""
		}

		certificationReturnData.CertificationsList = append(
			certificationReturnData.CertificationsList, cReturn,
		)
	}

	encodeAndWriteToBrowser(w, certificationReturnData)
}

// badgeshandler gets badges the Trailblazer has earned. Returns first 8. Optionally can
// provide filter criteria, or additional return count. i.e. "event" type badges, count by 30.
func badgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filter, after, count := vars["filter"], vars["after"], vars["count"]
	badgeRequestStruct := trailhead.BadgeRequest{}

	// Set filter
	if contains(getValidBadgeFilters(), filter) {
		var upperFilter = strings.ToUpper(filter)
		badgeRequestStruct.Filter = upperFilter
	} else if filter != "all" && filter != "" {
		writeErrorToBrowser(
			w,
			fmt.Sprintf("Expected badge filter to be one of: %s.", strings.Join(getValidBadgeFilters(), ", ")),
			501,
		)
		return
	}

	// Set count
	if count != "" {
		countConvert, err := strconv.Atoi(count)
		if err != nil {
			log.Println("Error parsing badge count from params.")
		}

		badgeRequestStruct.Count = countConvert
	} else {
		badgeRequestStruct.Count = 8
	}

	// Set after
	if after != "" {
		badgeRequestStruct.After = after
	}

	responseBody, err := doTrailheadCallout(
		trailhead.GetGraphqlPayload(
			"GetTrailheadBadges",
			vars["id"],
			trailhead.GetBadgesFilterPayload(vars["id"], badgeRequestStruct),
			trailhead.GetBadgesQuery(),
		),
	)
	if err != nil {
		writeErrorToBrowser(w, "No badge data returned from Trailhead.", 503)
	}

	var trailheadBadgeData trailhead.Badges
	json.Unmarshal([]byte(responseBody), &trailheadBadgeData)
	encodeAndWriteToBrowser(w, trailheadBadgeData.Data)
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
	writeErrorToBrowser(
		w,
		"Please provide a valid handle or visit a valid URL. Example: /trailblazer/{id}",
		501,
	)
}

// doTrailheadCallout makes a callout to the given URL using the given
func doTrailheadCallout(payload string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", trailheadApiUrl, strings.NewReader(payload))

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		log.Println(err)
	}

	body, err := io.ReadAll(res.Body)

	return string(body), err
}

// writeErrorToBrowser writes an HTTP error to the broswer in JSON.
func writeErrorToBrowser(w http.ResponseWriter, errorMsg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"error":"%s"`, errorMsg)))
}

// encodeAndWriteToBrowser encodes a given interface and writes it to the browser as JSON.
func encodeAndWriteToBrowser(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(i)
}

// contains simply checks if a string exists inside a slice.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// getValidBadgeFilters returns a slice containing valid filters for Trailblazer badges.
func getValidBadgeFilters() []string {
	return []string{"module", "project", "superbadge", "event", "standalone"}
}
