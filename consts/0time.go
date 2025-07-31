package consts

import "time"

// Datetime is the time in `yyyy-mm-dd hh:mm:ss` format
type Datetime string

func (t Datetime) Time() (time.Time, error) {
	ti, err := time.ParseInLocation(DateToHMSLayout, string(t), time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return ti, nil
}

const (
	ShortDateLayout  = "2006-01-02"
	DateToHMSLayout  = "2006-01-02 15:04:05"
	DateToHMSLayout2 = "20060102_150405"
	DateToHMSLayout3 = "150405"
)
