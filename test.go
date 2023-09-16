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
		save_dir   = cas.GetCurrentAbPath() + "/cas" //数据文件保存路径
		waitsecond = 2
	)
	//数字工贸帐号密码
	GDGM := cas.NewCAS("", "", save_dir, cas.Wtime(int64(waitsecond)), cas.AK(API_KEY), cas.SK(SECRET_KEY))
	defer GDGM.SaveCookies()
	if !GDGM.AutoLogin() {
		cas.LogPrintln("数字工贸登陆失败-----即将退出程序..")
		return
	}
	//-------------
	//提取教务系统课表和成绩
	GDGM.NewJW(save_dir).Jw_cas_start()
	//--------------
	//优慕课帐号密码
	GDGM_UMOOC := cas.NewUC(GDGM, "", "", save_dir)
	if GDGM_UMOOC == nil {
		cas.LogPrintln("优慕课平台设定密码不符合要求----即将退出程序")
		return
	}
	//传入amend，当设定的密码不符合要求时，会从数字工贸CAS登陆平台强制修改密码为设置密码
	GDGM_UMOOC.AutoLogin(true)
	//保存数据
	GDGM_UMOOC.Hwtask(save_dir).Save()
	//一卡通
	card := GDGM.NewCard(save_dir)
	card.Card_cas_start()
	//通过一卡通查询房间信息演示:
	//查询房间菜单
	menu1 := cas.Rooms_menu(card.Search)
	cas.LogPrintln(menu1)
	menu2 := cas.Rooms_menu(card.Search, menu1[0])
	cas.LogPrintln(menu2)
	menu3 := cas.Rooms_menu(card.Search, menu1[0], menu2[0])
	cas.LogPrintln(menu3)
	//参数:工贸白云校区 - 8栋 - 502 => id:
	if res, err := card.Search_elec_Bill(menu1[0], menu2[0], menu3[0]); err != nil && res != nil {
		cas.LogPrintln("查询指定房间信息失败")
	} else {
		cas.LogPrintln("查询:	"+menu1[0], menu2[0], menu3[0])
		bill := res.(cas.RoomBill)
		cas.LogPrintln("查询指定房间信息成功")
		cas.LogPrintln("房间编号:" + bill.RoomNum)
		cas.LogPrintln("热水费:" + bill.FormatHotWaterBalanceStr)
		cas.LogPrintln("水费:" + bill.FormatWaterBalanceStr)
		cas.LogPrintln("电费:" + bill.FormatPowerBalanceStr)
		cas.LogPrintln("更新时间:" + bill.LastDate)
	}
}
