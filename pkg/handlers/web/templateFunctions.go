package web

import "time"

func DisplayDate(t time.Time) string {
	return t.Format("Mon Jan 2 2006")
}
