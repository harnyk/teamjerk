package app

import "time"

type LogOptions struct {
	DryRun      bool
	NonBillable bool
	ProjectID   uint64
	TaskID      uint64
	Date        time.Time
	StartTime   time.Time
	Duration    time.Duration
	Description string
}
