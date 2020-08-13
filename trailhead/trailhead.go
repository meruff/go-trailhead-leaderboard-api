package trailhead

import "strings"

// Data represent a response from trailhead.salesforce.com
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

// GetAuraContext returns a JSON string containing the Aura "context" to use in the callout to Trailhead.
func GetAuraContext(fwuid string) string {
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
