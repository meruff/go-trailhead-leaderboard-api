package trailhead

import "strings"

// Data represents a response from trailhead.
type Data struct {
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

// Rank represents skill data returned from trailhead.
type Rank struct {
	Data struct {
		Profile struct {
			Typename       string `json:"__typename"`
			TrailheadStats struct {
				Typename            string `json:"__typename"`
				EarnedPointsSum     int    `json:"earnedPointsSum"`
				EarnedBadgesCount   int    `json:"earnedBadgesCount"`
				CompletedTrailCount int    `json:"completedTrailCount"`
				Rank                struct {
					Typename            string `json:"__typename"`
					Title               string `json:"title"`
					RequiredPointsSum   int    `json:"requiredPointsSum"`
					RequiredBadgesCount int    `json:"requiredBadgesCount"`
					ImageURL            string `json:"imageUrl"`
				} `json:"rank"`
				NextRank interface{} `json:"nextRank"`
			} `json:"trailheadStats"`
		} `json:"profile"`
	} `json:"data"`
}

// Skills represents skill data returned from trailhead.
type Skills struct {
	Profile struct {
		Typename     string `json:"__typename"`
		EarnedSkills []struct {
			Typename               string `json:"__typename"`
			EarnedPointsSum        int    `json:"earnedPointsSum"`
			ID                     string `json:"id"`
			ItemProgressEntryCount int    `json:"itemProgressEntryCount"`
			Skill                  struct {
				Typename string `json:"__typename"`
				APIName  string `json:"apiName"`
				ID       string `json:"id"`
				Name     string `json:"name"`
			} `json:"skill"`
		} `json:"earnedSkills"`
	} `json:"profile"`
}

// Badges represents skill data returned from trailhead.
type Badges struct {
	Profile struct {
		Typename     string `json:"__typename"`
		EarnedAwards struct {
			Edges []struct {
				Node struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Award    struct {
						Typename string `json:"__typename"`
						ID       string `json:"id"`
						Title    string `json:"title"`
						Type     string `json:"type"`
						Icon     string `json:"icon"`
						Content  struct {
							Typename    string `json:"__typename"`
							WebURL      string `json:"webUrl"`
							Description string `json:"description"`
						} `json:"content"`
					} `json:"award"`
					EarnedAt        string `json:"earnedAt"`
					EarnedPointsSum string `json:"earnedPointsSum"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				Typename        string `json:"__typename"`
				EndCursor       string `json:"endCursor"`
				HasNextPage     bool   `json:"hasNextPage"`
				StartCursor     string `json:"startCursor"`
				HasPreviousPage bool   `json:"hasPreviousPage"`
			} `json:"pageInfo"`
		} `json:"earnedAwards"`
	} `json:"profile"`
}

// ProfileAppConfig represents the full configuration for the Salesforce Trailhead profile app
type ProfileAppConfig struct {
	AuraConfig struct {
		Context struct {
			FwUID  string      `json:"fwuid"`
			Loaded interface{} `json:"loaded"`
		} `json:"context"`
	} `json:"auraConfig"`
}

// BadgeRequest represents a request to the /badges endpoint. The variables to send to graphql
type BadgeRequest struct {
	QueryProfile  bool    `json:"queryProfile"`
	TrailblazerId string  `json:"trailblazerId"`
	Filter        *string `json:"filter"`
	After         *string `json:"after"`
	Count         int     `json:"count"`
}

// GetAuraContext returns a JSON string containing the Aura "context" to use in the callout to Trailhead.
func GetAuraContext(fwUID string, loaded string) string {
	return `{
        "mode":"PROD",
        "fwuid":"` + fwUID + `",
        "app":"c:ProfileApp",
        "loaded":` + loaded + `,
        "dn":[],
        "globals":{
            "srcdoc":true
        },
        "uad":true
    }`
}

// GetApexAction returns a JSON string representing an Apex action to be used in the callout to Trailhead.
func GetApexAction(className string, methodName string, userID string, skip string, filter string) string {
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
                    "language":"en-US",
					"featureAdditionalCerts": true`

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

// GetGraphqlPayload returns a JSON string to use in Trailhead graphql callouts.
func GetGraphqlPayload(operationName string, userID string, query string) string {
	return `{
	"operationName": "` + operationName + `",
  	"variables": {
    	"hasSlug": true,
    	"slug": "` + userID + `"
  	},
  	"query": "` + query + `"
	}`
}
