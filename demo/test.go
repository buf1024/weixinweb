package main

import (
	"fmt"
	"os"

	"github.com/buf1024/golib/logging"

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

func setupLog(path string) *logging.Log {
	log, err := logging.NewLogging()
	if err != nil {
		fmt.Printf("NewLogging failed. err = %s\n", err.Error())
		return nil
	}
	_, err = logging.SetupLog("file",
		fmt.Sprintf(`{"prefix":"wx", "filedir":"%s", "level":0, "switchsize":0, "switchtime":0}`,
			path))
	if err != nil {
		fmt.Printf("setup file logger failed. err = %s\n", err.Error())
		return nil
	}
	_, err = logging.SetupLog("console", `{"level":0}`)
	if err != nil {
		fmt.Printf("setup file logger failed. err = %s\n", err.Error())
		return nil
	}
	log.StartSync()
	return log
}

func stopLog(log *logging.Log) {
	log.Stop()
}

func main() {
	os.Mkdir("./log/", 0774)
	log := setupLog("./log/")
	if log == nil {
		fmt.Printf("setup log failed.\n")
		return
	}

	wx := weixinweb.New(log)
	wx.Use(wxWatch)
	qrcode, err := wx.GetQRCode()
	if err != nil {
		log.Error("GetQRCode failed, err = %s.\n", err)
		stopLog(log)
		return
	}
	qrterminal.Generate(qrcode, qrterminal.M, os.Stdout)

	err = wx.WaitForLogin()
	if err != nil {
		log.Error("Loggin failed, err = %s\n", err)
		stopLog(log)
		return
	}
	exitChan := make(chan struct{})
	go wxLoop(wx, exitChan)

	log.Info("wait for exit.\n")
	<-exitChan
	log.Stop()
}
