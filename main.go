package main

import (
	"encoding/json"
	"io"
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
	trailheadApiUrl = "https://profile.api.trailhead.com/graphql"
	trailblazerUrl  = "https://trailblazer.me/"
	meIdUrl         = trailblazerUrl + "id/"
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

	body, err := io.ReadAll(res.Body)

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

// rankHandler returns information about a Trailblazer's rank and overall points
func rankHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	responseBody, err := doTrailheadCallout(trailhead.GetGraphqlPayload("GetTrailheadRank", vars["id"], `fragment TrailheadRank on TrailheadRank {\n __typename\n title\n requiredPointsSum\n requiredBadgesCount\n imageUrl\n}\n\nfragment PublicProfile on PublicProfile {\n __typename\n trailheadStats {\n __typename\n earnedPointsSum\n earnedBadgesCount\n completedTrailCount\n rank {\n ...TrailheadRank\n }\n nextRank {\n ...TrailheadRank\n }\n }\n}\n\nquery GetTrailheadRank($slug: String, $hasSlug: Boolean!) {\n profile(slug: $slug) @include(if: $hasSlug) {\n ... on PublicProfile {\n ...PublicProfile\n }\n ... on PrivateProfile {\n __typename\n }\n }\n}\n`))

	var trailheadRankData trailhead.Rank
	json.Unmarshal([]byte(responseBody), &trailheadRankData)

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No rank data returned from Trailhead."}`, 503)
	} else if trailheadRankData.Data.Profile.TrailheadStats.Typename != "" {
		encodeAndWriteToBrowser(w, trailheadRankData.Data)
	}
}

// skillsHandler returns information about a Trailblazer's skills
func skillsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	responseBody, err := doTrailheadCallout(trailhead.GetGraphqlPayload("GetEarnedSkills", vars["id"], `fragment EarnedSkill on EarnedSkill {\n __typename\n earnedPointsSum\n id\n itemProgressEntryCount\n skill {\n __typename\n apiName\n id\n name\n}\n}\n\nquery GetEarnedSkills($slug: String, $hasSlug: Boolean!) {\n profile(slug: $slug) @include(if: $hasSlug) {\n __typename\n ... on PublicProfile {\n id\n earnedSkills {\n ...EarnedSkill\n}\n}\n}\n}`))

	var trailheadSkillsData trailhead.Skills
	json.Unmarshal([]byte(responseBody), &trailheadSkillsData)

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No skills data returned from Trailhead."}`, 503)
	} else if trailheadSkillsData.Data.Profile.EarnedSkills != nil {
		encodeAndWriteToBrowser(w, trailheadSkillsData.Data)
	}
}

// certificationsHandler gets Salesforce certifications the Trailblazer has earned.
func certificationsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	responseBody, err := doTrailheadCallout(trailhead.GetGraphqlPayload("GetUserCertifications", vars["id"], `query GetUserCertifications($slug: String, $hasSlug: Boolean!) {\n profile(slug: $slug) @include(if: $hasSlug) {\n __typename\n id\n ... on PublicProfile {\n credential {\n messages {\n __typename\n body\n header\n location\n image\n cta {\n __typename\n label\n url\n }\n orientation\n }\n messagesOnly\n brands {\n __typename\n id\n name\n logo\n }\n certifications {\n cta {\n __typename\n label\n url\n }\n dateCompleted\n dateExpired\n downloadLogoUrl\n logoUrl\n infoUrl\n maintenanceDueDate\n product\n publicDescription\n status {\n __typename\n title\n expired\n date\n color\n order\n }\n title\n }\n }\n }\n }\n}\n`))

	var trailheadCertificationsData trailhead.Certifications
	json.Unmarshal([]byte(responseBody), &trailheadCertificationsData)

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No certification data returned from Trailhead."}`, 503)
	} else if trailheadCertificationsData.Data.Profile.Credential.Certifications != nil {
		encodeAndWriteToBrowser(w, trailheadCertificationsData.Data)
	}
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
		writeErrorToBrowser(w, `{"error":"Expected badge filter to be one of: MODULE, PROJECT, SUPERBADGE, EVENT, STANDALONE."}`, 501)
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

	responseBody, err := doTrailheadCallout(trailhead.GetGraphqlBadgesPayload("GetTrailheadBadges", vars["id"], `fragment EarnedAward on EarnedAwardBase {\n __typename\n id\n award {\n __typename\n id\n title\n type\n icon\n content {\n __typename\n webUrl\n description\n }\n }\n}\n\nfragment EarnedAwardSelf on EarnedAwardSelf {\n __typename\n id\n award {\n __typename\n id\n title\n type\n icon\n content {\n __typename\n webUrl\n description\n }\n }\n earnedAt\n earnedPointsSum\n}\n\nfragment StatsBadgeCount on TrailheadProfileStats {\n __typename\n earnedBadgesCount\n superbadgeCount\n}\n\nfragment ProfileBadges on PublicProfile {\n __typename\n trailheadStats {\n ... on TrailheadProfileStats {\n ...StatsBadgeCount\n }\n }\n earnedAwards(first: $count, after: $after, awardType: $filter) {\n edges {\n node {\n ... on EarnedAwardBase {\n ...EarnedAward\n }\n ... on EarnedAwardSelf {\n ...EarnedAwardSelf\n }\n }\n }\n pageInfo {\n ...PageInfoBidirectional\n }\n }\n}\n\nfragment PageInfoBidirectional on PageInfo {\n __typename\n endCursor\n hasNextPage\n startCursor\n hasPreviousPage\n}\n\nquery GetTrailheadBadges($slug: String, $hasSlug: Boolean!, $count: Int = 8, $after: String = null, $filter: AwardTypeFilter = null) {\n profile(slug: $slug) @include(if: $hasSlug) {\n __typename\n ... on PublicProfile {\n ...ProfileBadges\n }\n }\n}\n`, badgeRequestStruct))

	var trailheadBadgeData trailhead.Badges
	json.Unmarshal([]byte(responseBody), &trailheadBadgeData)

	if err != nil {
		writeErrorToBrowser(w, `{"error":"No badge data returned from Trailhead."}`, 503)
	} else if trailheadBadgeData.Data.Profile.EarnedAwards.Edges != nil {
		encodeAndWriteToBrowser(w, trailheadBadgeData.Data)
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
