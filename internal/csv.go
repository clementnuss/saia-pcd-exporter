package internal

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/spkg/bom"
)

// Convert the CSV string as internal date
func (rt *RegisterType) UnmarshalCSV(csv string) (err error) {
	switch csv {
	case "R":
		*rt = Register
	case "R Float":
		*rt = RegisterFloat
	case "Flag":
		*rt = Flag
	case "Input":
		*rt = Input
	case "Output":
		*rt = Output
	default:
		return fmt.Errorf("unkown register type: %s", csv)
	}
	return nil
}

func ParseCSV(csvPath string) ([]Metric, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	metrics := []Metric{}
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(bom.NewReader(in))
		r.Comma = ','
		return r
	})

	err = gocsv.UnmarshalFile(file, &metrics)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}
