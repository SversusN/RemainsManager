package models

import (
	"encoding/xml"
	"time"
)

type InterfirmRow struct {
	IDContractorGlobalTo    string     `json:"id_contractor_global_to"`
	IDInterfirmMoving       int64      `json:"id_interfirm_moving"`
	IDInterfirmMovingGlobal string     `json:"id_interfirm_moving_global"`
	Mnemocode               *string    `json:"mnemocode"`
	IDStoreFromMain         int64      `json:"id_store_from_main"`
	IDStoreFromTransit      int64      `json:"id_store_from_transit"`
	IDContractorTo          int64      `json:"id_contractor_to"`
	IDStoreToMain           int64      `json:"id_store_to_main"`
	IDStoreToTransit        int64      `json:"id_store_to_transit"`
	Date                    time.Time  `json:"date"`
	DocumentState           string     `json:"document_state"`
	Comment                 string     `json:"comment"`
	IDUser                  int64      `json:"id_user"`
	IDUser2                 *int64     `json:"id_user2"`
	SumSupplierHeader       float64    `json:"sum_supplier_header"`
	SVatSupplierHeader      float64    `json:"svat_supplier_header"`
	SumRetailHeader         float64    `json:"sum_retail_header"`
	SVatRetailHeader        float64    `json:"svat_retail_header"`
	GoodsSent               int        `json:"goods_sent"`
	AuthNum                 int        `json:"auth_num"`
	AuthValidPeriod         *time.Time `json:"auth_valid_period"`

	// Позиция
	IDInterfirmMovingItem       int64   `json:"id_interfirm_moving_item"`
	IDInterfirmMovingItemGlobal string  `json:"id_interfirm_moving_item_global"`
	Quantity                    float64 `json:"quantity"`
	IDLotFrom                   int64   `json:"id_lot_from"`
	IDLotTo                     int64   `json:"id_lot_to"`
	SumSupplier                 float64 `json:"sum_supplier"`
	SVatSupplier                float64 `json:"svat_supplier"`
	PVatRetail                  float64 `json:"pvat_retail"`
	VatRetail                   float64 `json:"vat_retail"`
	IsWeight                    int     `json:"is_weight"`
	Kiz                         *string `json:"kiz"`    // всегда NULL (по твоему SQL)
	IsKiz                       int     `json:"is_kiz"` // 1 или 0
}

type InterfirmItemXML struct {
	XMLName                     xml.Name `xml:"ITEM"`
	IDInterfirmMovingItem       int64    `xml:"ID_INTERFIRM_MOVING_ITEM"`
	IDInterfirmMovingItemGlobal string   `xml:"ID_INTERFIRM_MOVING_ITEM_GLOBAL"`
	Quantity                    float64  `xml:"QUANTITY"`
	IDLotFrom                   int64    `xml:"ID_LOT_FROM"`
	IDLotTo                     int64    `xml:"ID_LOT_TO"`
	SumSupplier                 float64  `xml:"SUM_SUPPLIER"`
	SVatSupplier                float64  `xml:"SVAT_SUPPLIER"`
	PVatRetail                  float64  `xml:"PVAT_RETAIL"`
	VatRetail                   float64  `xml:"VAT_RETAIL"`
	IsWeight                    int      `xml:"IS_WEIGHT"`
	Kiz                         *string  `xml:"KIZ"`    // может быть пусто
	IsKiz                       int      `xml:"IS_KIZ"` // обязательно: 1 или 0
}

type InterfirmMovingXML struct {
	XMLName                 xml.Name           `xml:"INTERFIRM_MOVING"`
	IDInterfirmMoving       int64              `xml:"ID_INTERFIRM_MOVING"`
	IDInterfirmMovingGlobal string             `xml:"ID_INTERFIRM_MOVING_GLOBAL"`
	Mnemocode               *string            `xml:"MNEMOCODE"`
	IDStoreFromMain         int64              `xml:"ID_STORE_FROM_MAIN"`
	IDStoreFromTransit      int64              `xml:"ID_STORE_FROM_TRANSIT"`
	IDContractorTo          int64              `xml:"ID_CONTRACTOR_TO"`
	IDStoreToMain           int64              `xml:"ID_STORE_TO_MAIN"`
	IDStoreToTransit        int64              `xml:"ID_STORE_TO_TRANSIT"`
	Date                    string             `xml:"DATE"`
	DocumentState           string             `xml:"DOCUMENT_STATE"`
	Comment                 string             `xml:"COMMENT"`
	IDUser                  int64              `xml:"ID_USER"`
	IDUser2                 *int64             `xml:"ID_USER2"`
	SumSupplier             float64            `xml:"SUM_SUPPLIER"`
	SVatSupplier            float64            `xml:"SVAT_SUPPLIER"`
	SumRetail               float64            `xml:"SUM_RETAIL"`
	SVatRetail              float64            `xml:"SVAT_RETAIL"`
	GoodsSent               int                `xml:"GOODS_SENT"`
	AuthNum                 int                `xml:"AUTH_NUM"`
	AuthValidPeriod         *string            `xml:"AUTH_VALID_PERIOD"`
	Items                   []InterfirmItemXML `xml:"ITEM"`
}

type InterfirmXML struct {
	XMLName xml.Name           `xml:"XML"`
	Moving  InterfirmMovingXML `xml:"INTERFIRM_MOVING"`
}
