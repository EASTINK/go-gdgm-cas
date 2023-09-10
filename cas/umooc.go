package cas

import (
	"github.com/dlclark/regexp2"
)

// 设置帐密
func NewUC(cas *CAS, username string, password string, savedir string) *UMOOC {
	re := regexp2.MustCompile(`^(?=.*\d)(?=.*[a-zA-Z])(?=.*[\W_])(?!.*[&><'"])[\da-zA-Z\W_]{8,20}$`, 0)
	//密码由8-20个字符组成，且必须包含数字、字母和特殊字符（&<>单双引号除外）
	if isMatch, _ := re.MatchString(password); isMatch {
		return &UMOOC{
			cas: cas,
			user: &user{
				username: username,
				password: password,
			},
			savedir: savedir,
		}
	} else {
		return nil
	}
}

/*
*7.9-19：46:

	由于cas只能跳转PC端，但移动端和PC端的接口不能用同一token进行验证
	为此 提供以下俩种方式进行跳转：
	amend:
		true.	CAS·PC -> 修改密码 -> UMOOC·UMOOC·移动端 -> 帐号密码 -> 访问
		false.	UMOOC·移动端 -> 帐号密码 -> 访问
	如果从CAS跳转会强制修正密码，请慎重选择登陆模式！
*/
func (u *UMOOC) AutoLogin(amend bool) {
	//移动端登陆，
	if !u.mobile_login() {
		//帐密登陆
		if amend {
			LogPrintln("进入自动修正程序")
			if u.casLogin() {
				LogPrintln("手机慕课修正登陆成功")
			} else {
				LogPrintln("手机慕课修正登陆失败")
			}
		}
	}
}

// 按所给的帐号密码设定帐号
func (u *UMOOC) casLogin() bool {
	//从数字工贸跳转
	err := u.cas.page.Navigate(cas_PCuc)
	if err == nil {
		if u.cas.cas_logon() {
			//PC登陆成功
			failds := 1
			for {
				if u.changePwd() {
					//修改密码成功
					if u.mobile_login() {
						//登陆移动端慕课
						break
					}
				}
				if failds >= 2 {
					LogPrintln("密保验证失败")
					return false
				}
				failds += 1
				LogPrintln("UMOOC密保验证出错，正在尝试重新提交验证..." + i2s(failds) + "/3")
			}
			return true
		}
	}
	return false
}

// 自动修改平台密码
func (u *UMOOC) changePwd() bool {
	//跳转到修改密保
	err := u.cas.page.Navigate(PC_uc_sq)
	u.cas.cas_wait()
	if err != nil {
		return false
	}
	//选填表单
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[1]/td/select").
		MustSelect("你高中班主任的名字")
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[2]/td/input").MustInput("1")
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[3]/td/select").
		MustSelect("你暗恋的人的名字")
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[4]/td/input").MustInput("2")
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[5]/td/select").
		MustSelect("你最喜欢的体育明星")
	u.cas.page.MustElementX("/html/body/div/form/div[2]/table/tbody/tr[6]/td/input").MustInput("3")
	//hook掉弹窗 失效
	//u.cas.page.MustEval("()=>{window.alert = function(s) {console.log(s)};}")
	//提交表单
	u.cas.cas_wait()
	u.cas.page.MustElementX("/html/body/div/form/div[3]/input[1]").MustClick()
	u.cas.cas_wait()
	//查询结果:
	result := u.cas.page.MustElement("head").MustText()
	if m, _ := regexp2.MustCompile(`[\u4e00-\u9fa5]+`, 0).FindStringMatch(result); m != nil {
		if m.String() != "密码问题设置成功" {
			return false
		}
	}
	//跳转改密页面
	err = u.cas.page.Navigate(PC_uc_cpsq)
	u.cas.cas_wait()
	if err != nil {
		return false
	}
	//获得验证码
	u.cas.cas_wait()
	u.cas.page.MustElementX(`/html/body/div[2]/div/div/div/div[5]/div/span/img`).MustWaitStable().MustScreenshot(u.savedir + "/UMooc.png")
	u.cas.cas_wait()
	//识别验证码
	code := "failds"

	//判别有没有传进来ak&sk
	if u.cas.API_KEY != "" && u.cas.SECRET_KEY != "" {
		code = u.cas.CaptchCode(u.savedir + "/UMooc.png")
	} else {
		code = input_code()
	}
	if code == "failds" {
		return false
	}
	//提交密保
	sq_sumbit := `$.ajax({
		type: "post",
		url: "findPasswdQuestionAccount.do",
		data: {
			"username": "` + u.user.username + `",
			"questionId": "401",
			"questionVal":encodeURIComponent("1", "gbk"),
			"questionId2": "403",
			"questionVal2":encodeURIComponent("2", "gbk"),
			"questionId3": "404",
			"questionVal3":encodeURIComponent("3", "gbk"),
			"imgcode": ` + `"` + code + `"
		},
		success: function (data) {
			data = JSON.parse(data);
			var status=data.status;
			if (status != 10000) {
				refresh_jcaptcha('#jcaptcha');
			} else {
				window.location.href = "/meol/findPasswdQuestionPreReset.do?code=" + data.code;
			}
		}
	});`
	u.cas.page.MustEval("()=>" + sq_sumbit)
	u.cas.cas_wait()
	//验证密保
	var (
		sq_url   string
		sq_split []string
	)
	if sq_url = u.cas.page.MustInfo().URL; !sq_url_val(sq_url) {
		LogPrintln("密保验证失误")
		return false
	}
	if sq_split = sq_split_val(sq_url); len(sq_split) <= 0 {
		LogPrintln("密保服务异常")
		return false
	}
	u.cas.cas_wait()
	//提交密码
	code = sq_split[1]
	pw_sumbit := `$.ajax({
		type: 'post',
		url: '/meol/findPasswdQuestionReset.do',
		data: {
			"code": "` + code + `",
			"firstPasswd": "` + u.user.password + `",
			"secondPasswd": "` + u.user.password + `"
		},
		success: function (data) {
			if (data == 20000) {
				window.location.href = "/meol/findPasswdQuestionDone.do";
			}
		}
	});`
	u.cas.page.MustEval("()=>" + pw_sumbit)
	u.cas.cas_wait()
	//验证密码
	if sq_url == u.cas.page.MustInfo().URL {
		LogPrintln("密码设置失败")
		return false
	}
	LogPrintln("密码设置成功")
	return true
}

// 手机慕课登陆
func (u *UMOOC) mobile_login() bool {
	err := u.cas.page.Navigate(mobile_login + "?j_username=" + u.user.username + "&j_password=" + MD5(u.user.password))
	u.cas.cas_wait()
	if err != nil || !u.uc_mobile_logon() {
		LogPrintln("手机慕课登陆出错")
		return false
	}
	LogPrintln("手机慕课登陆成功")
	return true
}

// 验证登陆状态
func (u *UMOOC) uc_mobile_logon() bool {
	if url := u.cas.page.MustInfo().URL; !logon_val(url) {
		return false
	}
	return true
}

//-----------------作业------------------

// 作业
func (u *UMOOC) Hwtask(savedir string) *Hwtask {
	err := u.cas.page.Navigate(HwtaskList)
	if err == nil {
		u.cas.cas_wait()
		body := u.cas.page.MustElement("body").MustText()
		if task := gethwtask(body, savedir); task != nil {
			return task
		}
	}
	return nil
}

// 保存待完成的作业数据到本地
func (t *Hwtask) Save() {
	if t == nil {
		LogPrintln("慕课作业保存失败")
		return
	}
	savejson(t, t.savedir+"/hwtask.json", "慕课作业已保存到本地", "慕课作业保存失败")
}

//-----------------测试------------------
//挖坑

//-----------------问卷------------------
//挖坑

//-----------------互评------------------
//挖坑
