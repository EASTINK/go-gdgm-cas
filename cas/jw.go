package cas

import (
	"reflect"
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
		LogPrintln("检查登陆状态出错")
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
		LogPrintln("教务网好像掉线了呢？")
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
	LogPrintln("获得教务基础信息 —— 当前教学周为" + SInfo.SWeek.CurrentWeek)
	savejson(SInfo, g.Savedir+"/Sinfo.json", "学生信息已保存", "学生信息保存失败")
}

// 访问当天课表并保存数据
func (g *JW) jw_save_today() {
	Table.Day = Table.Day[:0]
	//转到当天表
	g.CAS.page.Navigate(todaycourse_url())
	g.CAS.cas_wait()
	if !g.jw_logon() {
		LogPrintln("登陆状态掉了？")
		return
	}
	tr := g.CAS.page.MustElements("#tab1 > tbody > tr")
	g.CAS.cas_wait()
	for _, td := range tr {
		td := td.MustElements("td")
		for _, p := range td {
			text, err := p.Element("p")
			if err == nil {
				kb := s2l(text.MustProperty("title").String())
				kb_2s(kb)
			}
		}
	}
	bz := g.CAS.page.MustElementX("/html/body/form/table/tbody/tr[7]/td[2]").MustText()
	if bz == "" {
		Table.Week = nil
	}
	matches := weekTrim(bz)
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
	}
	savejson(Table, g.Savedir+"/week.json", "学生课表保存成功.", "学生课表保存失败.")
}

func (c *Myscore) init() {
	defaultValues := map[string]string{
		"Semester":         "--",
		"CourseId":         "--",
		"CourseName":       "--",
		"GroupName":        "--",
		"Score":            "--",
		"ScoreMark":        "--",
		"Credit":           "--",
		"Cperiod":          "--",
		"CGPA":             "--",
		"ReSemester":       "--",
		"AssessmentMethod": "--",
		"AssessmentType":   "--",
		"CourseProperties": "--",
		"CourseType":       "--",
		"CourseCategory":   "--",
	}
	for fName, dValue := range defaultValues {
		field := reflect.ValueOf(c).Elem().FieldByName(fName)
		if field.Kind() == reflect.String && field.Len() == 0 {
			field.SetString(dValue)
		}
	}
}

func (g *JW) jw_save_score() {
	//转到成绩页面
	g.CAS.page.Navigate(coure_score)
	g.CAS.cas_wait()
	if !g.jw_logon() {
		LogPrintln("登陆状态掉了？")
		return
	}
	div, err := g.CAS.page.ElementsX("/html/body/div") //[0].MustText()
	g.CAS.cas_wait()
	if len(div) <= 0 || err != nil {
		LogPrintln("成绩解析出错")
		return
	}
	t, err := div[0].Text()
	g.CAS.cas_wait()
	if err != nil {
		LogPrintln("成绩解析出错")
		return
	}
	Score := scoreTrim(t)
	savejson(Score.AvgInfo, g.Savedir+"/score.json", "学生成绩更新成功.", "学生成绩更新失败.")
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
