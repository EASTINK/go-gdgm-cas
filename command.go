package main

// import (
// 	"flag"
// 	"fmt"
// 	"log"
// 	"strconv"
// )

// func main() {
// 	u := flag.String("u", "", "数字工贸帐号")
// 	p := flag.String("p", "", "数字工贸密码")
// 	ak := flag.String("ak", "", "百度云服务API-KEY - Baidu.com ")
// 	sk := flag.String("sk", "", "百度云服务SEC-KEY -Baidu.com")
// 	sp := flag.String("sp", "./", "资源存放路径")
// 	s := flag.Int64("s", 2, "等待时间")
// 	mu := flag.String("mu", "", "优慕课帐号-umooc")
// 	mp := flag.String("mp", "", "优慕课密码-umooc")
// 	mv := flag.Bool("mv", false, "是否自动修正登录")
// 	flag.Parse()
// 	if *u == "" || *p == "" || *sp == "" {
// 		log.Fatalln("缺少必要的登录参数，请检查.")
// 	}
// 	if *sp == "" {
// 		log.Println("设定存储路径为" + *sp)
// 	}
// 	log.Println("设定等待时间为" + strconv.FormatInt(*s, 10) + "秒.")
// 	if *ak == "" || *sk == "" {
// 		log.Println("缺少OCR服务密钥，您可能需要手动输入验证码.")
// 	}
// 	if *mu == "" || *mp == "" {
// 		log.Println("缺少UMOOC登录参数.")
// 		if *mv {
// 			log.Println("检测到修正登录开启,优慕课帐号密码将与数字工贸保持一致.")
// 		}
// 	}
// 	fmt.Println(*u, *p, *ak, *sk, *sp, *s, *mu, *mp, *mv)
// 	//......
// }
