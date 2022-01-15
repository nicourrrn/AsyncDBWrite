package main

import (
	"encoding/json"
	"os"
)

type Product struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Image       string   `json:"image"`
	Type        string   `json:"type"`
	Ingredients []string `json:"ingredients"`
}

type Shop struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Image        string `json:"image"`
	WorkingHours struct {
		Opening string `json:"opening"`
		Closing string `json:"closing"`
	} `json:"workingHours"`
	Menu []Product `json:"menu"`
}

func NewFromJson(filename string) (Shop, error) {
	open, err := os.Open(filename)
	if err != nil {
		return Shop{}, err
	}
	var shop Shop
	err = json.NewDecoder(open).Decode(&shop)
	if err != nil {
		return Shop{}, err
	}
	return shop, nil
}
