package cas

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

/**
 * 将本地图片转换为BASE64编码后的URL
 * @param string  imagepath 图片路径
 * @return string 需要的拼接数据
 */
func (o *OCR) capbase() string {
	imageBytes, err := os.ReadFile(o.ImagePath)
	if err != nil {
		LogPrintln("打开验证码图片失败:", err)
		return "failds"
	}
	//编码base64->urlEncode
	base64String := base64.StdEncoding.EncodeToString(imageBytes)
	return url.QueryEscape(base64String)
}

/**
 * 使用 AK，SK 生成鉴权签名（Access Token）
 * @return string 鉴权签名信息（Access Token）
 */
func (o *OCR) getAccessToken() string {
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", o.API_KEY, o.SECRET_KEY)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		LogPrintln(err)
		return "failds"
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogPrintln(err)
		return "failds"
	}
	accessTokenObj := map[string]string{}
	json.Unmarshal([]byte(body), &accessTokenObj)
	return accessTokenObj["access_token"]
}

/**
 * 取得OCR服务返回的值
 * @return string 验证码
 */
func (o *OCR) Cap() string {

	var url string
	if o.Accurate {
		url = 通用文字识别高精度
	} else {
		url = 通用文字识别标准
	}
	url += "?access_token=" + o.getAccessToken()
	payload := strings.NewReader("image=" + o.capbase())
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		LogPrintln(err)
		return "failds"
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		LogPrintln(err)
		return "failds"
	}
	defer res.Body.Close()
	jsonData, err := io.ReadAll(res.Body)
	if err != nil {
		LogPrintln(err)
		return "failds"
	}
	var data JSONData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		LogPrintln("解析JSON失败:", err)
		return "failds"
	}
	// 获取words的值
	if len(data.WordsResult) > 0 {
		words := data.WordsResult[0].Words
		return words
	}
	return "failds"
}
