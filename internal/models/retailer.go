package models

import "strings"

type Retailer struct {
	Name    string  `json:"name"`
	Url     string  `json:"url"`
	Favicon *string `json:"favicon,omitempty"`
}

func (org *Retailer) ToCSV() []string {
	row := []string{
		org.Name,
		org.Url,
		"",
	}
	if org.Favicon != nil {
		row[2] = *org.Favicon
	}

	return row
}

func FromCSV(record, headers []string) (*Retailer, error) {
	retailer := &Retailer{
		Name: record[0],
		Url:  record[1],
	}
	if len(record) == 3 && record[2] != "" {
		retailer.Favicon = &record[2]
	}
	return retailer, nil
}

type Retailers map[string]*Retailer

func (r *Retailers) MatchBrandName(name string) *Retailer {
	for word := range strings.FieldsSeq(strings.ToUpper(name)) {
		if retailer, exists := (*r)[word]; exists {
			return retailer
		}
	}
	return nil
}
