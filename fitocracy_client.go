package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/headzoo/surf.v1"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"
	"log"
)

const (
	fitocracy_url = "https://www.fitocracy.com/"
)

type ApiActivity struct {
	Id    int    `json:"id"`
	Count int    `json:"count"`
	Name  string `json:"name"`
}

type ApiEffort struct {
	Id   int    `json:"id"`
	Abbr string `json:"abbr"`
	Name string `json:"name"`
}

type ApiAction struct {
	Id               int               `json:"id"`
	ActionTimeString string            `json:"actiontime"`
	ActionDateString string            `json:"actiondate"`
	ActionGroupId    int               `json:"action_group_id"`
	Effort0          float32           `json:"effort0"`
	Effort1          float32           `json:"effort1"`
	Effort2          float32           `json:"effort2"`
	Effort3          float32           `json:"effort3"`
	Effort0Unit      *ApiEffort        `json:"effort0_unit"`
	Effort1Unit      *ApiEffort        `json:"effort1_unit"`
	Effort2Unit      *ApiEffort        `json:"effort2_unit"`
	Effort3Unit      *ApiEffort        `json:"effort3_unit"`
	Activity         ApiActionActivity `json:"action"`
}

type ApiActionActivity struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ApiActivityHistory struct {
	Id                 int         `json:"id"`
	GroupId            int         `json:"action_group_id"`
	Points             int         `json:"points"`
	Name               string      `json:"name"`
	TimeString         string      `json:"time"`
	OriginalTimeString string      `json:"original_time"`
	Actions            []ApiAction `json:"actions"`
}

func (a ApiAction) Units() string {
	if a.Effort0Unit != nil {
		return a.Effort0Unit.Abbr
	}
	if a.Effort1Unit != nil {
		return a.Effort1Unit.Abbr
	}
	return ""
}

func (a ApiAction) PerformedAt() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05", a.ActionTimeString)
}

func activities_url(user_id int) string {
	return fmt.Sprintf("%sget_user_activities/%d/", fitocracy_url, user_id)
}

func activity_history_url(activity_id int) string {
	return fmt.Sprintf("%sget_history_json_from_activity/%d/?max_sets=-1&max_workouts=-1&reverse=1", fitocracy_url, activity_id)
}

/**
This function gets you a logged in, ready-to use http client in addition to returning the
user's fitocracy id (useful for future calls)
*/
func getClient(username string, password string) (err error, userId int, httpClient http.Client) {
	//Remove this when you're not testing through a proxy
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	bow := surf.NewBrowser()
	err = bow.Open(fitocracy_url)
	if err != nil {
		return
	}

	fm, err := bow.Form("form#login-modal-form ")
	if err != nil {
		return
	}

	fm.Input("username", username)
	fm.Input("password", password)
	fm.Dom().SetAttr("action", "/accounts/login/")
	fm.Dom().SetAttr("method", "post")

	if fm.Submit() != nil {
		panic(err)
	}

	userId, _ = strconv.Atoi(bow.ResponseHeaders().Get("X-Fitocracy-User"))
	cookies := bow.SiteCookies()

	//from this point forward we use the regular http library
	u, _ := url.Parse(fitocracy_url)
	cookieJar, _ := cookiejar.New(nil)
	cookieJar.SetCookies(u, cookies)

	//Long timeout because some of these API calls are slow
	httpClient = http.Client{Jar: cookieJar, Timeout: time.Second * 30}
	return
}

func getActivities(client http.Client, fitocracyUserId int) (err error, activities []ApiActivity) {
	log.Printf("Getting activities for user: %d\n", fitocracyUserId)
	resp, err := client.Get(activities_url(fitocracyUserId))
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal(body, &activities)
	return
}

func getActivityHistory(client http.Client, fitocracyActivityId int) (err error, activityHistories []ApiActivityHistory) {
	if 0 == fitocracyActivityId {
		return fmt.Errorf("Invalid Activity Id passed."), activityHistories
	}

	resp, err := client.Get(activity_history_url(fitocracyActivityId))
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Printf("Unmarshalling activityHistories for: %d\n", fitocracyActivityId)
	json.Unmarshal(body, &activityHistories)
	return
}
