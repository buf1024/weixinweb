package main

import (
	"fmt"
	"os"

	"github.com/buf1024/weixinweb"
	"github.com/mdp/qrterminal"
)

func wxWatch(c *weixinweb.WxContext) {
	fmt.Printf("function wxWatch called\n")
}

func wxLoop(wx *weixinweb.WxWeb, exitChan chan<- struct{}) {
	err := wx.StartWxLoop()
	if err != nil {
		fmt.Printf("StartWxLoop exit, err = %s\n", err)
		exitChan <- struct{}{}
		return
	}

	fmt.Printf("StartWxLoop exit\n")
	exitChan <- struct{}{}
}

func main() {
	wx := weixinweb.New()
	wx.Use(wxWatch)
	qrcode, err := wx.GetQRCode()
	if err != nil {
		fmt.Printf("GetQRCode failed, err = %s.\n", err)
		return
	}
	qrterminal.Generate(qrcode, qrterminal.M, os.Stdout)

	err = wx.WaitForLogin()
	if err != nil {
		fmt.Printf("Loggin failed, err = %s\n", err)
		return
	}
	exitChan := make(chan struct{})
	go wxLoop(wx, exitChan)

	fmt.Printf("wait for exit.\n")
	<-exitChan
}
