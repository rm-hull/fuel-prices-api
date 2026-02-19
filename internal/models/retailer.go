package models

type Retailer struct {
	Name    string
	Url     string
	Favicon *string
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
