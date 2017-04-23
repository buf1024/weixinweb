package weixinweb

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

type baseRequest struct {
	Uin      string
	Sid      string
	Skey     string
	DeviceID string
}

type baseResponse struct {
	Ret    int
	ErrMsg string
}

type member struct {
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

type user struct {
	Uin               int
	UserName          string
	NickName          string
	HeadImgUrl        string
	ContactFlag       int
	MemberCount       int
	MemberList        []member
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

type article struct {
	Title  string
	Digest string
	Cover  string
	Url    string
}
type subscribeMsg struct {
	UserName       string
	MPArticleCount int
	MPArticleList  []article
	Time           int64
	NickName       string
}

//////////////////

type wxReq struct {
	BaseRequest *baseRequest
}

type wxRsp struct {
	BaseResponse        *baseResponse
	Count               int
	ContactList         []user
	SyncKey             wxSyncKey
	User                user
	ChatSet             string
	SKey                string
	ClientVersion       int64
	SystemTime          int64
	GrayScale           int
	InviteStartCount    int
	MPSubscribeMsgCount int
	MPSubscribeMsgList  []subscribeMsg
	ClickReportInterval int
	MsgID               string
}

type wxStatusNotifyReq struct {
	BaseRequest  *baseRequest
	Code         int
	FromUserName string
	ToUserName   string
	ClientMsgId  int64
}
type wxSyncReq struct {
	BaseRequest *baseRequest
	SyncKey     wxSyncKey
	RR          int64 `json:"rr"`
}

type wxSyncCheckReq struct {
	BaseRequest  *baseRequest
	Code         int
	FromUserName string
	ToUserName   string
	ClientMsgId  int64
}
