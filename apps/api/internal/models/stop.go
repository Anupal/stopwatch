package models

type Stop struct {
	StopID   string  `db:"stop_id" json:"id"`
	StopCode string  `db:"stop_code" json:"code"`
	StopName string  `db:"stop_name" json:"name"`
	StopLat  float64 `db:"stop_lat" json:"latitude"`
	StopLon  float64 `db:"stop_lon" json:"longitude"`
}
