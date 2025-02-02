package controller

import (
	"strings"
	"time"
)

// CustomTime is a custom time type for unmarshalling json
var timeLayout = "02.01.2006"

type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	parsedTime, err := time.Parse(timeLayout, s)
	if err != nil {
		return err
	}
	*ct = CustomTime(parsedTime)
	return nil
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(ct).Format(timeLayout)), nil
}

func (ct CustomTime) Time() time.Time {
	return time.Time(ct)
}

type Song struct {
	Group       string     `json:"group,omitempty"`
	Title       string     `json:"title,omitempty"`
	ReleaseDate CustomTime `json:"releaseDate,omitempty"`
	Text        string     `json:"text,omitempty"`
	Link        string     `json:"link,omitempty"`
}

type SongNullable struct {
	Group       *string     `json:"group,omitempty"`
	Title       *string     `json:"title,omitempty"`
	ReleaseDate *CustomTime `json:"releaseDate,omitempty"`
	Text        *string     `json:"text,omitempty"`
	Link        *string     `json:"link,omitempty"`
}
