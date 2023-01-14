package twapi_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/harnyk/teamjerk/internal/twapi"
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
			TimeRange: twapi.TimeRange{
				StartEpoch: 1667260800000,
				EndEpoch:   1669766400000,
			},
			Billable: []twapi.TimeChartEntry{
				{
					Epoch: 1669680000000,
					Hours: 8.00,
					Min:   480,
				},
				{
					Epoch: 1669766400000,
					Hours: 8.00,
					Min:   480,
				},
			},
			NonBillable: []twapi.TimeChartEntry{
				{
					Epoch: 1669680000000,
					Hours: 0.00,
					Min:   0,
				},
				{
					Epoch: 1669766400000,
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

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
