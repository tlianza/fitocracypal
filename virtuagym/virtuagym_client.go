package virtuagym

import (
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

const (
	virtuagym_url = "https://virtuagym.com/"
)

type VirtuagymClient struct {
	httpClient http.Client
	username   string
	password   string
	apikey     string
}

type VirtuagymApiActivity struct {
	ActivityInstanceId int    `json:"act_inst_id,omitempty"`
	ActivityId         int    `json:"act_id"`
	Timestamp          int    `json:"timestamp"`
	Reps               []int  `json:"reps"`
	Weights            []int  `json:"weights"`
	Order              int    `json:"order"`
	Done               int    `json:"done"`
	Deleted            int    `json:"deleted"`
	PersonalNote       string `json:"personal_note"`
	ExternalActivityId string `json:"external_activity_id"`
	ExternalOrigin     string `json:"external_origin"`
}

type VirtuagymApiResultResponse struct {
	StatusCode    int    `json:"statuscode"`
	StatusMessage string `json:"statusmessage"`
	ResultCount   int    `json:"result_count"`
	Timestamp     int    `json:"timestamp"`
}

type VirtuagymApiCreateActivityResponse struct {
	VirtuagymApiResultResponse
	Result VirtuagymApiCreateActivityResult `json:"result"`
}

type VirtuagymApiGetActivityResponse struct {
	VirtuagymApiResultResponse
	Result []VirtuagymApiActivity `json:"result"`
}

type VirtuagymApiCreateActivityResult struct {
	ActivityInstanceId int `json:"act_inst_id"`
}

func CreateClient(username string, password string, apikey string) *VirtuagymClient {
	return &VirtuagymClient{
		http.Client{},
		username,
		password,
		apikey,
	}
}

func (c VirtuagymClient) GetActivityInstance(activityInstanceId int) (err error, apiResponse VirtuagymApiGetActivityResponse) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v0/activity/%d?api_key=%s", virtuagym_url, activityInstanceId, c.apikey), nil)
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.httpClient.Do(req)
	if nil != err {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	apiResponse = VirtuagymApiGetActivityResponse{}
	err = json.Unmarshal(body, &apiResponse)
	return
}

//func (c VirtuagymClient) CreateActivity(activity VirtuagymApiActivity) (err error, response VirtuagymApiCreateActivityResponse) {
//
//}
