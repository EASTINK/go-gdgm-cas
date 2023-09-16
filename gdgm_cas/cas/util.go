package cas

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

func LogPrintln(v ...any) {
	log.Println(v)
}

func savebytes(data []byte, savepath string, msg string) {
	if err := os.MkdirAll(filepath.Dir(savepath), os.ModePerm); err == nil {
		if err = os.WriteFile(savepath, data, os.ModePerm); err == nil {
			LogPrintln(msg)
		}
	}
}

// 输入课程有效周
// 输入： 1,4,7,9-11周,14-15周,17-18周
// 输出：[1,4,7,9,10,11,14,15,17,18]
func kb_week_trim(strweek string) []int {
	re := regexp.MustCompile(`(\d+(-\d+)+|\d+)`)
	matches := re.FindAllString(strweek, -1)
	weekMap := make(map[int]bool)
	for _, part := range matches {
		if strings.Contains(part, "-") {
			// 处理数字范围
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) >= 2 {
				start, _ := strconv.Atoi(rangeParts[0])
				end, _ := strconv.Atoi(rangeParts[len(rangeParts)-1])
				for i := start; i <= end; i++ {
					weekMap[i] = true
				}
			}
		} else {
			// 处理单个数字
			num, _ := strconv.Atoi(part)
			weekMap[num] = true
		}
	}
	// 将 map 中的周次提取为切片并排序
	var weeks []int
	for week := range weekMap {
		weeks = append(weeks, week)
	}
	sort.Ints(weeks)
	return weeks
}

// 将课表信息转换为json
func kb_2s(text string) {
	regexPattern := `课程学分：([\d.]+)<br/>课程属性：([^<]+)<br/>课程名称：([^<]+)<br/>上课时间：([^<]+)<br/>上课地点：([^<]+)`
	re := regexp.MustCompile(regexPattern)
	match := re.FindStringSubmatch(text)
	if len(match) != 6 {
		LogPrintln("JSON:无法解析文本")
		return
	}
	table := &Timetable{
		Credit:  match[1],
		Cprop:   match[2],
		Cname:   match[3],
		Ctime:   match[4],
		Clocale: match[5],
	}
	Table.Day = append(Table.Day, table)
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

// 登陆密码编码
func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

// 返回验证码的base64格式
func casbase64(savedir string) string {
	imageBytes, err := os.ReadFile(savedir + "/captcha.jpg")
	if err != nil {
		LogPrintln("打开验证码图片失败:", err)
		return ""
	}
	base64String := base64.StdEncoding.EncodeToString(imageBytes)
	return base64String
}

// 保存Cookies
func saveCookies(savedir string, cookies []*proto.NetworkCookie) {
	data, err := json.Marshal(cookies)
	if err != nil {
		LogPrintln("读取客户端Cookies失败")
		return
	}
	savebytes(data, savedir+"/Cookies.json", "客户端Cookies已保存")
}

// 加载Cookies
func loadCookies(savedir string) ([]*proto.NetworkCookieParam, error) {
	var cookies []*proto.NetworkCookieParam
	data, err := os.ReadFile(savedir + "/Cookies.json")
	if err != nil {
		LogPrintln("读取本地Cookies失败.")
		return nil, err
	}
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}

// i2s
func i2s(x int) string {
	return strconv.Itoa(x)
}

// 2lower
func s2l(x string) string {
	return strings.ToLower(x)
}

// wait sometime
func caswait(value int64) {
	time.Sleep(time.Duration(value) * time.Second)
}

// 保存结构到本地json
func savejson(v interface{}, savepath string, okmsg string, errmsg string) {
	jsondata, err := json.Marshal(v)
	if err == nil {
		savebytes(jsondata, savepath, okmsg)
	} else {
		LogPrintln(errmsg)
	}
}

// weekTrim
func weekTrim(bz string) [][]string {
	return regexp.MustCompile(`([^ ]+) ([^ ]+) ([^;]+);`).FindAllStringSubmatch(strings.Trim(bz, "\t"), -1)
}

// scoreTrim
func scoreTrim(t string) *Avgcore {
	text := strings.Split(t, "\n")
	// text :=
	avgtext := regexp.MustCompile(`[+-]?\d+(\.\d+)?`).FindAllString(text[0], -1)
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
	return Score
}

// NewSweek
func NewSweek(text string) *Week {
	//text=第X周 ？ 当前日期不在教学周历内
	var res string
	if res = regexp.MustCompile("[0-9]+").FindString(text); res == "" {
		LogPrintln("获取学期周信息失败")
		return nil
	}
	cweek, err := strconv.Atoi(res)
	if err != nil {
		LogPrintln("获取学期周信息失败")
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

// todaystr
func todaycourse_url() string {
	return today_course + time.Now().Format(time.DateOnly)
}

// sq_split
func sq_split_val(sq_url string) []string {
	return strings.Split(regexp.MustCompile(`code=[^&]+`).FindString(sq_url), "code=")
}

// sq_url_val
func sq_url_val(sq_url string) bool {
	return regexp.MustCompile(PC_uc_pwd).MatchString(sq_url)
}

// logon_val
func logon_val(url string) bool {
	return regexp.MustCompile(M_uc_loginSuccess).MatchString(url)
}

// gettask
func gethwtask(body string, savedir string) *Hwtask {
	var task Hwtask
	//反序列
	if err := json.Unmarshal([]byte(body), &task); err != nil {
		return nil
	}
	task.savedir = savedir
	//返回任务列表
	return &task
}

// wait input code
func input_code() string {
	LogPrintln("登陆环境异常，请手动输入验证码：")
	var code string
	fmt.Scanf("%s", &code)
	code = string([]rune(strings.Trim(code, " "))[:4])
	return code
}

// save logfile
func Log_save() *os.File {
	LogPrintln(GetCurrentAbPath() + `/log/`)
	if err := os.MkdirAll(filepath.Dir(GetCurrentAbPath()+`/log/`), os.ModePerm); err == nil {
		//格式改动 win下不支持名字带":" | 注意在Linux下不要使用go run .,你应该通过build或者IDE的debug,不然拿到的目录是错的
		if logFile, err := os.OpenFile(GetCurrentAbPath()+`/log/`+time.Now().Format("2006-01-02 15-04-05")+`.log`, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err == nil {
			multiWriter := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(multiWriter)
			return logFile
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}

}

// 最终方案-全兼容
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取系统临时目录，兼容go run
func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

// params: level int
// 返回rooms map-0-3级菜单
func Rooms_menu(source_map map[string]map[string]map[string]string, key ...string) []string {
	var (
		menu1 []string
		menu2 []string
		menu3 []string
	)
	for i, opt := range key {
		if i == 0 && len(key) >= 1 {
			//keys()
			for l, _ := range source_map[opt] {
				menu2 = append(menu2, l)
			}
			// return menu2 由于后面还有可能要处理 所以不能直接return
		}
		if i == 1 && len(key) == 2 {
			for l, _ := range source_map[key[0]][opt] {
				menu3 = append(menu3, l)
			}
			return menu3
		}
	}
	if len(key) == 0 {
		for l, _ := range source_map {
			menu1 = append(menu1, l)
		}
		return menu1
	}

	if len(key) == 1 {
		return menu2
	}
	return nil
}
