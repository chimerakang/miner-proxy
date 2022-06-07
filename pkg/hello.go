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
	fmt.Println("Disclaimer: This tool is for testing and learning purposes only!!")
}

func StringHelp() string {
	myFigure := figure.NewFigure("Miner Proxy", "", true)
	return myFigure.String()
}
