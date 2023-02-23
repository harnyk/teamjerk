package twapi_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/harnyk/teamjerk/internal/twapi"
	"github.com/ysmood/got"
)

func TestTimeChartResponse_UnmarshalJSON(t *testing.T) {
	payload := []byte(`
	{
		"STATUS": "OK",
		"user": {
			"billable": [
				[
					"1669680000000",
					"8.00",
					"480"
				],
				[
					"1669766400000",
					"8.00",
					"480"
				]
			],
			"lastname": "Smith",
			"firstname": "John",
			"nonbillable": [
				[
					"1669680000000",
					"0.00",
					"0"
				],
				[
					"1669766400000",
					"0.00",
					"0"
				]
			],
			"id": "123456",
			"endepoch": "1669766400000",
			"startepoch": "1667260800000"
		}
	}	
	`)

	expected := &twapi.TimeChartResponse{
		Status: "OK",
		User: twapi.TimeChart{
			StartEpoch: twapi.TimeX(time.Unix(1667260800000/1000, 0)),
			EndEpoch:   twapi.TimeX(time.Unix(1669766400000/1000, 0)),
			Billable: []twapi.TimeChartEntry{
				{
					Epoch: time.Unix(1669680000000/1000, 0),
					Hours: 8.00,
					Min:   480,
				},
				{
					Epoch: time.Unix(1669766400000/1000, 0),
					Hours: 8.00,
					Min:   480,
				},
			},
			NonBillable: []twapi.TimeChartEntry{
				{
					Epoch: time.Unix(1669680000000/1000, 0),
					Hours: 0.00,
					Min:   0,
				},
				{
					Epoch: time.Unix(1669766400000/1000, 0),
					Hours: 0.00,
					Min:   0,
				},
			},
			Id:        "123456",
			FirstName: "John",
			LastName:  "Smith",
		},
	}

	actual := &twapi.TimeChartResponse{}

	err := json.Unmarshal(payload, actual)

	got.T(t).Eq(err, nil)
	got.T(t).Eq(actual, expected)

	got.T(t).Eq(
		time.Time(actual.User.StartEpoch).UTC().Format(time.RFC3339),
		"2022-11-01T00:00:00Z",
	)
	got.T(t).Eq(
		time.Time(actual.User.EndEpoch).UTC().Format(time.RFC3339),
		"2022-11-30T00:00:00Z",
	)
}
