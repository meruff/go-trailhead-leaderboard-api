package trailhead

import (
	"strconv"
)

type ProfileReturn struct {
	Error           string
	ProfilePhotoUrl string
	ProfileUser     struct {
		TBID_Role     string
		CompanyName   string
		TrailblazerId string
		Title         string
		FirstName     string
		LastName      string
		Id            string
	}
}

// Profile represents basic trailhead data i.e. name, title, company
type Profile struct {
	ID                       string `json:"id"`
	FirstName                string `json:"firstName"`
	LastName                 string `json:"lastName"`
	Username                 string `json:"username"`
	ProfileURL               string `json:"profileUrl"`
	BackgroundImageURL       string `json:"backgroundImageUrl"`
	IsPublicProfile          bool   `json:"isPublicProfile"`
	Role                     string `json:"role"`
	Title                    string `json:"title"`
	RelationshipToSalesforce string `json:"relationshipToSalesforce"`
	Nickname                 string `json:"nickname"`
	PhotoURL                 string `json:"photoUrl"`
	Bio                      string `json:"bio"`
	LinkedinHandle           string `json:"linkedinHandle"`
	WebsiteURL               string `json:"websiteUrl"`
	Company                  struct {
		Name    string `json:"name"`
		Size    string `json:"size"`
		Website string `json:"website"`
	} `json:"company"`
	Address struct {
		State   string `json:"state"`
		Country string `json:"country"`
	} `json:"address"`
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
	Data struct {
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
	} `json:"data"`
}

// Certifications represents certification records returned from trailhead.
type Certifications struct {
	Data struct {
		Profile struct {
			Typename   string `json:"__typename"`
			ID         string `json:"id"`
			Credential struct {
				Messages []struct {
					Typename string `json:"__typename"`
					Body     string `json:"body"`
					Header   string `json:"header"`
					Location string `json:"location"`
					Image    string `json:"image"`
					Cta      struct {
						Typename string `json:"__typename"`
						Label    string `json:"label"`
						URL      string `json:"url"`
					} `json:"cta"`
					Orientation string `json:"orientation"`
				} `json:"messages"`
				MessagesOnly bool `json:"messagesOnly"`
				Brands       []struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Logo     string `json:"logo"`
				} `json:"brands"`
				Certifications []struct {
					Cta struct {
						Typename string `json:"__typename"`
						Label    string `json:"label"`
						URL      string `json:"url"`
					} `json:"cta"`
					DateCompleted      string `json:"dateCompleted"`
					DateExpired        any    `json:"dateExpired"`
					DownloadLogoURL    string `json:"downloadLogoUrl"`
					LogoURL            string `json:"logoUrl"`
					InfoURL            string `json:"infoUrl"`
					MaintenanceDueDate string `json:"maintenanceDueDate"`
					Product            string `json:"product"`
					PublicDescription  string `json:"publicDescription"`
					Status             struct {
						Typename string `json:"__typename"`
						Title    string `json:"title"`
						Expired  bool   `json:"expired"`
						Date     string `json:"date"`
						Color    string `json:"color"`
						Order    int    `json:"order"`
					} `json:"status"`
					Title string `json:"title"`
				} `json:"certifications"`
			} `json:"credential"`
		} `json:"profile"`
	} `json:"data"`
}

// Badges represents skill data returned from trailhead.
type Badges struct {
	Data struct {
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
	} `json:"data"`
}

// BadgeRequest represents a request to the /badges endpoint. The variables to send to graphql
type BadgeRequest struct {
	Filter string `json:"filter"`
	After  string `json:"after"`
	Count  int    `json:"count"`
}

// GetGraphqlPayload returns a JSON string to use in Trailhead graphql callouts.
func GetGraphqlPayload(operationName string, userID string, variables string, query string) string {
	var variablesJsonString string

	if variables != "" {
		variablesJsonString = variables
	} else {
		variablesJsonString = `"variables": {
    		"hasSlug": true,
    		"slug": "` + userID + `"
  		}`
	}

	return `{
		"operationName": "` + operationName + `",
  		` + variablesJsonString + `,
  		"query": "` + query + `"
	}`
}

func GetBadgesFilterPayload(userID string, badgeFilters BadgeRequest) string {
	var afterLine, filterLine string

	if badgeFilters.After != "" {
		afterLine = `"after": "` + badgeFilters.After + `",`
	} else {
		afterLine = `"after": null,`
	}

	if badgeFilters.Filter != "" {
		filterLine = `"filter": "` + badgeFilters.Filter + `",`
	} else {
		filterLine = `"filter": null,`
	}

	return `"variables": {
		"count": ` + strconv.Itoa(badgeFilters.Count) + `,
		` + afterLine + `
		` + filterLine + `
		"hasSlug": true,
		"slug": "` + userID + `"
	}`
}
