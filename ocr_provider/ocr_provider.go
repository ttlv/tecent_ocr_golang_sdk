package ocr_provider

type OcrProvider interface {
	OcrCheck(imageUrl string, isBack bool) (ocrCommonData OcrCommonData, err error)
}

type OcrCommonData struct {
	Name      string
	Number    string
	ValidDate string
}
