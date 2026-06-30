package utils

import "time"

func JakartaLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Local
	}
	return location
}

func NowJakarta() time.Time {
	return time.Now().In(JakartaLocation())
}
