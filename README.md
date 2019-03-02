## Install

Run `go get github.com/gopherSteven/tecent_ocr_golang_sdk`

## Usage

```go
package test

import (
	"testing"

	"github.com/gopherSteven/tecent_ocr_golang_sdk/ocr_provider"
	"github.com/theplant/testingutils"
)

func TestTecentProvider(t *testing.T) {
	var (
		imageUrl  = "http://yyb.gtimg.com/aiplat/static/ai-demo/large/odemo-pic-1.jpg"
		processor = New()
	)
	//请求正常的情况
	ocrData := ocr_provider.OcrCommonData{
		Name:      "李明",
		Number:    "440524198701010014",
		ValidDate: "",
	}
	ocrResponse, _ := processor.OcrCheck(imageUrl, false)
	if ocrData != ocrResponse {
		t.Errorf("api请求异常")
	}
	//图片异常
	_, err := processor.OcrCheck("https://kyc1process-1253546493.cos.ap-shanghai.myqcloud.com/timg.jpeg", true)
	if diff := testingutils.PrettyJsonDiff("返回码描述: 输入图片不是身份证, 建议处理方式: 请检查图片是否为身份证", err.Error()); len(diff) > 0 {
		t.Errorf("tecent provider post error: %v", diff)
	}
}

```