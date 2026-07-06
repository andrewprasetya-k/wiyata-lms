package service

import (
	"context"
	"io"
	"os"

	"github.com/ledongthuc/pdf"
)

type ledongthucPDFTextExtractor struct{}

func NewPDFTextExtractor() PDFTextExtractor {
	return ledongthucPDFTextExtractor{}
}

func (ledongthucPDFTextExtractor) ExtractText(ctx context.Context, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp("", "wiyata-material-summary-*.pdf")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	file, reader, err := pdf.Open(tmpName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	limited := io.LimitReader(textReader, int64(materialSummaryMaxExtractedText*8))
	text, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return string(text), nil
}
