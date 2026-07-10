package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
)

// CSV parsing helpers for AdminSchoolMemberImportService. Split out of
// admin_school_member_import_service.go to keep the main file focused on
// orchestration; behavior is unchanged.

func (s *adminSchoolMemberImportService) parseCSV(content []byte) ([]normalizedImportRow, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("file CSV tidak bisa dibaca: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("file CSV kosong")
	}

	header := make(map[string]int)
	for idx, value := range records[0] {
		header[strings.ToLower(strings.TrimSpace(value))] = idx
	}

	requiredHeaders := []string{"fullname", "email", "role"}
	for _, name := range requiredHeaders {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("kolom %s wajib ada", name)
		}
	}

	rows := make([]normalizedImportRow, 0, len(records)-1)
	classCodeIndex, hasClassCode := header["classcode"]
	if !hasClassCode {
		classCodeIndex = -1
	}

	for index, record := range records[1:] {
		rowNumber := index + 2
		if isEmptyCSVRecord(record) {
			continue
		}
		rows = append(rows, normalizedImportRow{
			RowNumber: rowNumber,
			FullName:  csvValue(record, header["fullname"]),
			Email:     strings.ToLower(csvValue(record, header["email"])),
			Role:      strings.ToLower(csvValue(record, header["role"])),
			ClassCode: csvValue(record, classCodeIndex),
		})
	}

	if len(rows) == 0 {
		return nil, errors.New("file CSV tidak memiliki baris data")
	}
	return rows, nil
}

func csvValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func isEmptyCSVRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
