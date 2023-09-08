package cas

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type JW struct {
	CAS     *CAS
	Savedir string
	// today_course_path string
	// course_score_path string
}

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
	ValidWeek []int  `json:"有课周次"` //课程会因为调课安排导致不连续
}

type table struct {
	Day  []*Timetable `json:"日程:`
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

var (
	cas_jw       = "https://eportal.gdgm.cn/appShow?appId=5759540940956162"
	jw_info      = "https://jw.gdgm.cn/jsxsd/framework/xsMain_new.jsp"
	today_course = "https://jw.gdgm.cn/jsxsd/framework/main_index_loadkb.jsp?rq="
	coure_score  = "https://jw.gdgm.cn/jsxsd/kscj/cjcx_list"
	SInfo        *User
	Score        *Avgcore
	Table        *table
)

/**
 * 从页面标题判断教务系统
 * @return bool 教务系统的登陆状态
 */
func (g *JW) jw_logon() bool {
	//失效 ：9.06 遇到没有页面没有标题的情况 增加一个异常判断 err -> false
	// title, err := g.CAS.page.Element("head > title")
	// if err == nil {
	// 	return title.MustText() != "登录"
	// }
	// return false
	//----------
	//9.06 	尝试用document.Title获取页面标题
	title, err := g.CAS.page.Info()
	if err == nil {
		return title.Title != "登录"
	} else {
		fmt.Println("检查登陆状态出错")
	}
	return false
}

// CAS登陆教务 自动提取数据
func (g *JW) Jw_cas_start() {
	err := g.CAS.page.Navigate(cas_jw)
	g.CAS.cas_wait()
	if err == nil {
		//登陆成功 开始干活！：
		if g.jw_logon() {
			//-获取基础信息
			g.jw_save_info()
			//-获取当周课表
			g.jw_save_today()
			//-获取学生成绩
			g.jw_save_score()
		}
	} else {
		fmt.Println("教务网好像掉线了呢？")
	}
}
func (g *JW) jw_save_info() {
	SInfo = &User{}
	g.CAS.page.MustNavigate(jw_info)
	g.CAS.cas_wait()
	//名
	SInfo.Sname = g.CAS.page.MustElementX("/html/body/div/div[1]/div[1]/div[1]/div[2]/div/div[2]/div[2]").MustText()
	// //号
	SInfo.Stuid = g.CAS.page.MustElementX("/html/body/div/div[1]/div[1]/div[1]/div[2]/div/div[3]/div[2]").MustText()
	// //学院
	SInfo.Stuyx = g.CAS.page.MustElementX("/html/body/div/div[1]/div[1]/div[1]/div[2]/div/div[4]/div[2]").MustText()
	// //专业
	SInfo.Smajor = g.CAS.page.MustElementX("/html/body/div/div[1]/div[1]/div[1]/div[2]/div/div[5]/div[2]").MustText()
	// //班级
	SInfo.Sclass = g.CAS.page.MustElementX("/html/body/div/div[1]/div[1]/div[1]/div[2]/div/div[6]/div[2]").MustText()
	SInfo.SWeek = NewSweek(g.CAS.page.MustElementX(`/html/body/div/div[1]/div[1]/div[2]/div[2]/div/div/div[1]/div[1]/span`).MustText())
	fmt.Println("获得教务基础信息 —— 当前教学周为" + SInfo.SWeek.CurrentWeek)
	jsondata, err := json.Marshal(SInfo)
	if err == nil {
		savebytes(jsondata, g.Savedir+"/Sinfo.json", "学生信息已保存")
	} else {
		fmt.Println("学生信息保存失败")
	}
}

// 访问当天课表并保存数据
func (g *JW) jw_save_today() {
	Table.Day = Table.Day[:0]
	//转到当天表
	g.CAS.page.Navigate(today_course + time.Now().Format(time.DateOnly))
	g.CAS.cas_wait()
	if !g.jw_logon() {
		fmt.Println("登陆状态掉了？")
		return
	}
	tr := g.CAS.page.MustElements("#tab1 > tbody > tr")
	g.CAS.cas_wait()
	for _, td := range tr {
		td := td.MustElements("td")
		for _, p := range td {
			text, err := p.Element("p")
			if err == nil {
				kb := strings.ToLower(text.MustProperty("title").String())
				// fmt.Println(kb_2json(kb))
				data, ok := kb_2s(kb)
				if ok {
					Table.Day = append(Table.Day, data)
				}
			}
		}
	}
	bz := g.CAS.page.MustElementX("/html/body/form/table/tbody/tr[7]/td[2]").MustText()
	if bz == "" {
		Table.Week = nil
	}
	matches := regexp.MustCompile(`([^ ]+) ([^ ]+) ([^;]+);`).FindAllStringSubmatch(strings.Trim(bz, "\t"), -1)
	// 提取课程信息
	for _, match := range matches {
		courseName := match[1]
		teacher := match[2]
		week := match[3]

		Table.Week = append(Table.Week, &Weektable{
			Cname:     courseName,
			Cter:      teacher,
			ValidWeek: kb_week_trim(week),
		})
		// fmt.Printf("课程名：%s\n", courseName)
		// fmt.Printf("教师：%s\n", teacher)
		// fmt.Printf("周次：%s\n", week)
	}

	jsondata, err := json.Marshal(Table.Day)
	if err == nil {
		savebytes(jsondata, g.Savedir+"/week.json", "今日课表保存成功。")
		return
	}
	fmt.Println("今日课表保存失败。")
}

// 输入课程有效周
// 输入： 1,4,7,9-11周,14-15周,17-18周
// 输出：[1,4,7,9,10,11,14,15,17,18]
func kb_week_trim(strweek string) []int {
	//。。。。
	return nil
}

// 将课表信息转换为json
func kb_2s(text string) (*Timetable, bool) {
	regexPattern := `课程学分：([\d.]+)<br/>课程属性：([^<]+)<br/>课程名称：([^<]+)<br/>上课时间：([^<]+)<br/>上课地点：([^<]+)`
	re := regexp.MustCompile(regexPattern)
	match := re.FindStringSubmatch(text)
	if len(match) != 6 {
		fmt.Println("JSON:无法解析文本")
		return &Timetable{}, false
	}
	table := &Timetable{
		Credit:  match[1],
		Cprop:   match[2],
		Cname:   match[3],
		Ctime:   match[4],
		Clocale: match[5],
	}
	return table, true
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

func (c *Myscore) init() {
	if c.Semester == "" {
		c.Semester = "--"
	}
	if c.CourseId == "" {
		c.CourseId = "--"
	}
	if c.CourseName == "" {
		c.CourseName = "--"
	}
	if c.GroupName == "" {
		c.GroupName = "--"
	}
	if c.Score == "" {
		c.Score = "--"
	}
	if c.ScoreMark == "" {
		c.ScoreMark = "--"
	}
	if c.Credit == "" {
		c.Credit = "--"
	}
	if c.Cperiod == "" {
		c.Cperiod = "--"
	}
	if c.CGPA == "" {
		c.CGPA = "--"
	}
	if c.ReSemester == "" {
		c.ReSemester = "--"
	}
	if c.AssessmentMethod == "" {
		c.AssessmentMethod = "--"
	}
	if c.AssessmentType == "" {
		c.AssessmentType = "--"
	}
	if c.CourseProperties == "" {
		c.CourseProperties = "--"
	}
	if c.CourseType == "" {
		c.CourseType = "--"
	}
	if c.CourseCategory == "" {
		c.CourseCategory = "--"
	}
}

type Avgcore struct {
	CourseNumber int64   `json:"所修门数"`   //所修门数
	CourseTotal  float64 `json:"所修总学分"`  //所修总学分
	ASGPA        float64 `json:"平均学分绩点"` //平均学分绩点
	ACGPA        float64 `json:"平均成绩"`   //平均成绩
	AvgInfo      []*Myscore
}

func (g *JW) jw_save_score() {
	//转到成绩页面
	g.CAS.page.Navigate(coure_score)
	g.CAS.cas_wait()
	if !g.jw_logon() {
		fmt.Printf("登陆状态掉了？")
		return
	}
	div, err := g.CAS.page.ElementsX("/html/body/div") //[0].MustText()
	g.CAS.cas_wait()
	if len(div) <= 0 || err != nil {
		fmt.Printf("成绩解析出错")
		return
	}
	t, err := div[0].Text()
	g.CAS.cas_wait()
	if err != nil {
		fmt.Printf("成绩解析出错")
		return
	}
	text := strings.Split(t, "\n")
	// text :=
	re, _ := regexp.Compile(`[+-]?\d+(\.\d+)?`)
	avgtext := re.FindAllString(text[0], -1)
	number, _ := strconv.ParseInt(avgtext[0], 10, 64)
	total, _ := strconv.ParseFloat(avgtext[1], 64)
	asgpa, _ := strconv.ParseFloat(avgtext[2], 64)
	acgpa, _ := strconv.ParseFloat(avgtext[3], 64)
	Score = &Avgcore{
		CourseNumber: number,
		CourseTotal:  total,
		ASGPA:        asgpa,
		ACGPA:        acgpa,
		AvgInfo:      []*Myscore{},
	}
	for _, item := range text[2:] {
		c := strings.Split(item, "\t")
		k := &Myscore{
			Semester:         c[1],
			CourseId:         c[2],
			CourseName:       c[3],
			GroupName:        c[4],
			Score:            c[5],
			ScoreMark:        c[6],
			Credit:           c[7],
			Cperiod:          c[8],
			CGPA:             c[9],
			ReSemester:       c[10],
			AssessmentMethod: c[11],
			AssessmentType:   c[12],
			CourseProperties: c[13],
			CourseType:       c[14],
			CourseCategory:   c[15],
		}
		k.init()
		Score.AvgInfo = append(Score.AvgInfo, k)
	}
	jsondata, err := json.Marshal(Score)
	if err == nil {
		savebytes(jsondata, g.Savedir+"/score.json", "学生成绩更新成功。")
		return
	}
	fmt.Println("学生成绩更新失败。")
}

// 计算各类的起始日期
func getWeekInfo(startDate time.Time) (time.Time, time.Time, time.Time, time.Time) {
	// 计算当前日期距离起始日期的天数
	days := int(time.Now().Sub(startDate).Hours() / 24)
	// 计算当前所在的周数
	currentWeek := (days / 7) + 1
	// 计算本周的起始日期
	currentWeekStart := startDate.AddDate(0, 0, (currentWeek-1)*7)
	// 计算下一周的起始日期
	nextWeekStart := currentWeekStart.AddDate(0, 0, 7)
	// 计算上一周的起始日期
	prevWeekStart := currentWeekStart.AddDate(0, 0, -7)
	return time.Now(), currentWeekStart, nextWeekStart, prevWeekStart
}

// 依据当前周 返回本学期起始时间
func getStartDate(weeks int) time.Time {
	// 往前推算 weeks 周，每周 7 天
	startDate := time.Now().AddDate(0, 0, -(weeks-1)*7)
	// 找到最近的周日
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, -1)
	}
	return startDate
}

func NewSweek(text string) *Week {
	//text=第X周 ？ 当前日期不在教学周历内
	var res string
	if res = regexp.MustCompile("[0-9]+").FindString(text); res == "" {
		fmt.Println("获取学期周信息失败")
		return nil
	}
	cweek, err := strconv.Atoi(res)
	if err != nil {
		fmt.Println("获取学期周信息失败")
		return nil
	}
	startDate := getStartDate(cweek)
	currentDate, currentWeekStart, nextWeekStart, prevWeekStart := getWeekInfo(startDate)
	return &Week{
		CurrentWeek:      text,
		startDate:        startDate,
		currentDate:      currentDate,
		currentWeekStart: currentWeekStart,
		prevWeekStart:    prevWeekStart,
		nextWeekStart:    nextWeekStart,
	}
}

// 返回下一周次课表
func (g *User) nextweek() {
	//g.SWeek -> 第x+1周 9.3
	// 切换下个周末即可
}

// 返回上一周次课表
func (g *User) lastweek() {
	//g.Sweek -> 第x-1周末 9.3开始
}
