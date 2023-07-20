package main

import (
	"os"
	"stock-bot/src/app"
)

func main() {
	app.App().Run(os.Args)
}
