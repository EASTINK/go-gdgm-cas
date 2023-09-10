package cas

import (
	"time"

	"github.com/go-rod/rod"
)

type CAS struct {
	Username   string
	Password   string
	Savedir    string
	API_KEY    string
	SECRET_KEY string
	WaitTime   int64
	page       *rod.Page
}

type CASOptions func(*CAS)

//----------

type JW struct {
	CAS     *CAS
	Savedir string
	// today_course_path string
	// course_score_path string
}

//------------------

type OCR struct {
	API_KEY    string
	SECRET_KEY string
	Accurate   bool   //是否启动高精度模式
	ImagePath  string //本地图片路径
}

//------------------

type UMOOC struct {
	cas     *CAS
	user    *user
	savedir string
}

//------------------

type Timetable struct {
	Credit  string `json:"课程学分"`
	Cprop   string `json:"课程属性"`
	Cname   string `json:"课程名称"`
	Ctime   string `json:"上课时间"`
	Clocale string `json:"上课地点"`
}

type Weektable struct {
	Cname     string `json:"课程名称"`
	Cter      string `json:"上课老师"`
	ValidWeek []int  `json:"有课周次"` //课程会因为调课安排导致不连续,所以直接生产有课数组
}

type table struct {
	Day  []*Timetable `json:"日程:"`
	Week []*Weektable `json:"周程"`
}

/*
学生姓名:
学生编号：
所属院系：
专业名称：
班级名称：
*/
type User struct {
	Sname  string `json:"学生姓名"`
	Stuid  string `json:"学生编号"`
	Stuyx  string `json:"所属院系"`
	Smajor string `json:"专业名称"`
	Sclass string `json:"班级名称"`
	SWeek  *Week  `json:"学期周"`
}

type Week struct {
	CurrentWeek      string    `json:"当前周"` //当前所在周
	startDate        time.Time //学期开始日期
	currentDate      time.Time //今日日期
	currentWeekStart time.Time //本周起始日期
	prevWeekStart    time.Time //上一周起始日期
	nextWeekStart    time.Time //下一周起始日期
}

type WordsResult struct {
	Words string `json:"words"`
}

type JSONData struct {
	WordsResult    []WordsResult `json:"words_result"`
	WordsResultNum int           `json:"words_result_num"`
	LogID          int64         `json:"log_id"`
}

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

type user struct {
	username string `json:"帐号"`
	password string `json:"密码"`
}

type Avgcore struct {
	CourseNumber int64   `json:"所修门数"`   //所修门数
	CourseTotal  float64 `json:"所修总学分"`  //所修总学分
	ASGPA        float64 `json:"平均学分绩点"` //平均学分绩点
	ACGPA        float64 `json:"平均成绩"`   //平均成绩
	AvgInfo      []*Myscore
}

type Myscore struct {
	Semester         string `json:"开课学期"`  //学期
	CourseId         string `json:"课程编号"`  //课程编号
	CourseName       string `json:"课程名称"`  //课程名称
	GroupName        string `json:"分组名"`   //课程分组
	Score            string `json:"成绩"`    //课程成绩
	ScoreMark        string `json:"成绩标识"`  //成绩标识
	Credit           string `json:"学分"`    //课程学分
	Cperiod          string `json:"总学时"`   //课程学时
	CGPA             string `json:"绩点"`    //课程绩点
	ReSemester       string `json:"补重学期"`  //补重学期
	AssessmentMethod string `json:"考核方式"`  //考核方式
	AssessmentType   string `json:"考试性质"`  //考试性质
	CourseProperties string `json:"课程属性"`  //课程属性
	CourseType       string `json:"课程性质"`  //课程性质
	CourseCategory   string `json:"通选课类别"` //课程类别
}

// --------------
const (
	cas_sfrz = "https://sfrz.gdgm.cn/authserver/login?service=https://eportal.gdgm.cn/login?service=https://eportal.gdgm.cn/new/index.html?browser=no"
	//----------------------
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
	//----------------------
	通用文字识别标准  = "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic"
	通用文字识别高精度 = "https://aip.baidubce.com/rest/2.0/ocr/v1/accurate_basic"
	//----------------------
	cas_jw       = "https://eportal.gdgm.cn/appShow?appId=5759540940956162"
	jw_info      = "https://jw.gdgm.cn/jsxsd/framework/xsMain_new.jsp"
	today_course = "https://jw.gdgm.cn/jsxsd/framework/main_index_loadkb.jsp?rq="
	coure_score  = "https://jw.gdgm.cn/jsxsd/kscj/cjcx_list"
)

var (
	SInfo *User
	Score *Avgcore
	Table = &table{}
)
