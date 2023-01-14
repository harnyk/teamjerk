package twapi

import (
	"encoding/json"
	"strconv"
)

type TimeChartResponse struct {
	Status string    `json:"STATUS"`
	User   TimeChart `json:"user"`
}

type TimeChart struct {
	Billable    []TimeChartEntry `json:"billable"`
	NonBillable []TimeChartEntry `json:"nonbillable"`

	TimeRange

	Id        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type TimeChartEntry struct {
	Epoch uint64
	Hours float64
	Min   uint64
}

type TimeRange struct {
	StartEpoch uint64
	EndEpoch   uint64
}

func (t *TimeChartEntry) UnmarshalJSON(b []byte) error {
	var s []string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	t.Epoch, err = strconv.ParseUint(s[0], 10, 64)
	if err != nil {
		return err
	}

	t.Hours, err = strconv.ParseFloat(s[1], 64)
	if err != nil {
		return err
	}

	t.Min, err = strconv.ParseUint(s[2], 10, 64)
	if err != nil {
		return err
	}

	return nil
}

func (t *TimeRange) UnmarshalJSON(b []byte) error {
	type alias struct {
		StartEpoch string `json:"startepoch"`
		EndEpoch   string `json:"endepoch"`
	}

	var s alias
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	t.StartEpoch, err = strconv.ParseUint(s.StartEpoch, 10, 64)
	if err != nil {
		return err
	}

	t.EndEpoch, err = strconv.ParseUint(s.EndEpoch, 10, 64)
	if err != nil {
		return err
	}

	return nil
}
