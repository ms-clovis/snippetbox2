package models

import (
	"time"
)

type Snippet struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	//NullABLEFoo sql.NullString
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"`
	Author  string    `json:"author"`
}

func (s *Snippet) IsModel() bool {
	return true
}

func NewEmptySnippet() *Snippet {
	return &Snippet{}
}
