package models

type Arrival struct {
	TripId         string `db:"trip_id"`
	ArrivalTime    string `db:"arrival_time"`
	DepartureTime  string `db:"departure_time"`
	RouteShortName string `db:"route_short_name"`
	RouteLongName  string `db:"route_long_name"`
	TripHeadsign   string `db:"trip_headsign"`
}

type GetArrivalsResponse struct {
	StopID        string    `json:"stop_id,omitempty"`
	Stop          *Stop     `json:"stop,omitempty"`
	MinutesAhead  int       `json:"minutes_ahead,omitempty"`
	MinutesBehind int       `json:"minutes_behind,omitempty"`
	Arrivals      []Arrival `json:"arrivals,omitempty"`
}

const (
	DefaultArrivalMinutesAhead  = 30
	DefaultArrivalMinutesBehind = 5
)

type GetStopArrivalsParams struct {
	StopID        string
	MinutesAhead  int
	MinutesBehind int
}

// Normalize to ensure parameters have valid defaults applied
func (p *GetStopArrivalsParams) NormalizeMinutes() {
	if p.MinutesAhead <= 0 {
		p.MinutesAhead = DefaultArrivalMinutesAhead
	}
	if p.MinutesBehind < 0 {
		p.MinutesBehind = DefaultArrivalMinutesBehind
	}
}
