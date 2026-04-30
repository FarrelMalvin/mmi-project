package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type SignatureConfig struct {
	Path   string  
	Offset string  
	Scale  float64 
	Page   string  
}

func GenerateFormPDF(templatePath string, formData map[string]string, signatures []SignatureConfig, w io.Writer) error{

	pdfBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("gagal membaca template: %v", err)
	}

	fields, err := api.FormFields(bytes.NewReader(pdfBytes), nil)
	if err != nil {
		return fmt.Errorf("gagal membaca form fields: %v", err)
	}

	textFields := []map[string]interface{}{}
	matchCount := 0
	for _, field := range fields {
		if val, exists := formData[field.Name]; exists {
			textFields = append(textFields, map[string]interface{}{
				"id":    field.ID,
				"name":  field.Name,
				"value": val,
			})
			matchCount++
		}
	}

	if matchCount == 0 {
		return fmt.Errorf("pdfcpu: missing form data - tidak ada Nama Field yang cocok")
	}

	finalJSON := map[string]interface{}{
		"forms": []map[string]interface{}{
			{"textfield": textFields},
		},
	}
	jsonData, _ := json.Marshal(finalJSON)
	var filledBuffer bytes.Buffer 
	conf := model.NewDefaultConfiguration()

	err = api.FillForm(bytes.NewReader(pdfBytes), bytes.NewReader(jsonData), &filledBuffer, conf)
	if err != nil {
		return fmt.Errorf("api.FillForm error: %v", err)
	}

	var lockedBuffer bytes.Buffer
	err = api.LockFormFields(bytes.NewReader(filledBuffer.Bytes()), &lockedBuffer, nil, conf)
	if err != nil {
		return fmt.Errorf("gagal mengunci form: %v", err)
	}

	currentPDF := lockedBuffer.Bytes()

    if len(signatures) == 0 {
        _, err = w.Write(currentPDF)
        return err
    }

 
    for _, sig := range signatures {
        if sig.Path == "" {
            continue
        }

        wmConf := fmt.Sprintf("scale: %.2f, pos: bl, offset: %s, rot: 0", sig.Scale, sig.Offset)
        
        wm, err := api.ImageWatermark(sig.Path, wmConf, true, false, types.POINTS)
        if err != nil {
            return fmt.Errorf("gagal konfigurasi tanda tangan %s: %v", sig.Path, err)
        }

        var tempBuffer bytes.Buffer
        
        pages := []string{sig.Page}
        if sig.Page == "" {
            pages = []string{"1"}
        }

        err = api.AddWatermarks(bytes.NewReader(currentPDF), &tempBuffer, pages, wm, conf)
        if err != nil {
            return fmt.Errorf("gagal menempelkan tanda tangan %s: %v", sig.Path, err)
        }
        currentPDF = tempBuffer.Bytes()
    }

    _, err = w.Write(currentPDF)
	return nil
}