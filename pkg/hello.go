package pkg

import (
	"fmt"

	"github.com/common-nighthawk/go-figure"
)

func PrintHelp() {
	myFigure := figure.NewFigure("Miner Proxy", "", true)
	myFigure.Print()
	// 免責聲明以及專案地址
	fmt.Println("github repository: https://github.com/chimerakang/miner-proxy")
	fmt.Println("免責聲明: 本工具只適用於測試與學習使用!!")
}

func StringHelp() string {
	myFigure := figure.NewFigure("Miner Proxy", "", true)
	return myFigure.String()
}
