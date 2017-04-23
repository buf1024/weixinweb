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
)

const (
	appid     = "wx782c26e4c19acffb"
	userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"
	referer   = "https://wx2.qq.com/?&lang=zh_CN"
	jsonType  = "application/json; charset=UTF-8"
	lang      = "zh_CN"
	fun       = "new"
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

	userSelf user
	syncHost string

	cookies []*http.Cookie
}
type WxContext struct {
}
type WxHandler func(c *WxContext) error
type WxHandlerChain []WxHandler

func newDeiviceID() string {
	result := ""
	for i := 0; i < 15; i++ {
		result = result + fmt.Sprintf("%d", rand.Intn(10))
	}
	return "e" + result

}

func New() *WxWeb {
	w := &WxWeb{
		deviceID: newDeiviceID(),
	}

	return w
}
func (w *WxWeb) timeNow() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/1000000)

}
func (w *WxWeb) timeNowInt64() int64 {
	return time.Now().UnixNano() / 1000000

}

func (w *WxWeb) Start() error {
	return nil

}
func (w *WxWeb) GetUUID() (string, error) {
	uri := "https://login.weixin.qq.com/jslogin?"

	v := url.Values{}
	v.Add("appid", appid)
	v.Add("fun", fun)
	v.Add("lang", lang)
	v.Add("_", w.timeNow())

	uri = uri + v.Encode()

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-agent", userAgent)

	fmt.Printf("req %s\n", uri)
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer rsp.Body.Close()

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", nil
	}

	re := regexp.MustCompile("window.QRLogin.code = (\\d+); window.QRLogin.uuid = \"(\\S+?)\"")

	match := re.FindStringSubmatch(string(bytes))
	if len(match) < 3 || match[1] != "200" {
		return "", fmt.Errorf("not expect ret, ret = %s", string(bytes))
	}
	w.uuid = match[2]

	qruri := "https://login.weixin.qq.com/l/" + w.uuid

	fmt.Printf("rsp %s, uuid = %s, rquri = %s\n", string(bytes), w.uuid, qruri)

	return qruri, nil

}

func (w *WxWeb) Login(tip int) error {
	uri := "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?"

	v := url.Values{}
	v.Add("tip", fmt.Sprintf("%d", tip))
	v.Add("uuid", w.uuid)
	v.Add("_", w.timeNow())

	uri = uri + v.Encode()

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-agent", userAgent)

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
	fmt.Printf("rsp %s\n", string(bytes))
	re := regexp.MustCompile("window.code=(\\d+);")
	match := re.FindStringSubmatch(string(bytes))
	fmt.Printf("MATCH %v\n", match)
	if len(match) < 2 {
		return fmt.Errorf("not expect ret, ret = %s", string(bytes))
	}
	fmt.Printf("RET = %s\n", match[1])
	if match[1] == "201" {
		// 已经扫描
		return nil
	} else if match[1] == "200" {
		re = regexp.MustCompile("window.redirect_uri=\"(\\S+?)\"")

		match = re.FindStringSubmatch(string(bytes))
		if len(match) < 2 {
			return fmt.Errorf("not expect ret, ret = %s", string(bytes))
		}
		w.redirectURI = match[1] + "&fun=" + fun
		_, err := url.Parse(w.redirectURI)
		if err != nil {
			return err
		}
		w.baseURL = w.redirectURI[:strings.LastIndex(w.redirectURI, "/")]

		fmt.Printf("redicect = %s baseurl = %s\n", w.redirectURI, w.baseURL)

	} else if match[1] == "408" {
		return fmt.Errorf("timeout, ret = %s", string(bytes))
	} else {
		return fmt.Errorf("not expect ret, ret = %s", string(bytes))
	}
	return nil

}
func (w *WxWeb) NewLoginPage() error {
	req, err := http.NewRequest("GET", w.redirectURI, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)

	fmt.Printf("req %s\n", w.redirectURI)
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
	fmt.Printf("rsp %s\n", string(bytes))

	type newLoginRst struct {
		XMLName     xml.Name `xml:"error"`
		Ret         int      `xml:"ret"`
		Message     string   `xml:"message"`
		Skey        string   `xml:"skey"`
		Wxsid       string   `xml:"wxsid"`
		Wxuin       string   `xml:"wxuin"`
		PassTicket  string   `xml:"pass_ticket"`
		IsGrayscale int      `xml:"isgrayscale"`
	}
	rst := &newLoginRst{}
	err = xml.Unmarshal(bytes, rst)
	if err != nil {
		return err
	}
	if rst.Ret != 0 {
		return fmt.Errorf("rsp failed ,rst = %s", string(bytes))
	}
	w.skey = rst.Skey
	w.wxsid = rst.Wxsid
	w.wxuin = rst.Wxuin
	w.passTicket = rst.PassTicket

	w.cookies = rsp.Cookies()

	fmt.Printf("login rst: %v", *w)

	return nil

}
func (w *WxWeb) WxInit() error {
	uri := fmt.Sprintf("%s/webwxinit?r=%d&pass_ticket=%s&skey=%s",
		w.baseURL, w.timeNow(), w.passTicket, w.skey)

	jsReq := wxReq{
		BaseRequest: &baseRequest{
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
	//fmt.Printf("rsp %s\n", string(bytes))

	var jsRsp wxRsp
	err = json.Unmarshal(bytes, &jsRsp)
	if err != nil {
		fmt.Printf("Unmarshal ERROR, err = %s\n", err)
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
	w.userSelf = jsRsp.User

	fmt.Printf("syn key=%s, %v\n", w.synckey, jsRsp.SyncKey)

	f, _ := os.OpenFile("/home/heidong/wxinit.txt", os.O_CREATE|os.O_RDWR|os.O_SYNC, 0666)
	f.WriteString(string(bytes))
	f.Close()

	return nil

}
func (w *WxWeb) StatusNotify() error {
	uri := fmt.Sprintf("%s/webwxstatusnotify?lang=%s&pass_ticket=%s",
		w.baseURL, lang, w.passTicket)

	jsReq := wxStatusNotifyReq{
		BaseRequest: &baseRequest{
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

	var jsRsp wxRsp
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
func (w *WxWeb) GetContact() error {
	uri := fmt.Sprintf("%s/webwxgetcontact?lang=%s&pass_ticket=%s&skey=%s&seq=0&r=%s",
		w.baseURL, lang, w.passTicket, w.skey, w.timeNow())

	jsReq := wxReq{
		BaseRequest: &baseRequest{
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

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsByte))
	if err != nil {
		return err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)
	req.Header.Add("Content-Type", jsonType)

	fmt.Printf("getcontact req %s\n", uri)
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
	//fmt.Printf("rsp %s\n", string(bytes))
	f, _ := os.OpenFile("/home/heidong/getcontact.txt", os.O_CREATE|os.O_RDWR|os.O_SYNC, 0666)
	f.WriteString(string(bytes))
	f.Close()

	return nil

}
func (w *WxWeb) BatchGetContact() error {
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

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return -1, -1
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)

	fmt.Printf("sync check req %s\n", uri)
	jar, _ := cookiejar.New(nil)
	reqUri, _ := url.Parse(uri)
	jar.SetCookies(reqUri, w.cookies)
	client := &http.Client{Jar: jar}
	rsp, err := client.Do(req)
	if err != nil {
		return -1, -1
	}
	defer rsp.Body.Close()

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return -1, -1
	}
	fmt.Printf("rsp %s\n", string(bytes))

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
	uri := fmt.Sprintf("%s/webwxsync?sid=%s&skey=%s&pass_ticket=%s",
		w.baseURL, w.wxsid, w.skey, w.passTicket)

	jsReq := wxSyncReq{
		BaseRequest: &baseRequest{
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
	fmt.Printf("uri = %s, js = %s\n", uri, string(jsByte))

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsByte))
	if err != nil {
		return err
	}
	req.Header.Add("User-agent", userAgent)
	req.Header.Add("Referer", referer)
	req.Header.Add("Content-Type", jsonType)

	fmt.Printf("Sync req %s\n", uri)
	jar, _ := cookiejar.New(nil)
	reqUri, _ := url.Parse(uri)
	jar.SetCookies(reqUri, w.cookies)
	client := &http.Client{Jar: jar}
	rsp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer rsp.Body.Close()

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil
	}

	fmt.Printf("rsp %s\n", string(bytes))
	return nil

}

func (w *WxWeb) SendMsg() error {
	return nil

}
func (w *WxWeb) RevokeMsg() error {
	return nil

}
func (w *WxWeb) SendMsgEmotion() error {
	return nil

}

func (w *WxWeb) GetIcon() error {
	return nil

}
func (w *WxWeb) GetHeadImg() error {
	return nil

}
func (w *WxWeb) GetMsgImg() error {
	return nil

}
func (w *WxWeb) GetVideo() error {
	return nil

}
func (w *WxWeb) GetVoice() error {
	return nil

}

func init() {
	rand.NewSource(time.Now().Unix())
}
