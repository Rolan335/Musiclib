package entity

import (
	"time"
)

// params for GET /songs
type GetSongsParams struct {
	Group    *string    `form:"group,omitempty" json:"group,omitempty"`
	Title    *string    `form:"title,omitempty" json:"title,omitempty"`
	Text     *string    `form:"lyrics,omitempty" json:"lyrics,omitempty"`
	DateFrom *time.Time `form:"dateFrom,omitempty" json:"date_from,omitempty"`
	DateTo   *time.Time `form:"dateTo,omitempty" json:"date_to,omitempty"`
	Page     *int       `form:"page,omitempty" json:"page,omitempty"`
	PageSize *int       `form:"pageSize,omitempty" json:"page_size,omitempty"`
}

// Song represents full information about the song
type Song struct {
	ID          int       `json:"id,omitempty"`
	Group       string    `json:"group,omitempty"`
	Title       string    `json:"title,omitempty"`
	ReleaseDate time.Time `json:"releaseDate,omitempty"`
	Text        string    `json:"text,omitempty"`
	Link        string    `json:"link,omitempty"`
}

type SongNullable struct {
	ID          *int       `json:"id,omitempty"`
	Group       *string    `json:"group,omitempty"`
	Title       *string    `json:"title,omitempty"`
	ReleaseDate *time.Time `json:"releaseDate,omitempty"`
	Text        *string    `json:"text,omitempty"`
	Link        *string    `json:"link,omitempty"`
}

// Text represents text of the song
// Return text like slice (verse) of slice (string) of strings
type Text struct {
	Text [][]string `json:"text,omitempty"`
}
