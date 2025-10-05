package models

import (
	"time"
)

// OfferItem представляет товар в заявке
type OfferItem struct {
	ID                     int64  `json:"id,omitempty"`
	OfferID                int64  `json:"offer_id"`
	IdContractorGlobalFrom string `json:"id_contractor_global_from"`
	IdContractorGlobalTo   string `json:"id_contractor_global_to"`
	GoodsId                string `json:"goods_id"`
	Quantity               int    `json:"quantity"`
}

// Offer представляет заявку
type Offer struct {
	ID                     int64       `json:"id,omitempty"`
	Name                   string      `json:"name"`
	IdContractorGlobalFrom string      `json:"id_contractor_global_from"`
	CreatedAt              time.Time   `json:"created_at"`
	OfferItems             []OfferItem `json:"items"`
}

type OfferJournalItem struct {
	ID         int64  `json:"id"`
	Mnemocode  string `json:"mnemocode"`  // имя заявки
	Contractor string `json:"contractor"` // контрагент-отправитель
	CreatedAt  string `json:"created"`    // дата создания (YYYY-MM-DD)
}

// OfferDetailItem — детализация одной позиции заявки
type OfferDetailItem struct {
	GoodsName    string `json:"goods_name"`    // "Название | Производитель"
	ContractorTo string `json:"contractor_to"` // контрагент-получатель
	Quantity     int    `json:"quantity"`
	ID           int    `json:"id_item"`
}
