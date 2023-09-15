package cas

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func AK(value string) CASOptions {
	return func(g *CAS) {
		g.API_KEY = value
	}
}

func SK(value string) CASOptions {
	return func(g *CAS) {
		g.SECRET_KEY = value
	}
}

func Wtime(value int64) CASOptions {
	return func(g *CAS) {
		g.WaitTime = value
	}
}

// 模拟登陆
func (g *CAS) AutoLogin() bool {
	var logon bool
	failds := 0
	for {
		logon = g.cas_login(g.Username, g.Password)
		if logon {
			LogPrintln("登陆成功")
			break
		}
		if failds >= 2 {
			LogPrintln("登陆失败")
			break
		}
		failds += 1
		LogPrintln("登陆出错，正在尝试重新提交登陆..." + i2s(failds) + "/2")
	}
	return logon
}

// 通用验证码识别
func (g *CAS) CaptchCode(imgpath string) string {
	cap_n := 0
	code := ""
	for {
		code = "failds"
		accurate := false
		//如果标准版返回错误1次 就切换到高精度
		if cap_n >= 1 {
			accurate = true
		}
		//如何识别超过2次还是没有返回，立即中断
		if cap_n >= 2 {
			LogPrintln("验证码识别出错")
			return "failds"
		}
		//尝试向OCR服务上传图片
		code = (&OCR{
			API_KEY:    g.API_KEY,
			SECRET_KEY: g.SECRET_KEY,
			Accurate:   accurate,
			ImagePath:  imgpath,
		}).Cap()
		cap_n += 1
		//如果服务正常返回文本：
		if code != "failds" {
			return code
		}
	}
}

// 返回验证码的base64格式
func (g *CAS) capbase64() string {
	return casbase64(g.Savedir)
}

/**
 * return title != "统一身份认证"
 * 从页面标题判断数字工贸是否登陆
 * @return bool 数字工贸登陆状态
 */
func (g *CAS) cas_logon() bool {
	title := g.page.MustElement("head > title").MustText()
	return title != "统一身份认证"
}

// 保存Cookies
func (g *CAS) SaveCookies() {
	cookies := g.page.Browser().MustGetCookies()
	saveCookies(g.Savedir, cookies)
}

// 加载Cookies
func (g *CAS) loadCookies(browser *rod.Browser) {
	// cookies := (g.Savedir,cookcookies)
	cookies, err := loadCookies(g.Savedir)
	if err != nil {
		return
	}
	err = browser.SetCookies(cookies)
	if err != nil {
		LogPrintln("加载本地Cookies失败")
	} else {
		LogPrintln("已启用本地Cookies")
	}
}

// 模拟登陆数字工贸
func (g *CAS) cas_login(username string, password string) bool {
	u := launcher.New().
		Set("--disable-popup-blocking").MustLaunch()
	browser := rod.New().
		ControlURL(u).
		MustConnect().NoDefaultDevice()

	g.loadCookies(browser)
	g.page = browser.MustPage(cas_sfrz).MustWindowFullscreen()

	//异步处理弹窗事件
	go g.page.EachEvent(func(e *proto.PageJavascriptDialogOpening) {
		_ = proto.PageHandleJavaScriptDialog{Accept: false, PromptText: ""}.Call(g.page)
	})()

	//判断携带的Cookes后是否已经登陆成功：
	if !g.cas_logon() {
		//还需要登陆：
		g.page.MustElement("#username").MustInput(username)
		g.page.MustElement("#password").MustInput(password)
		cap := g.page.MustElementX("/html/body/div[3]/div[2]/div[2]/div/div[3]/div/form/p[3]")
		//判断是否需要填写验证码：
		jpg, err := cap.Elements("img")
		if err == nil && len(jpg) > 0 {
			//保存到本地，等待前端交互
			jpg[0].MustWaitStable().MustScreenshot(g.Savedir + "/caplogon.jpg")
			var code string
			//判别有没有传进来ak&sk
			if g.API_KEY != "" && g.SECRET_KEY != "" {
				code = g.CaptchCode(g.Savedir + "/caplogon.jpg")
			} else {
				code = input_code()
			}
			if code != "failds" {
				cap.MustElements("#captchaResponse")[0].MustInput(code)
			} //如果识别失败会返回failds,随着表单提交一起刷新
		}
		//启动表单
		g.page.MustElement("#casLoginForm > p:nth-child(4) > button").MustClick()
		return g.cas_logon()
	}
	return true
}

// 等待一段时间
func (g *CAS) cas_wait() {
	caswait(g.WaitTime)
}

func (g *CAS) NewJW(save_dir string) *JW {
	return &JW{
		CAS:     g,
		Savedir: save_dir,
	}
}

func (g *CAS) NewCard(save_dir string) *Card {
	return &Card{
		CAS:     g,
		Savedir: save_dir,
	}
}

// @params :数字工贸帐号，密码，验证码保存地址，每步等待时间，百度云ak,sk
func NewCAS(username string, password string, savedir string,
	options ...CASOptions) *CAS {
	///, waittime int64
	//ocr_api_key string, ocr_sec_key string) *CAS {
	cas := &CAS{
		Username: username,
		Password: password,
		Savedir:  savedir,
		// WaitTime:   waittime,
		// API_KEY:    ocr_api_key,
		// SECRET_KEY: ocr_sec_key,
		page: &rod.Page{},
	}
	for _, option := range options {
		option(cas)
	}
	if cas.WaitTime == 0 {
		cas.WaitTime = 1 //至少等待1秒
	}

	return cas
}
