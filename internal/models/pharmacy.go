package models

type Pharmacy struct {
	ID_CONTRACTOR_GLOBAL string `json:"id_global"` // uuid
	Name                 string `json:"name"`
	Address              string `json:"address"`
	Phone                string `json:"phone"`
	INN                  string `json:"inn"`
}
