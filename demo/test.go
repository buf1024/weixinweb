package main

import (
	"os"

	"fmt"
	"time"

	"github.com/buf1024/weixinweb"
	"github.com/mdp/qrterminal"
)

func main() {
	w := weixinweb.New()
	url, _ := w.GetUUID()
	qrterminal.Generate(url, qrterminal.M, os.Stdout)
	w.Login(1)
	w.Login(0)

	w.NewLoginPage()
	w.WxInit()
	w.StatusNotify()
	w.GetContact()

	for {
		r, s := w.SyncCheck()
		fmt.Printf("synccheck now = %d, r=%d, s=%d\n", time.Now().Unix(), r, s)
		w.Sync()
		time.Sleep(time.Second * 20)
	}

}
