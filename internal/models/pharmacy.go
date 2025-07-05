package models

type Pharmacy struct {
	IDGlobal string `json:"id_global"` // uuid
	Name     string `json:"name"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	INN      string `json:"inn"`
}
