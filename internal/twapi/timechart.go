package twapi

import (
	"encoding/json"
	"strconv"
	"time"
)

type TimeChartResponse struct {
	Status string    `json:"STATUS"`
	User   TimeChart `json:"user"`
}

type TimeX time.Time

func (t *TimeX) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}

	*t = TimeX(time.Unix(i/1000, 0))

	return nil
}

type TimeChart struct {
	Billable    []TimeChartEntry `json:"billable"`
	NonBillable []TimeChartEntry `json:"nonbillable"`

	StartEpoch TimeX `json:"startepoch"`
	EndEpoch   TimeX `json:"endepoch"`

	Id        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type TimeChartEntry struct {
	Epoch time.Time
	Hours float64
	Min   uint64
}

func (t *TimeChartEntry) UnmarshalJSON(b []byte) error {
	var s []string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	millis, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return err
	}
	t.Epoch = time.Unix(millis/1000, 0)

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
