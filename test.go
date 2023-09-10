package main

// "gdgm_cas/cas"
import (
	cas "gdgm_cas/cas"
)

func main() {
	defer cas.Log_save().Close()

	var (
		API_KEY    = "" //百度OCR
		SECRET_KEY = ""
		save_dir   = "./cas" //数据文件保存路径
		waitsecond = 2
	)
	//数字工贸帐号密码
	GDGM := cas.NewCAS("数字工贸帐号", "数字工贸密码", save_dir, cas.Wtime(int64(waitsecond)), cas.AK(API_KEY), cas.SK(SECRET_KEY))
	if !GDGM.AutoLogin() {
		cas.LogPrintln("数字工贸登陆失败-----即将退出程序..")
		return
	}
	//-------------
	//提取教务系统课表和成绩
	GDGM.NewJW(save_dir).Jw_cas_start()
	//--------------
	//优慕课帐号密码
	GDGM_UMOOC := cas.NewUC(GDGM, "优慕课帐号", "优慕课密码", save_dir)
	if GDGM_UMOOC == nil {
		cas.LogPrintln("优慕课平台设定密码不符合要求----即将退出程序")
		return
	}
	//传入amend，当设定的密码不符合要求时，会从数字工贸CAS登陆平台强制修改密码为设置密码
	GDGM_UMOOC.AutoLogin(true)
	//保存数据
	GDGM_UMOOC.Hwtask(save_dir).Save()
	defer GDGM.SaveCookies()

}
