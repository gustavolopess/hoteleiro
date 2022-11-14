package format

import (
	"time"
)

func DDMMYYYYstringToTimeObj(dateStr string) (time.Time, error) {
	var t time.Time
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return t, err
	}
	return t, nil
}
