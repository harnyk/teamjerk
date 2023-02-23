package twapi

type LogtimeTimelog struct {
	TaskID  uint64 `json:"taskId"`
	Hours   uint64 `json:"hours"`
	Minutes uint64 `json:"minutes"`
	// Date is in the format YYYY-MM-DD
	Date string `json:"date"`
	// Time is in the format HH:MM:SS
	Time        string   `json:"time"`
	Description string   `json:"description"`
	IsBillable  bool     `json:"isBillable"`
	UserID      uint64   `json:"userId"`
	TagIDs      []uint64 `json:"tagIds"`
}

type LogtimeTimelogOptions struct {
	MarkTaskComplete bool
}

type LogtimeRequest struct {
	Timelog        LogtimeTimelog        `json:"timelog"`
	TimelogOptions LogtimeTimelogOptions `json:"timelogOptions"`
}

type LogtimeRequestWithProjectID struct {
	LogtimeRequest
	ProjectID uint64 `json:"projectId"`
}
