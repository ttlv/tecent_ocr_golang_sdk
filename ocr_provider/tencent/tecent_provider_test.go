package tecent

import (
	"testing"

	"github.com/gopherSteven/tecent_ocr_golang_sdk/ocr_provider"
	"github.com/theplant/testingutils"
)

func TestTecentProvider(t *testing.T) {
	var (
		imageUrl  = "http://yyb.gtimg.com/aiplat/static/ai-demo/large/odemo-pic-1.jpg"
		appKey    = "6WkPfV0DQtwWifGN" // you can apply ypur app_key in ai.qq.com
		appID     = "1106697055"       //  you can apply your app_id in ai.qq.com
		processor = New(appKey, appID)
	)
	//request normal
	ocrData := ocr_provider.OcrCommonData{
		Name:      "李明",
		Number:    "440524198701010014",
		ValidDate: "",
	}
	ocrResponse, _ := processor.OcrCheck(imageUrl, false)
	if ocrData != ocrResponse {
		t.Errorf("api request error")
	}
	//abnormal
	_, err := processor.OcrCheck("https://kyc1process-1253546493.cos.ap-shanghai.myqcloud.com/timg.jpeg", true)
	if diff := testingutils.PrettyJsonDiff("返回码描述: 输入图片不是身份证, 建议处理方式: 请检查图片是否为身份证", err.Error()); len(diff) > 0 {
		t.Errorf("tecent provider post error: %v", diff)
	}
}
