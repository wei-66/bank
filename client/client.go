package client

import (
	"fmt"
	"os"
)

/**
 * @author: SuZhiXiaoWei
 * @DateTime: 2022/4/16 15:25
 **/

func Run() {
	switch os.Args[1] {
	case "login":
		Login()
		break
	case "register":
		Register()
		break
	case "exit":
		Exit()
		break
	case "getMoney":
		GetMoney()
		break
	case "saveMoney":
		SaveMoney()
		break
	case "selectMoney":
		SelectMoney()
	case "help":
		Help()
		break
	default:
		fmt.Println("没有对应的指令，输入--help指令查看详细命令")
		os.Exit(1)
	}
}
