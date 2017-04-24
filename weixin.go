package weixinweb

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/buf1024/golib/logging"
)

var debugFlag = true
var log *logging.Log

func debugStart() {
	if debugFlag {
		var err error
		log, err = logging.NewLogging()
		if err != nil {
			fmt.Printf("NewLogging failed. err = %s\n", err.Error())
			return
		}
		_, err = logging.SetupLog("file",
			`{"prefix":"wx", "filedir":"./log/", "level":0, "switchsize":0, "switchtime":0}`)
		if err != nil {
			fmt.Printf("setup file logger failed. err = %s\n", err.Error())
			return
		}
		_, err = logging.SetupLog("console", `{"level":0}`)
		if err != nil {
			fmt.Printf("setup file logger failed. err = %s\n", err.Error())
			return
		}
		log.StartSync()
		log.Debug("log ready\n")
	}
}
func debugStop() {
	if debugFlag {
		log.Stop()
	}
}

func debug(format string, a ...interface{}) {
	if debugFlag {
		log.Debug(format, a...)
	}
}

const (
	appid       = "wx782c26e4c19acffb"
	userAgent   = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"
	referer     = "https://wx2.qq.com/?&lang=zh_CN"
	jsonType    = "application/json; charset=UTF-8"
	lang        = "zh_CN"
	fun         = "new"
	sysInterval = 25
)

var (
	syncHosts = []string{
		"wx2.qq.com",
		"webpush.wx2.qq.com",
		"wx8.qq.com",
		"webpush.wx8.qq.com",
		"qq.com",
		"webpush.wx.qq.com",
		"web2.wechat.com",
		"webpush.web2.wechat.com",
		"webpush.web.wechat.com",
		"webpush.weixin.qq.com",
		"webpush.wechat.com",
		"webpush1.wechat.com",
		"webpush2.wechat.com",
		"webpush.wx.qq.com",
		"webpush2.wx.qq.com",
	}
	specialUsers = []string{
		"newsapp", "fmessage", "filehelper", "weibo", "qqmail",
		"fmessage", "tmessage", "qmessage", "qqsync", "floatbottle",
		"lbsapp", "shakeapp", "medianote", "qqfriend", "readerapp",
		"blogapp", "facebookapp", "masssendapp", "meishiapp", "feedsapp",
		"voip", "blogappweixin", "weixin", "brandsessionholder", "weixinreminder",
		"wxid_novlwrv3lqwv11", "gh_22b87fa7cb3c", "officialaccounts",
		"notification_messages", "wxid_novlwrv3lqwv11", "gh_22b87fa7cb3c", "wxitil", "userexperience_alarm",
		"notification_messages",
	}
)

type WxWeb struct {
	deviceID string

	uuid        string
	redirectURI string
	baseURL     string

	skey       string
	wxsid      string
	wxuin      string
	passTicket string

	exSyncKey wxSyncKey
	synckey   string

	userSelf wxUser
	syncHost string

	cookies []*http.Cookie

	handlers WxHandlerChain
}
type WxContext struct {
}
type WxHandler func(c *WxContext)
type WxHandlerChain []WxHandler

func New() *WxWeb {
	w := &WxWeb{}
	w.deviceID = w.newDeiviceID()

	debugStart()

	return w
}

func (w *WxWeb) newDeiviceID() string {
	result := ""
	for i := 0; i < 15; i++ {
		result = result + fmt.Sprintf("%d", rand.Intn(10))
	}
	return "e" + result

}
func (w *WxWeb) timeNow() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/1000000)

}
func (w *WxWeb) timeNowInt64() int64 {
	return time.Now().UnixNano() / 1000000

}
func (w *WxWeb) post(reqURL string, data []byte, jsData bool, cookies []*http.Cookie, saveCookie bool) ([]byte, error) {
	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)
	if jsData {
		req.Header.Add("Content-Type", jsonType)
	}

	debug("http post req ->\nurl:%s\ndata=%s\n", reqURL, string(data))
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	if cookies != nil {
		jar, _ := cookiejar.New(nil)
		reqURL, _ := url.Parse(reqURL)
		jar.SetCookies(reqURL, cookies)
		client.Jar = jar
	}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	if saveCookie {
		w.cookies = rsp.Cookies()
	}

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	debug("http post rsp ->\nurl:%s\ndata=%s\n", reqURL, string(bytes))
	return bytes, nil
}
func (w *WxWeb) get(reqURL string, data []byte, cookies []*http.Cookie, saveCookie bool) ([]byte, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)

	debug("http get req ->\nurl:%s\ndata=%s\n", reqURL, string(data))
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	if cookies != nil {
		jar, _ := cookiejar.New(nil)
		reqURL, _ := url.Parse(reqURL)
		jar.SetCookies(reqURL, cookies)
		client.Jar = jar
	}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	if saveCookie {
		w.cookies = rsp.Cookies()
	}
	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	debug("http get rsp ->\nurl:%s\ndata=%s\n", reqURL, string(bytes))
	return bytes, nil
}
func (w *WxWeb) Use(handlers ...WxHandler) *WxWeb {
	if w.handlers == nil {
		w.handlers = handlers[:]
		return w
	}
	size := len(w.handlers) + len(handlers)
	newHandlers := make(WxHandlerChain, size)
	copy(newHandlers, w.handlers)
	copy(newHandlers[len(w.handlers):], handlers)
	w.handlers = newHandlers
	return w
}

func (w *WxWeb) WaitForLogin() error {
	debug("WaitForLogin\n")
	if err := w.Login(1); err != nil {
		return err
	}
	if err := w.Login(0); err != nil {
		return err
	}
	return nil
}

func (w *WxWeb) StartWxLoop() error {
	debug("StartWxLoop\n")
	if err := w.NewLoginPage(); err != nil {
		return err
	}
	if err := w.WxInit(); err != nil {
		return err
	}
	if err := w.StatusNotify(); err != nil {
		return err
	}
	last := time.Now().Unix()
	for {
		now := time.Now().Unix()
		ret, sel := w.SyncCheck()
		debug("retcode=%d, selector=%d\n", ret, sel)
		if ret != 0 {
			break
		}
		if sel == 2 {
			err := w.Sync()
			if err != err {
				return err
			}
			for _, handler := range w.handlers {
				ctx := &WxContext{}
				handler(ctx)
			}
			continue
		}
		sleep := now - last
		if sleep < sysInterval {
			time.Sleep(time.Second * time.Duration(sleep))
		}
		last = now
	}

	return nil
}

func (w *WxWeb) GetQRCode() (string, error) {
	debug("GetQRCode\n")
	uri := "https://login.weixin.qq.com/jslogin?"

	v := url.Values{}
	v.Add("appid", appid)
	v.Add("fun", fun)
	v.Add("lang", lang)
	v.Add("_", w.timeNow())
	uri = uri + v.Encode()

	bytes, err := w.get(uri, nil, nil, false)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("window.QRLogin.code = (\\d+); window.QRLogin.uuid = \"(\\S+?)\"")

	match := re.FindStringSubmatch(string(bytes))
	if len(match) < 3 || match[1] != "200" {
		return "", fmt.Errorf("unexpect ret, ret = %s", string(bytes))
	}
	w.uuid = match[2]

	qruri := "https://login.weixin.qq.com/l/" + w.uuid

	return qruri, nil

}

func (w *WxWeb) Login(tip int) error {
	debug("Login tip = %d\n", tip)
	uri := "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?"

	v := url.Values{}
	v.Add("tip", fmt.Sprintf("%d", tip))
	v.Add("uuid", w.uuid)
	v.Add("_", w.timeNow())
	uri = uri + v.Encode()

	bytes, err := w.get(uri, nil, nil, false)
	if err != nil {
		return nil
	}
	re := regexp.MustCompile("window.code=(\\d+);")
	match := re.FindStringSubmatch(string(bytes))
	if len(match) < 2 {
		return fmt.Errorf("unexpect ret, ret = %s", string(bytes))
	}
	if match[1] == "201" {
		// 已经扫描
		return nil
	} else if match[1] == "200" {
		re = regexp.MustCompile("window.redirect_uri=\"(\\S+?)\"")

		match = re.FindStringSubmatch(string(bytes))
		if len(match) < 2 {
			return fmt.Errorf("unexpect ret, ret = %s", string(bytes))
		}
		w.redirectURI = match[1] + "&fun=" + fun
		_, err := url.Parse(w.redirectURI)
		if err != nil {
			return err
		}
		w.baseURL = w.redirectURI[:strings.LastIndex(w.redirectURI, "/")]

		debug("Login redicect = %s baseurl = %s\n", w.redirectURI, w.baseURL)

	} else if match[1] == "408" {
		return fmt.Errorf("timeout, ret = %s", string(bytes))
	} else {
		return fmt.Errorf("unexpect ret, ret = %s", string(bytes))
	}
	return nil

}
func (w *WxWeb) NewLoginPage() error {
	debug("NewLoginPage\n")
	bytes, err := w.get(w.redirectURI, nil, nil, true)
	if err != nil {
		return nil
	}
	rst := &wxLoginPageRsp{}
	err = xml.Unmarshal(bytes, rst)
	if err != nil {
		return err
	}

	if rst.Ret != 0 {
		return fmt.Errorf("unexpect NewLoginPage result ,rsp = %s", string(bytes))
	}
	w.skey = rst.Skey
	w.wxsid = rst.Wxsid
	w.wxuin = rst.Wxuin
	w.passTicket = rst.PassTicket

	return nil

}
func (w *WxWeb) WxInit() error {
	debug("WxInit\n")
	uri := fmt.Sprintf("%s/webwxinit?r=%d&pass_ticket=%s&skey=%s",
		w.baseURL, w.timeNow(), w.passTicket, w.skey)

	jsReq := wxInitReq{
		BaseRequest: wxBaseRequest{
			Uin:      w.wxuin,
			Sid:      w.wxsid,
			Skey:     w.skey,
			DeviceID: w.deviceID,
		},
	}
	jsByte, err := json.Marshal(jsReq)
	if err != nil {
		return err
	}
	bytes, err := w.post(uri, jsByte, true, w.cookies, false)
	if err != nil {
		return nil
	}

	var jsRsp wxInitRsp
	err = json.Unmarshal(bytes, &jsRsp)
	if err != nil {
		return err
	}

	if jsRsp.BaseResponse.Ret != 0 {
		return fmt.Errorf("BaseResponse.Ret != 0, err = %s", string(bytes))
	}
	w.exSyncKey = jsRsp.SyncKey
	keys := make([]string, len(jsRsp.SyncKey.List))
	for i, key := range jsRsp.SyncKey.List {
		keys[i] = fmt.Sprintf("%d_%d", key.Key, key.Val)
	}
	w.synckey = strings.Join(keys, "|")
	w.userSelf = jsRsp.User

	debug("syn key=%s, %v\n", w.synckey, jsRsp.SyncKey)

	return nil

}
func (w *WxWeb) StatusNotify() error {
	debug("StatusNotify\n")
	uri := fmt.Sprintf("%s/webwxstatusnotify?lang=%s&pass_ticket=%s",
		w.baseURL, lang, w.passTicket)

	jsReq := wxStatusNotifyReq{
		BaseRequest: wxBaseRequest{
			Uin:      w.wxuin,
			Sid:      w.wxsid,
			Skey:     w.skey,
			DeviceID: w.deviceID,
		},
		Code:         3,
		FromUserName: w.userSelf.UserName,
		ToUserName:   w.userSelf.UserName,
		ClientMsgId:  w.timeNowInt64(),
	}
	jsByte, err := json.Marshal(jsReq)
	if err != nil {
		return err
	}
	bytes, err := w.post(uri, jsByte, true, w.cookies, false)
	if err != nil {
		return nil
	}

	var jsRsp wxStatusNotifyRsp
	err = json.Unmarshal(bytes, &jsRsp)
	if err != nil {
		return err
	}

	if jsRsp.BaseResponse.Ret != 0 {
		return fmt.Errorf("BaseResponse.Ret != 0, err = %s", string(bytes))
	}
	return nil

}
func (w *WxWeb) GetContact() error {
	debug("GetContact\n")
	uri := fmt.Sprintf("%s/webwxgetcontact?lang=%s&pass_ticket=%s&skey=%s&seq=0&r=%s",
		w.baseURL, lang, w.passTicket, w.skey, w.timeNow())

	jsReq := wxGetContactReq{
		BaseRequest: wxBaseRequest{
			Uin:      w.wxuin,
			Sid:      w.wxsid,
			Skey:     w.skey,
			DeviceID: w.deviceID,
		},
	}
	jsByte, err := json.Marshal(jsReq)
	if err != nil {
		return err
	}
	_, err = w.post(uri, jsByte, true, w.cookies, false)
	if err != nil {
		return nil
	}
	// TODO

	return nil

}
func (w *WxWeb) BatchGetContact() error {
	debug("BatchGetContact\n")
	_ = fmt.Sprintf("%s/webwxbatchgetcontact?type=ex&pass_ticket=%s&r=%s",
		w.baseURL, w.passTicket, w.timeNow())
	/* {
	     BaseRequest: { Uin: xxx, Sid: xxx, Skey: xxx, DeviceID: xxx },
	     Count: 群数量,
	     List: [
	         { UserName: 群ID, EncryChatRoomId: "" },
	         ...
	     ],
	}*/
	return nil

}

func (w *WxWeb) syncCheck(host string) (int, int) {
	uri := "https://" + host + "/cgi-bin/mmwebwx-bin/synccheck?"
	v := url.Values{}
	v.Add("r", w.timeNow())
	v.Add("sid", w.wxsid)
	v.Add("uin", w.wxuin)
	v.Add("skey", w.skey)
	v.Add("deviceid", w.deviceID)
	v.Add("synckey", w.synckey)
	v.Add("_", w.timeNow())
	uri = uri + v.Encode()

	bytes, err := w.get(uri, nil, w.cookies, false)
	if err != nil {
		return -1, -1
	}

	re := regexp.MustCompile("window.synccheck={retcode:\"(\\d+)\",selector:\"(\\d+)\"}")
	match := re.FindStringSubmatch(string(bytes))
	if len(match) < 3 {
		return -1, -1
	}
	retcode, err := strconv.Atoi(match[1])
	if err != nil {
		return -1, -1
	}
	selector, err := strconv.Atoi(match[2])
	if err != nil {
		return -1, -1
	}
	return retcode, selector
}

func (w *WxWeb) SyncCheck() (int, int) {
	debug("SyncCheck\n")
	if len(w.syncHost) == 0 {
		for _, host := range syncHosts {
			retcode, selector := w.syncCheck(host)
			if retcode == 0 && selector >= 0 {
				w.syncHost = host
				return retcode, selector
			}
		}
		return -1, -1
	}
	retcode, selector := w.syncCheck(w.syncHost)

	if retcode >= 0 && selector >= 0 {
		return retcode, selector
	}
	w.syncHost = ""
	return w.SyncCheck()

}
func (w *WxWeb) Sync() error {
	debug("Sync\n")
	uri := fmt.Sprintf("%s/webwxsync?sid=%s&skey=%s&pass_ticket=%s",
		w.baseURL, w.wxsid, w.skey, w.passTicket)

	jsReq := wxSyncReq{
		BaseRequest: &wxBaseRequest{
			Uin:      w.wxuin,
			Sid:      w.wxsid,
			Skey:     w.skey,
			DeviceID: w.deviceID,
		},
		SyncKey: w.exSyncKey,
		RR:      ^w.timeNowInt64(),
	}
	jsByte, err := json.Marshal(jsReq)
	if err != nil {
		return err
	}
	bytes, err := w.post(uri, jsByte, true, w.cookies, false)
	if err != nil {
		return nil
	}
	var jsRsp wxSyncRsp
	err = json.Unmarshal(bytes, &jsRsp)
	if err != nil {
		return err
	}

	if jsRsp.BaseResponse.Ret != 0 {
		return fmt.Errorf("retcode != 0, err = %s", string(bytes))
	}

	w.exSyncKey = jsRsp.SyncKey
	keys := make([]string, len(jsRsp.SyncKey.List))
	for i, key := range jsRsp.SyncKey.List {
		keys[i] = fmt.Sprintf("%d_%d", key.Key, key.Val)
	}
	w.synckey = strings.Join(keys, "|")

	debug("syn key=%s, %v\n", w.synckey, jsRsp.SyncKey)

	// TODO Msg

	return nil

}

func (w *WxWeb) SendMsg() error {
	debug("SendMsg\n")
	uri := fmt.Sprintf("%s/webwxstatusnotify?lang=%s&pass_ticket=%s",
		w.baseURL, lang, w.passTicket)

	jsReq := wxSendMsgReq{
		BaseRequest: wxBaseRequest{
			Uin:      w.wxuin,
			Sid:      w.wxsid,
			Skey:     w.skey,
			DeviceID: w.deviceID,
		},
	}
	jsByte, err := json.Marshal(jsReq)
	if err != nil {
		return err
	}
	fmt.Printf("uri = %s, js = %s\n", uri, string(jsByte))

	req, err := http.NewRequest("POST", uri, bytes.NewReader(jsByte))
	if err != nil {
		return err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)
	req.Header.Add("Content-Type", jsonType)

	fmt.Printf("req %s\n", uri)
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer rsp.Body.Close()

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil
	}
	f, _ := os.OpenFile("/home/heidong/statusnotify.txt", os.O_CREATE|os.O_RDWR|os.O_SYNC, 0666)
	f.WriteString(string(bytes))
	f.Close()

	var jsRsp wxSendMsgRsp
	err = json.Unmarshal(bytes, &jsRsp)
	if err != nil {
		fmt.Printf("Unmarshal ERROR, err = %s\n", err)
		return err
	}

	if jsRsp.BaseResponse.Ret != 0 {
		return fmt.Errorf("retcode != 0, err = %s", string(bytes))
	}
	return nil

}
func (w *WxWeb) RevokeMsg() error {
	/*
			https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxrevokemsg
			{
		     BaseRequest: { Uin: xxx, Sid: xxx, Skey: xxx, DeviceID: xxx },
		     SvrMsgId: msg_id,
		     ToUserName: user_id,
		     ClientMsgId: local_msg_id
		}*/
	return nil

}
func (w *WxWeb) SendMsgEmotion() error {
	/*
			https://wx2.qq.com/cgi-bin/mmwebwx-bin/webwxsendemoticon?fun=sys&f=json&pass_ticket=xxx
			{
		     BaseRequest: { Uin: xxx, Sid: xxx, Skey: xxx, DeviceID: xxx },
		     Msg: {
		         Type: 47 emoji消息,
		         EmojiFlag: 2,
		         MediaId: 表情上传后的媒体ID,
		         FromUserName: 自己ID,
		         ToUserName: 好友ID,
		         LocalID: 与clientMsgId相同,
		         ClientMsgId: 时间戳左移4位随后补上4位随机数
		     }
		}*/
	return nil

}

func (w *WxWeb) GetIcon() error {
	/*https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgeticon
		 GET
	params	seq: 数字，可为空
	username: ID
	skey: xxx
	*/
	return nil

}
func (w *WxWeb) GetHeadImg() error {
	/*
			url	https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetheadimg
		method	GET
		params	seq: 数字，可为空
		username: 群ID
		skey: xxx*/
	return nil

}
func (w *WxWeb) GetMsgImg() error {
	/*
			url	https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetmsgimg
		method	GET
		params	MsgID: 消息ID
		type: slave 略缩图 or 为空时加载原图
		skey: xxx
	*/
	return nil

}
func (w *WxWeb) GetVideo() error {
	/*
			url	https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetvideo
		method	GET
		params	msgid: 消息ID
		skey: xxx
	*/
	return nil

}
func (w *WxWeb) GetVoice() error {
	/*
			url	https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetvoice
		method	GET
		params	msgid: 消息ID
		skey: xxx
	*/
	return nil

}

func init() {
	rand.NewSource(time.Now().Unix())
}
