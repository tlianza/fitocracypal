package fitocracy

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)


func TestParseActivityHistory(t *testing.T) {
	var activityHistories []ApiActivityHistory
	dat, err := ioutil.ReadFile("../test_assets/sample_activity_history.json")
	if nil != err {
		t.Fatal(err)
	}
	err = json.Unmarshal(dat, &activityHistories)
	if nil != err {
		t.Fatal(err)
	}

	firstActivity := activityHistories[0]
	assert.Equal(t, 45255911, firstActivity.Id)
	assert.Equal(t, "Workout A", firstActivity.Name)
	assert.Equal(t, 898, firstActivity.Points)
	assert.Equal(t, "2016-04-28T14:36:57", firstActivity.TimeString)
	assert.Equal(t, "2016-04-28T15:27:42", firstActivity.OriginalTimeString)

	//each action
	assert.Equal(t, 336990561, firstActivity.Actions[0].Id)
	assert.Equal(t, float32(35), firstActivity.Actions[0].Effort1)
	assert.Equal(t, 31, firstActivity.Actions[0].Effort1Unit.Id) //31 = reps
	assert.Equal(t, "2016-04-28T14:36:57", firstActivity.Actions[0].ActionTimeString)

	assert.Equal(t, 336990562, firstActivity.Actions[1].Id)
	assert.Equal(t, float32(30), firstActivity.Actions[1].Effort1)
	assert.Equal(t, 31, firstActivity.Actions[1].Effort1Unit.Id) //31 = reps

	//group id's match the above activity entry
	assert.Equal(t, 45255911, firstActivity.Actions[0].ActionGroupId)
	assert.Equal(t, 45255911, firstActivity.Actions[1].ActionGroupId)

	//action id's should match the action id we keep in our db
	assert.Equal(t, 396, firstActivity.Actions[0].Activity.Id)
	assert.Equal(t, 396, firstActivity.Actions[1].Activity.Id)
}
