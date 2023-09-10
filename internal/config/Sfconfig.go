package config

// 顺丰配置
type SfConf struct {
	PartnerID   string `json:"PartnerID"`
	CheckCode   string `json:"CheckCode"`
	MonthlyCard string `json:"MonthlyCard"`
	SfUrl       string `json:"SfUrl"`
}
