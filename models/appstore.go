package models

import "time"

type Links struct {
	Self    string `json:"self"`
	Related string `json:"related"`
}

type Meta struct {
	Paging struct {
		Limit int `json:"limit"`
	} `json:"paging"`
}

type AppInfo struct {
	Data struct {
		Relationships struct {
			CiProduct struct {
				Links Links `json:"links"`
			} `json:"ciProduct"`
		} `json:"relationships"`
	} `json:"data"`
}

type CiProduct struct {
	Data struct {
		Relationships struct {
			BuildRuns struct {
				Links Links `json:"links"`
			} `json:"buildRuns"`
		} `json:"relationships"`
	} `json:"data"`
}

type BuildRuns struct {
	Data []struct {
		Attributes struct {
			Number            int    `json:"number"`
			FinishedDate      string `json:"finishedDate"`
			ExecutionProgress string `json:"executionProgress"`
			CompletionStatus  string `json:"completionStatus"`
		} `json:"attributes"`
		Relationships struct {
			Builds struct {
				Links Links `json:"links"`
			} `json:"builds"`
			Actions struct {
				Links Links `json:"links"`
			} `json:"actions"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type Actions struct {
	Data []struct {
		Attributes struct {
			ActionType        string    `json:"actionType"`
			ExecutionProgress string    `json:"executionProgress"`
			Name              string    `json:"name"`
			CompletionStatus  string    `json:"completionStatus"`
			FinishedDate      time.Time `json:"finishedDate"`
		} `json:"attributes"`
		Relationships struct {
			Artifacts struct {
				Links Links `json:"links"`
			} `json:"artifacts"`
		} `json:"relationships"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type Artifacts struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			ActionType        string    `json:"actionType"`
			IssueCounts       any       `json:"issueCounts"`
			ExecutionProgress string    `json:"executionProgress"`
			Name              string    `json:"name"`
			StartedDate       time.Time `json:"startedDate"`
			CompletionStatus  string    `json:"completionStatus"`
			IsRequiredToPass  bool      `json:"isRequiredToPass"`
			FinishedDate      time.Time `json:"finishedDate"`
		} `json:"attributes"`
		Relationships struct {
			Artifacts struct {
				Links Links `json:"links"`
			} `json:"artifacts"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type CIArtifacts struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			FileName string `json:"fileName"`
			FileSize int    `json:"fileSize"`
			FileType string `json:"fileType"`
		} `json:"attributes"`
		Links Links `json:"links"`
	} `json:"data"`
	Meta Meta `json:"meta"`
}

type DownloadURL struct {
	Data struct {
		Attributes struct {
			FileType    string `json:"fileType"`
			FileName    string `json:"fileName"`
			FileSize    int    `json:"fileSize"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"attributes"`
	} `json:"data"`
}

type Apps struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Name     string `json:"name"`
			BundleID string `json:"bundleId"`
		} `json:"attributes"`
		Relationships struct {
			CiProduct struct {
				Links Links `json:"links"`
			} `json:"ciProduct"`
		} `json:"relationships"`
		Links Links `json:"links"`
	} `json:"data"`
}
