package models

type Agency struct {
	AgencyID   string `db:"agency_id" json:"id"`
	AgencyName string `db:"agency_name" json:"name"`
	AgencyURL  string `db:"agency_url" json:"url"`
}
