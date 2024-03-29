package cas

import "encoding/json"

func (c *Card) logon() bool {
	title, err := c.CAS.page.Info()
	if err == nil {
		return title.Title == "服务中心"
	} else {
		LogPrintln("检查登陆状态出错")
	}
	return false
}

func (c *Card) card_sso() bool {
	//第一步跳转amp
	err := c.CAS.page.Navigate(cas_card)
	if err != nil {
		return false
	}
	c.CAS.cas_wait()
	//检查进程
	if c.logon() {
		return true
	}
	c.CAS.cas_wait()
	//单击登陆第二步跳转
	c.CAS.page.MustElement("#pc > div.login > div.rightLogo > div.flexLogin > div:nth-child(2) > a").MustClick()
	c.CAS.cas_wait()
	//返回跳转结果
	return c.logon()
}

func (c *Card) Card_cas_start() {
	if c.card_sso() {
		LogPrintln("一卡通登陆成功")
		c.card_save_rooms()
	} else {
		LogPrintln("一卡通登陆失败")
	}
}

func (c *Card) card_save_rooms() {
	var bylw_roomlist RoomList
	var th_roomlist RoomList
	c.Search = make(map[string]map[string]map[string]string)
	bylw_byte := c.Post(card_room_bylw_list, "")
	json.Unmarshal(bylw_byte, &bylw_roomlist)
	if bylw_roomlist.Ret {
		for _, v := range bylw_roomlist.RoomList {
			c.adds(c.Search, v.SchoolArea, v.Building, v.RoomName, v.RoomId)
		}
	}
	th_byte := c.Post(card_room_th_list, "")
	json.Unmarshal(th_byte, &th_roomlist)
	if th_roomlist.Ret {
		for _, v := range th_roomlist.RoomList {
			c.adds(c.Search, v.SchoolArea, v.Building, v.RoomName, v.RoomId)
		}
	}
	savejson(c.Search, c.Savedir+"/rooms.json", "房间列表已保存", "房间列表保存失败")
}

func (c *Card) Post(url string, body string) []byte {
	c.CAS.cas_wait()
	command := `fetch("` + url + `", {
		"headers": {
			"accept": "*/*",
			"accept-language": "zh-CN,zh;q=0.9,en;q=0.8",
			"content-type": "application/x-www-form-urlencoded; charset=UTF-8",
			"sec-ch-ua": "\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Google Chrome\";v=\"116\"",
			"sec-ch-ua-mobile": "?0",
			"sec-ch-ua-platform": "\"Linux\"",
			"sec-fetch-dest": "empty",
			"sec-fetch-mode": "cors",
			"sec-fetch-site": "same-origin",
			"x-requested-with": "XMLHttpRequest"
		},
		"referrer": "https://carduser.gdgm.cn/powerfee/index",
		"referrerPolicy": "strict-origin-when-cross-origin",
		"body": "` + body + `",
		"method": "POST",
		"mode": "cors",
		"credentials": "include"
	  })  .then((response) => response.json())
		.then((data) => document.querySelector("html").innerHTML=JSON.stringify(data));`
	c.CAS.page.MustEval("()=>" + command)
	c.CAS.cas_wait()
	return []byte(c.CAS.page.MustElement("body").MustText())
}

func (c *Card) adds(item map[string]map[string]map[string]string, schoolArea string, building string, room string, roomNum string) {
	if schoolArea == "" {
		schoolArea = "工贸天河校区"
	}
	if _, ok := item[schoolArea]; !ok {
		item[schoolArea] = make(map[string]map[string]string)
	}

	if _, ok := item[schoolArea][building]; !ok {
		item[schoolArea][building] = make(map[string]string)
	}

	if _, ok := item[schoolArea][building][room]; !ok {
		item[schoolArea][building][room] = roomNum
	}

}

/*
查询房间电费

params: 校区|楼栋|房间号
返回:

	Bill{str(balance,water,hotwater,lastdate)}
	其中水费和热水费没有实际传回数据,lastdate只有天河校区有数据
*/
func (c *Card) Search_elec_Bill(schoolArea string, building string, roomno string) (any, error) {
	//从map取回roomid
	id := c.Search[schoolArea][building][roomno]
	//设置body
	var body string
	if schoolArea != "工贸天河校区" {
		body = "implType=" + bylw_flag
	} else {
		body = "implType=" + th_flag
	}
	body += "&roomNum=" + id
	//查询房间信息
	ele_byte := c.Post(card_room_Balance, body)
	var res Roominfo
	err := json.Unmarshal(ele_byte, &res)
	if err != nil {
		return nil, err
	}
	if !res.Ret {
		return nil, err
	}
	return res.Bill, nil
}
