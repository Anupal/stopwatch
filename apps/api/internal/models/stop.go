package models

type Stop struct {
	StopID   string   `json:"id"`
	StopCode string   `json:"code"`
	StopName string   `json:"name"`
	StopLat  float64  `json:"latitude"`
	StopLon  float64  `json:"longitude"`
	Agencies []Agency `json:"agencies"`
	Distance float64  `json:"distance_meters,omitempty"`
}

type StopAgencyMV struct {
	StopID     string  `db:"stop_id"`
	StopCode   string  `db:"stop_code"`
	StopName   string  `db:"stop_name"`
	StopLat    float64 `db:"stop_lat"`
	StopLon    float64 `db:"stop_lon"`
	AgencyID   string  `db:"agency_id"`
	AgencyName string  `db:"agency_name"`
	AgencyURL  string  `db:"agency_url"`
	Distance   float64 `db:"distance_meters"`
}

type GetNearestStopsParams struct {
	AgencyIDs       []string
	Latitude        string
	Longitude       string
	MaximumDistance string
}
