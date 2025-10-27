package models

type OfferItemReport struct {
	MnemoCode      string  `db:"mnemocode"`
	GoodsName      string  `db:"goods_name"`
	Qty            int     `db:"qty"`
	ContractorTo   string  `db:"contrator_to"`
	PriceSal       float64 `db:"price_sal"`
	LotName        string  `db:"lot_name"`
	ContractorFrom string  `db:"contractor_from"`
}

type OfferHeader struct {
	MnemoCode      string
	ContractorFrom string
}

type GroupedReport struct {
	Header OfferHeader
	Groups map[string][]OfferItemReport // ключ — ContractorTo
}
