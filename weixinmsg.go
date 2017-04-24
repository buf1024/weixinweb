package weixinweb

import "encoding/xml"

const (
	MsgText       = 1     // 文本消息
	MsgPicture    = 3     // 图片消息
	MsgVoice      = 34    // 语音消息
	MsgConfirm    = 37    // 好友确认消息
	MsgFrind      = 40    // POSSIBLEFRIEND_MSG
	MsgShareCard  = 42    // 共享名片
	MsgVideo      = 43    // 视频消息
	MsgEmotion    = 47    // 动画表情
	MsgLocation   = 48    // 位置消息
	MsgShareLink  = 49    // 分享链接
	MsgVOIP       = 50    // VOIPMSG
	MsgWeixinInit = 51    // 微信初始化消息
	MsgVOIPNotify = 52    // VOIPNOTIFY
	MsgVOIPInvite = 53    // VOIPINVITE
	MsgShortVideo = 62    // 小视频
	MsgSYSNotice  = 9999  // SYSNOTICE
	MsgSYS        = 10000 // 系统消息
	MsgRevoke     = 10002 // 撤回消息
)

type wxBaseRequest struct {
	Uin      string
	Sid      string
	Skey     string
	DeviceID string
}

type wxBaseResponse struct {
	Ret    int
	ErrMsg string
}

type wxMember struct {
	Uin             int
	UserName        string
	NickName        string
	AttrStatus      int
	PYInitial       string
	PYQuanPin       string
	RemarkPYInitial string
	RemarkPYQuanPin string
	MemberStatus    int
	DisplayName     string
	KeyWord         string
}

type wxUser struct {
	Uin               int
	UserName          string
	NickName          string
	HeadImgUrl        string
	ContactFlag       int
	MemberCount       int
	MemberList        []wxMember
	RemarkName        string
	HideInputBarFlag  int
	Sex               int
	Signature         string
	VerifyFlag        int
	OwnerUin          int
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	StarFriend        int
	AppAccountFlag    int
	Statues           int
	AttrStatus        int
	Province          string
	City              string
	Alias             string
	SnsFlag           int
	UniFriend         int
	DisplayName       string
	ChatRoomId        int
	KeyWord           string
	EncryChatRoomId   string
	IsOwner           int
	WebWxPluginSwitch int
	HeadImgFlag       int
}

type wxSyncKeyVal struct {
	Key int
	Val int64
}

type wxSyncKey struct {
	Count int
	List  []wxSyncKeyVal
}

type wxArticle struct {
	Title  string
	Digest string
	Cover  string
	Url    string
}
type wxSubscribeMsg struct {
	UserName       string
	MPArticleCount int
	MPArticleList  []wxArticle
	Time           int64
	NickName       string
}

//////////////////

type wxLoginPageRsp struct {
	XMLName     xml.Name `xml:"error"`
	Ret         int      `xml:"ret"`
	Message     string   `xml:"message"`
	Skey        string   `xml:"skey"`
	Wxsid       string   `xml:"wxsid"`
	Wxuin       string   `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayscale int      `xml:"isgrayscale"`
}

type wxInitReq struct {
	BaseRequest wxBaseRequest
}

type wxInitRsp struct {
	BaseResponse        wxBaseResponse
	Count               int
	ContactList         []wxUser
	SyncKey             wxSyncKey
	User                wxUser
	ChatSet             string
	SKey                string
	ClientVersion       int64
	SystemTime          int64
	GrayScale           int
	InviteStartCount    int
	MPSubscribeMsgCount int
	MPSubscribeMsgList  []wxSubscribeMsg
	ClickReportInterval int
	MsgID               string
}

type wxStatusNotifyReq struct {
	BaseRequest  wxBaseRequest
	Code         int
	FromUserName string
	ToUserName   string
	ClientMsgId  int64
}
type wxStatusNotifyRsp struct {
	wxInitRsp
}

type wxGetContactReq struct {
	BaseRequest wxBaseRequest
}
type wxGetContactRsp struct {
	wxInitRsp
}
type wxGetBatchContactReq struct {
	BaseRequest wxBaseRequest
	Count       int
	List        []struct {
		UserName        string
		EncryChatRoomId string
	}
}
type wxGetBatchContactRsp struct {
	wxInitRsp
}

type wxSyncReq struct {
	BaseRequest *wxBaseRequest
	SyncKey     wxSyncKey
	RR          int64 `json:"rr"`
}

type wxSyncCheckReq struct {
	BaseRequest  *wxBaseRequest
	Code         int
	FromUserName string
	ToUserName   string
	ClientMsgId  int64
}

type wxSyncRsp struct {
	BaseResponse           wxBaseResponse
	SyncKey                wxSyncKey
	ContinueFlag           int
	AddMsgCount            int
	AddMsgList             wxMsg
	ModChatRoomMemberCount int
	ModContactList         string //TODO
	DelContactList         string //TODO
	ModChatRoomMemberList  string //TODO
	DelContactCount        int
}

type wxSendMsg struct {
	Type         int
	Content      string
	FromUserName string
	ToUserName   string
	LocalID      string
	ClientMsgId  int64
}

type wxSendMsgReq struct {
	BaseRequest wxBaseRequest
	Msg         wxSendMsg
}
type wxSendMsgRsp struct {
	BaseResponse wxBaseResponse
}
type wxMsg struct {
	FromUserName         string
	ToUserName           string
	Content              string // todo
	StatusNotifyUserName string
	ImgWidth             int
	PlayLength           int
	RecommendInfo        string // todo
	StatusNotifyCode     int
	NewMsgId             string
	Status               int
	VoiceLength          int
	ForwardFlag          int
	AppMsgType           int
	Ticket               string
	AppInfo              string //todo
	Url                  string
	ImgStatus            int
	MsgType              int
	ImgHeight            int
	MediaId              string
	MsgId                string
	FileName             string
	HasProductId         int
	FileSize             string
	CreateTime           int64
	SubMsgType           int
}

type wxTextMsg struct {
}
