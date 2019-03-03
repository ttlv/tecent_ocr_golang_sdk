package tecent

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encoding/base64"

	"github.com/gopherSteven/tecent_ocr_golang_sdk/ocr_provider"
)

// 腾讯AI OCR SDK Struct

type Response struct {
	Ret  int    `json:"ret"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Name       string `json:"name"`
	Sex        string `json:"sex"`
	Nation     string `json:"nation"`
	Birth      string `json:"birth"`
	Address    string `json:"address"`
	Id         string `json:"id"`
	FrontImage string `json:"frontimage"`
	Authority  string `json:"authority"`
	ValidData  string `json:"valid_date"`
	BackImage  string `json:"backimage"`
}

var (
	RejectReasonMaps map[string]string
)

type TenCentOCRProvider struct {
	AppKey string
	AppID  string
}

func New(appKey, appID string) TenCentOCRProvider {
	return TenCentOCRProvider{
		AppKey: appKey,
		AppID:  appID,
	}
}

//create random string
func randString(n int) string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

//base64 encode image

func getImageAndBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(buffer) > 1048576 {
		return "", fmt.Errorf("图片资源大于1MB，无法调用API")
	}
	return base64.StdEncoding.EncodeToString(buffer), nil
}

// get request sign
func getReqSign(u url.Values, app_key string) (sign string) {
	byte := []byte(fmt.Sprintf("%s&%s=%s", u.Encode(), "app_key", app_key))
	//md5 operation
	newmd := md5.New()
	newmd.Write(byte)
	sign_byte := newmd.Sum(nil)
	sign = strings.ToUpper(fmt.Sprintf("%x", sign_byte))
	return sign
}

func httpPost(u url.Values) (respBody []byte, err error) {
	url := "https://api.ai.qq.com/fcgi-bin/ocr/ocr_idcardocr"
	contentType := "application/x-www-form-urlencoded"
	resp, err := http.Post(url, contentType, strings.NewReader(u.Encode()))
	if err != nil {
		// TODO handler err
		return respBody, err
	}
	respBody, _ = ioutil.ReadAll(resp.Body)
	return respBody, nil
}

func (tenCentOCRProvider TenCentOCRProvider) OcrCheck(imageUrl string, isBack bool) (ocrCommonData ocr_provider.OcrCommonData, err error) {
	//调用api,返回api的结构
	var (
		now          = time.Now().Unix()
		v            = url.Values{}
		formatBase64 string
	)
	v.Add("app_id", tenCentOCRProvider.AppID)
	v.Add("time_stamp", fmt.Sprintf("%d", now))
	v.Add("nonce_str", randString(32))
	formatBase64, err = getImageAndBase64(imageUrl)
	if err != nil {
		return ocrCommonData, err
	}
	v.Add("image", formatBase64)
	if isBack {
		v.Add("card_type", "1")
	} else {
		v.Add("card_type", "0")
	}
	sign := getReqSign(v, tenCentOCRProvider.AppKey)
	v.Add("sign", sign)
	respData := []byte{}
	if respData, err = httpPost(v); err != nil {
		return ocrCommonData, err
	}
	var response Response
	if err := json.Unmarshal(respData, &response); err != nil {
		return ocrCommonData, err
	}
	if response.Ret != 0 {
		return ocrCommonData, fmt.Errorf(getRejectReason(fmt.Sprintf("%v", response.Ret)))
	}

	ocrCommonData.ValidDate = response.Data.ValidData
	ocrCommonData.Number = response.Data.Id
	ocrCommonData.Name = response.Data.Name

	return ocrCommonData, nil
}

func getRejectReason(code string) string {
	RejectReasonMaps := make(map[string]string)
	RejectReasonMaps["9"] = "返回码描述: qps超过限制; 建议处理方式: 用户认证升级或者降低调用频率"
	RejectReasonMaps["4096"] = "返回码描述: 参数非法, 建议处理方式: 请检查请求参数是否符合要求"
	RejectReasonMaps["12289"] = "返回码描述: 应用不存在, 建议处理方式: 请检查app_id是否有效的应用标识（AppId）"
	RejectReasonMaps["12801"] = "返回码描述: 素材不存在, 建议处理方式: 请检查app_id对应的素材模版id"
	RejectReasonMaps["12802"] = "返回码描述: 素材ID与应用ID不匹配, 建议处理方式: 请检查app_id对应的素材模版id"
	RejectReasonMaps["16385"] = "返回码描述: 缺少app_id参数, 建议处理方式: 请检查请求中是否包含有效的app_id参数"
	RejectReasonMaps["16386"] = "返回码描述: 缺少time_stamp参数, 建议处理方式: 请检查请求中是否包含有效的time_stamp参数"
	RejectReasonMaps["16387"] = "返回码描述: 缺少nonce_str参数, 建议处理方式: 请检查请求中是否包含有效的nonce_str参数"
	RejectReasonMaps["16388"] = "返回码描述: 请求签名无效, 建议处理方式: 检查请求中的签名信息（sign）是否有效"
	RejectReasonMaps["16389"] = "返回码描述: 缺失API权限, 建议处理方式: 请检查应用是否勾选当前API所属接口的权限"
	RejectReasonMaps["16390"] = "返回码描述: time_stamp参数无效, 建议处理方式: 请检查time_stamp距离当前时间是否超过5分钟"
	RejectReasonMaps["16391"] = "返回码描述: 同义词识别结果为空, 建议处理方式: 请尝试更换文案"
	RejectReasonMaps["16392"] = "返回码描述: 专有名词识别结果为空, 建议处理方式: 请尝试更换文案"
	RejectReasonMaps["16393"] = "返回码描述: 意图识别结果为空, 建议处理方式: 请尝试更换文案"
	RejectReasonMaps["16394"] = "返回码描述: 闲聊返回结果为空, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16396"] = "返回码描述: 图片格式非法, 建议处理方式: 请检查图片格式是否符合API要求"
	RejectReasonMaps["16397"] = "返回码描述: 图片体积过大, 建议处理方式: 请检查图片大小是否超过API限制"
	RejectReasonMaps["16402"] = "返回码描述: 图片没有人脸, 建议处理方式: 请检查图片是否包含人脸"
	RejectReasonMaps["16403"] = "返回码描述: 相似度错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16404"] = "返回码描述: 人脸检测失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16405"] = "返回码描述: 图片解码失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16406"] = "返回码描述: 特征处理失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16407"] = "返回码描述: 提取轮廓错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16408"] = "返回码描述: 提取性别错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16409"] = "返回码描述: 提取表情错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16410"] = "返回码描述: 提取年龄错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16411"] = "返回码描述: 提取姿态错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16412"] = "返回码描述: 提取眼镜错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16413"] = "返回码描述: 提取魅力值错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16414"] = "返回码描述: 语音合成失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16415"] = "返回码描述: 图片为空, 建议处理方式: 请检查图片是否正常"
	RejectReasonMaps["16416"] = "返回码描述: 个体已存在, 建议处理方式: 请检查个体是否已存在"
	RejectReasonMaps["16417"] = "返回码描述: 个体不存在, 建议处理方式: 请检查个体是否不存在"
	RejectReasonMaps["16418"] = "返回码描述: 人脸不存在, 建议处理方式: 请检查人脸是否不存在"
	RejectReasonMaps["16419"] = "返回码描述: 分组不存在, 建议处理方式: 请检查分组是否不存在"
	RejectReasonMaps["16420"] = "返回码描述: 分组列表不存在, 建议处理方式: 请检查分组列表是否不存在"
	RejectReasonMaps["16421"] = "返回码描述: 人脸个数超过限制, 建议处理方式: 请检查是否超过系统限制"
	RejectReasonMaps["16422"] = "返回码描述: 个体个数超过限制, 建议处理方式: 请检查是否超过系统限制"
	RejectReasonMaps["16423"] = "返回码描述: 组个数超过限制, 建议处理方式: 请检查是否超过系统限制"
	RejectReasonMaps["16424"] = "返回码描述: 对个体添加了几乎相同的人脸, 建议处理方式: 请检查个体已添加的人脸"
	RejectReasonMaps["16425"] = "返回码描述: 无效的图片格式, 建议处理方式: 请检查图片格式是否符号API要求"
	RejectReasonMaps["16426"] = "返回码描述: 图片模糊度检测失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16427"] = "返回码描述: 美食图片检测失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16428"] = "返回码描述: 提取图像指纹失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16429"] = "返回码描述: 图像特征比对失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16430"] = "返回码描述: OCR照片为空, 建议处理方式: 请检查待处理图片是否为空"
	RejectReasonMaps["16431"] = "返回码描述: OCR识别失败, 建议处理方式: 请尝试更换带有文字的图片"
	RejectReasonMaps["16432"] = "返回码描述: 输入图片不是身份证, 建议处理方式: 请检查图片是否为身份证"
	RejectReasonMaps["16433"] = "返回码描述: 名片无足够文本, 建议处理方式: 请检查名片是否正常"
	RejectReasonMaps["16434"] = "返回码描述: 名片文本行倾斜角度太大, 建议处理方式: 请检查名片是否正常"
	RejectReasonMaps["16435"] = "返回码描述: 名片模糊, 建议处理方式: 请检查名片是否正常"
	RejectReasonMaps["16436"] = "返回码描述: 名片姓名识别失败, 建议处理方式: 请尝试更换姓名显示清晰的名片图片"
	RejectReasonMaps["16437"] = "返回码描述: 名片电话识别失败, 建议处理方式: 请尝试更换电话显示清晰的名片图片"
	RejectReasonMaps["16438"] = "返回码描述: 图像为非名片图像, 建议处理方式: 请尝试更换有效的名片图片"
	RejectReasonMaps["16439"] = "返回码描述: 检测或者识别失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16440"] = "返回码描述: 未检测到身份证, 建议处理方式: 请对准边框(避免拍摄时倾角和旋转角过大、摄像头)"
	RejectReasonMaps["16441"] = "返回码描述: 请使用第二代身份证件进行扫描, 建议处理方式: 请使用第二代身份证进行处理"
	RejectReasonMaps["16442"] = "返回码描述: 不是身份证正面照片, 建议处理方式: 请使用带证件照的一面进行处理"
	RejectReasonMaps["16443"] = "返回码描述: 不是身份证反面照片, 建议处理方式: 请使用身份证反面进行进行处理"
	RejectReasonMaps["16444"] = "返回码描述: 证件图片模糊, 建议处理方式: 请确保证件图片清晰"
	RejectReasonMaps["16445"] = "返回码描述: 请避开灯光直射在证件表面, 建议处理方式: 请避开灯光直射在证件表面"
	RejectReasonMaps["16446"] = "返回码描述: 行驾驶证OCR识别失败, 建议处理方式: 请尝试更换有效的行驾驶证图片"
	RejectReasonMaps["16447"] = "返回码描述: 通用OCR识别失败, 建议处理方式: 请尝试更换带有文字的图片"
	RejectReasonMaps["16448"] = "返回码描述: 银行卡OCR预处理错误, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16449"] = "返回码描述: 银行卡OCR识别失败, 建议处理方式: 请尝试更换有效的银行卡图片"
	RejectReasonMaps["16450"] = "返回码描述: 营业执照OCR预处理失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16451"] = "返回码描述: 营业执照OCR识别失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16452"] = "返回码描述: 意图识别超时, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16453"] = "返回码描述: 闲聊处理超时, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16454"] = "返回码描述: 语音识别解码失败, 建议处理方式: 请检查语音参数是否正确编码"
	RejectReasonMaps["16455"] = "返回码描述: 语音过长或空, 建议处理方式: 请检查语音参数是否正确编码或者长度是否合法"
	RejectReasonMaps["16456"] = "返回码描述: 翻译引擎失败, 建议处理方式: 请联系客服反馈问题"
	RejectReasonMaps["16457"] = "返回码描述: 不支持的翻译类型, 建议处理方式: 请检查翻译类型参数是否合法"
	RejectReasonMaps["16460"] = "返回码描述: 输入图片与识别场景不匹配, 建议处理方式: 请检查场景参数是否正确，所传图片与场景是否匹配"
	RejectReasonMaps["16461"] = "返回码描述: 识别结果为空, 建议处理方式: 当前图片无法匹配已收录的标签，请尝试更换图片"
	RejectReasonMaps["16462"] = "返回码描述: 多人脸检测识别结果为空, 建议处理方式: 图片中识别不出人脸，请尝试更换图片"
	RejectReasonMaps["16467"] = "返回码描述: 跨年龄人脸识别出错, 建议处理方式: 请尝试更换有人脸的图片"
	RejectReasonMaps["16468"] = "返回码描述: 跨年龄人脸识别结果为空, 建议处理方式: 源图片与目标图片中识别不出匹配的人脸，请尝试更换图片"
	RejectReasonMaps["16472"] = "返回码描述: 音频鉴黄识别出错, 建议处理方式: 请确保音频地址能正常下载音频，尝试更换音频"
	for k, _ := range RejectReasonMaps {
		flag := false
		if k == code {
			flag = true
		}
		if flag {
			return RejectReasonMaps[code]
		}
	}
	return "未知错误"
}
