package model

type Session struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Creator      string `json:"creator"`
	BaseLocation string `json:"baseLocation"`
}
