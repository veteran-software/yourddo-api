package main

import (
	"encoding/xml"
	"fmt"
	"github.com/veteran-software/yourddo-api/shared/types"
	"io"
	"regexp"
	"strings"
)

func ParseDatacenterXML(data io.Reader) (*types.ArrayOfDatacenterStruct, error) {
	content, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	xmlStr := string(content)

	re := regexp.MustCompile(`<\?xml[^>]+\?>`)
	xmlStr = re.ReplaceAllString(xmlStr, "")

	xmlStr = strings.ReplaceAll(xmlStr, "\r\n", "")
	xmlStr = strings.TrimSpace(xmlStr)

	reader := strings.NewReader(xmlStr)

	var result types.ArrayOfDatacenterStruct
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding XML: %w", err)
	}

	return &result, nil
}

func ParseStatusXML(data io.Reader) (*types.Status, error) {
	content, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	xmlStr := string(content)

	re := regexp.MustCompile(`<\?xml[^>]+\?>`)
	xmlStr = re.ReplaceAllString(xmlStr, "")

	// Clean up whitespace and line endings
	xmlStr = strings.ReplaceAll(xmlStr, "\r\n", "")
	xmlStr = strings.ReplaceAll(xmlStr, "\n", "")
	xmlStr = strings.ReplaceAll(xmlStr, "\r", "")
	xmlStr = strings.TrimSpace(xmlStr)

	reader := strings.NewReader(xmlStr)

	var result types.Status
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding XML: %w", err)
	}
	return &result, nil
}
