package cas

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

type UMOOC struct {
	cas     *CAS
	user    *user
	savedir string
}

type user struct {
	username string `json:"帐号"`
	password string `json:"密码"`
}

var (
	//移动端登陆地址 （支持GET)
	mobile_login = "https://umooc.gdgm.cn/mobile/login_check.do"
	//CAS跳转到PCUMOOC
	cas_PCuc = "https://eportal.gdgm.cn/appShow?appId=5749077492672304"
	//设置密保
	PC_uc_sq = "https://umooc.gdgm.cn/meol/validateQuestion.do"
	//验证密保
	PC_uc_cpsq = "https://umooc.gdgm.cn/meol/findPasswdQuestionPreAccount.do"
	//设置密码
	PC_uc_pwd = "https://umooc.gdgm.cn/meol/findPasswdQuestionPreReset.do"
	//验证密码
	PC_uc_cpwd = "https://umooc.gdgm.cn/meol/findPasswdQuestionDone.do"
	//登陆失败
	M_uc_loginFaild = "https://umooc.gdgm.cn/mobile/loginFailed.do"
	//登陆成功
	M_uc_loginSuccess = "https://umooc.gdgm.cn/mobile/loginSuccess.do"
	//互评
	Mutualeval = "https://umooc.gdgm.cn/mobile/hw/stu/findOtherUnCommentHwtList.do"
	//问卷
	Questionnaire = "https://umooc.gdgm.cn/mobile/stuUnDoPaperTaskList.do"
	//测试
	TaskList = "https://umooc.gdgm.cn/mobile/stuUnDoTestTaskList.do"
	//作业
	HwtaskList = "https://umooc.gdgm.cn/mobile/hw/stu/findStuUnDoHwTaskList.do"
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
		fmt.Println("手机慕课帐密登陆失败")
		//帐密登陆
		if amend {
			fmt.Println("进入自动修正程序")
			u.casLogin()
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
			if u.changePwd() {
				//修改密码成功
				if u.mobile_login() {
					//登陆移动端慕课
					return true
				}
			}
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
	u.cas.page.MustElementX(`/html/body/div[2]/div/div/div/div[5]/div/span/img`).MustWaitStable().MustScreenshot("./jcaptcha.png")
	u.cas.cas_wait()
	//识别验证码
	var code string
	if code = u.cas.CaptchCode("./jcaptcha.png"); code == "failds" {
		fmt.Println("验证码出错")
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
	if sq_url = u.cas.page.MustInfo().URL; !regexp.MustCompile(PC_uc_pwd).MatchString(sq_url) {
		fmt.Println("密保验证失误")
		return false
	}
	if sq_split = strings.Split(regexp.MustCompile(`code=[^&]+`).FindString(sq_url), "code="); len(sq_split) <= 0 {
		fmt.Println("密保服务异常")
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
		fmt.Println("密码设置失败")
		return false
	}
	fmt.Println("密码设置成功")
	return true
}

// 手机慕课登陆
func (u *UMOOC) mobile_login() bool {
	err := u.cas.page.Navigate(mobile_login + "?j_username=" + u.user.username + "&j_password=" + MD5(u.user.password))
	u.cas.cas_wait()
	if err != nil || !u.uc_mobile_logon() {
		fmt.Println("手机慕课登陆出错")
		return false
	}
	fmt.Println("手机慕课登陆成功")
	return true
}

// 验证登陆状态
func (u *UMOOC) uc_mobile_logon() bool {
	if url := u.cas.page.MustInfo().URL; !regexp.MustCompile(M_uc_loginSuccess).MatchString(url) {
		return false
	}
	return true
}

// 登陆密码编码
func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

//-----------------作业------------------

type HWtListData struct {
	CourseList []struct {
		TeaRealName string `json:"teaRealName"` //教师
		Name        string `json:"name"`        //课程名
		ID          int    `json:"id"`          //课程ID
	} `json:"courseList"`
	HwtList []struct {
		CourseName    string `json:"courseName"`    //课程名
		StartDateTime string `json:"startDateTime"` //开始时间
		HWStatus      bool   `json:"hwStatus"`      //未知
		ID            int    `json:"id"`            //任务id
		Title         string `json:"title"`         //任务名
		Deadline      string `json:"deadline"`      //截止时间
		CourseID      int    `json:"courseId"`      //课程id
	} `json:"hwtList"`
}

type Hwtask struct {
	Datas     HWtListData `json:"datas"`     //作业任务列表
	SessionID string      `json:"sessionid"` //token
	Status    int         `json:"status"`    //状态 -> 正常为1
	Error     interface{} `json:"error"`     //错误提示
	savedir   string
}

// 作业
func (u *UMOOC) Hwtask(savedir string) *Hwtask {
	err := u.cas.page.Navigate(HwtaskList)
	if err == nil {
		u.cas.cas_wait()
		body := u.cas.page.MustElement("body").MustText()
		var task Hwtask
		//反序列
		if err = json.Unmarshal([]byte(body), &task); err == nil {
			task.savedir = savedir
			//返回任务列表
			return &task
		}
	}
	return nil
}

// 保存待完成的作业数据到本地
func (t *Hwtask) Save() {
	jsondata, err := json.Marshal(t)
	if err == nil {
		savebytes(jsondata, t.savedir+"/hwtask.json", "慕课作业已保存到本地")
		return
	}
	//保存到本地
	fmt.Println("慕课作业保存失败")
}

//-----------------测试------------------
//挖坑

//-----------------问卷------------------
//挖坑

//-----------------互评------------------
//挖坑
