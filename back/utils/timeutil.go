package utils

import "time"

// JST は日本標準時（UTC+9）。
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

// WeekStartJST は指定時刻が属する週の月曜0:00(JST)を返す。
func WeekStartJST(t time.Time) time.Time {
	jt := t.In(JST)
	// Weekday: Sunday=0 ... Monday=1
	weekday := int(jt.Weekday())
	// 月曜を週初めとする（日曜は前週の月曜から6日後）
	diff := weekday - int(time.Monday)
	if diff < 0 {
		diff += 7
	}
	monday := time.Date(jt.Year(), jt.Month(), jt.Day(), 0, 0, 0, 0, JST).AddDate(0, 0, -diff)
	return monday
}

// CurrentWeekStartJST は今週の月曜0:00(JST)を返す。
func CurrentWeekStartJST() time.Time {
	return WeekStartJST(time.Now())
}
