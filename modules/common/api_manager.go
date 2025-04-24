package common

import (
	"errors"
	"strings"

	"github.com/tangseng-vge/TangSengDaoDaoServerLib/config"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/log"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Manager 通用后台管理api
type Manager struct {
	ctx *config.Context
	log.Log
	db          *db
	appconfigDB *AppConfigDb
}

// NewManager NewManager
func NewManager(ctx *config.Context) *Manager {
	return &Manager{
		ctx:         ctx,
		Log:         log.NewTLog("commonManager"),
		db:          newDB(ctx.DB()),
		appconfigDB: newAppConfigDB(ctx),
	}
}

// Route 配置路由规则
func (m *Manager) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/manager", m.ctx.AuthMiddleware(r))
	{
		auth.GET("/common/appconfig", m.appconfig)               // 获取app配置
		auth.POST("/common/appconfig", m.updateConfig)           // 修改app配置
		auth.GET("/common/appmodule", m.getAppModule)            // 获取app模块
		auth.PUT("/common/appmodule", m.updateAppModule)         // 修改app模块
		auth.POST("/common/appmodule", m.addAppModule)           // 新增app模块
		auth.DELETE("/common/:sid/appmodule", m.deleteAppModule) // 删除app模块
	}
}
func (m *Manager) deleteAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	sid := c.Param("sid")
	if strings.TrimSpace(sid) == "" {
		c.ResponseError(errors.New("sid不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(sid)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("删除的模块不存在"))
		return
	}
	err = m.db.deleteAppModule(sid)
	if err != nil {
		m.Error("删除app模块错误", zap.Error(err))
		c.ResponseError(errors.New("删除app模块错误"))
		return
	}
	c.ResponseOK()
}

// 新增app模块
func (m *Manager) addAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module != nil && module.SID != "" {
		c.ResponseError(errors.New("该sid模块已存在"))
		return
	}
	_, err = m.db.insertAppModule(&appModuleModel{
		SID:    req.SID,
		Name:   req.Name,
		Desc:   req.Desc,
		Status: req.Status,
	})
	if err != nil {
		m.Error("新增app模块错误", zap.Error(err))
		c.ResponseError(errors.New("新增app模块错误"))
		return
	}
	c.ResponseOK()
}
func (m *Manager) updateAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("不存在该模块"))
		return
	}
	module.Name = req.Name
	module.Desc = req.Desc
	module.Status = req.Status
	err = m.db.updateAppModule(module)
	if err != nil {
		m.Error("修改app模块错误", zap.Error(err))
		c.ResponseError(errors.New("修改app模块错误"))
		return
	}
	c.ResponseOK()
}

// 获取app模块
func (m *Manager) getAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	modules, err := m.db.queryAppModule()
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	list := make([]*managerAppModule, 0)
	if len(modules) > 0 {
		for _, module := range modules {
			list = append(list, &managerAppModule{
				SID:    module.SID,
				Name:   module.Name,
				Desc:   module.Desc,
				Status: module.Status,
			})
		}
	}
	c.Response(list)
}
func (m *Manager) updateConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type reqVO struct {
		RevokeSecond                   int    `json:"revoke_second"`
		WelcomeMessage                 string `json:"welcome_message"`
		NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
		SearchByPhone                  int    `json:"search_by_phone"`
		RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
		SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
		InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
		RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
		ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
		CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
		ApiAddr                        string `json:"api_addr"`                            // 是否可以修改api地址
		ApiAddrJw                      string `json:"api_addr_jw"`                         // 是否可以修改api地址
		WebAddr                        string `json:"web_addr"`                            // 是否可以修改api地址
		WebAddrJw                      string `json:"web_addr_jw"`                         // 是否可以修改api地址
		WsAddr                         string `json:"ws_addr"`                             // 是否可以修改api地址
		WsAddrJw                       string `json:"ws_addr_jw"`                          // 是否可以修改api地址
		WssAddr                        string `json:"wss_addr"`                            // 是否可以修改api地址
		WssAddrJw                      string `json:"wss_addr_jw"`                         // 是否可以修改api地址
		SocketAddr                     string `json:"socket_addr"`                         // 是否可以修改api地址
		SocketAddrJw                   string `json:"socket_addr_jw"`                      // 是否可以修改api地址
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	appConfigM, err := m.appconfigDB.Query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}
	configMap := map[string]interface{}{}
	configMap["revoke_second"] = req.RevokeSecond
	configMap["welcome_message"] = req.WelcomeMessage
	configMap["new_user_join_system_group"] = req.NewUserJoinSystemGroup
	configMap["search_by_phone"] = req.SearchByPhone
	configMap["register_invite_on"] = req.RegisterInviteOn
	configMap["send_welcome_message_on"] = req.SendWelcomeMessageOn
	configMap["invite_system_account_join_group_on"] = req.InviteSystemAccountJoinGroupOn
	configMap["register_user_must_complete_info_on"] = req.RegisterUserMustCompleteInfoOn
	configMap["channel_pinned_message_max_count"] = req.ChannelPinnedMessageMaxCount
	configMap["can_modify_api_url"] = req.CanModifyApiUrl
	configMap["api_addr"] = req.ApiAddr
	configMap["api_addr_jw"] = req.ApiAddrJw
	configMap["web_addr"] = req.WebAddr
	configMap["web_addr_jw"] = req.WebAddrJw
	configMap["ws_addr"] = req.WsAddr
	configMap["ws_addr_jw"] = req.WsAddrJw
	configMap["wss_addr"] = req.WssAddr
	configMap["wss_addr_jw"] = req.WssAddrJw
	configMap["socket_addr"] = req.SocketAddr
	configMap["socket_addr_jw"] = req.SocketAddrJw

	err = m.appconfigDB.updateWithMap(configMap, appConfigM.Id)
	if err != nil {
		m.Error("修改app配置信息错误", zap.Error(err))
		c.ResponseError(errors.New("修改app配置信息错误"))
		return
	}
	c.ResponseOK()
}
func (m *Manager) appconfig(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	appconfig, err := m.appconfigDB.Query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}
	var revokeSecond = 0
	var newUserJoinSystemGroup = 1
	var welcomeMessage = ""
	var searchByPhone = 1
	var registerInviteOn = 0
	var sendWelcomeMessageOn = 0
	var inviteSystemAccountJoinGroupOn = 0
	var registerUserMustCompleteInfoOn = 0
	var channelPinnedMessageMaxCount = 10
	var canModifyApiUrl = 0
	var api_addr = ""
	var api_addr_jw = ""
	var web_addr = ""
	var web_addr_jw = ""
	var ws_addr = ""
	var ws_addr_jw = ""
	var wss_addr = ""
	var wss_addr_jw = ""
	var socket_addr = ""
	var socket_addr_jw = ""

	if appconfig != nil {
		revokeSecond = appconfig.RevokeSecond
		welcomeMessage = appconfig.WelcomeMessage
		newUserJoinSystemGroup = appconfig.NewUserJoinSystemGroup
		searchByPhone = appconfig.SearchByPhone
		registerInviteOn = appconfig.RegisterInviteOn
		sendWelcomeMessageOn = appconfig.SendWelcomeMessageOn
		inviteSystemAccountJoinGroupOn = appconfig.InviteSystemAccountJoinGroupOn
		registerUserMustCompleteInfoOn = appconfig.RegisterUserMustCompleteInfoOn
		channelPinnedMessageMaxCount = appconfig.ChannelPinnedMessageMaxCount
		canModifyApiUrl = appconfig.CanModifyApiUrl
		api_addr = appconfig.ApiAddr
		api_addr_jw = appconfig.ApiAddrJw
		web_addr = appconfig.WebAddr
		web_addr_jw = appconfig.WebAddrJw
		ws_addr = appconfig.WsAddr
		ws_addr_jw = appconfig.WsAddrJw
		wss_addr = appconfig.WssAddr
		wss_addr_jw = appconfig.WssAddrJw
		socket_addr = appconfig.SocketAddr
		socket_addr_jw = appconfig.SocketAddrJw
	}
	if revokeSecond == 0 {
		revokeSecond = 120
	}
	if welcomeMessage == "" {
		welcomeMessage = m.ctx.GetConfig().WelcomeMessage
	}
	c.Response(&managerAppConfigResp{
		RevokeSecond:                   revokeSecond,
		WelcomeMessage:                 welcomeMessage,
		NewUserJoinSystemGroup:         newUserJoinSystemGroup,
		SearchByPhone:                  searchByPhone,
		RegisterInviteOn:               registerInviteOn,
		SendWelcomeMessageOn:           sendWelcomeMessageOn,
		InviteSystemAccountJoinGroupOn: inviteSystemAccountJoinGroupOn,
		RegisterUserMustCompleteInfoOn: registerUserMustCompleteInfoOn,
		ChannelPinnedMessageMaxCount:   channelPinnedMessageMaxCount,
		CanModifyApiUrl:                canModifyApiUrl,
		ApiAddr:                        api_addr,
		ApiAddrJw:                      api_addr_jw,
		WebAddr:                        web_addr,
		WebAddrJw:                      web_addr_jw,
		WsAddr:                         ws_addr,
		WsAddrJw:                       ws_addr_jw,
		WssAddr:                        wss_addr,
		WssAddrJw:                      wss_addr_jw,
		SocketAddr:                     socket_addr,
		SocketAddrJw:                   socket_addr_jw,
	})
}

type managerAppConfigResp struct {
	RevokeSecond                   int    `json:"revoke_second"`
	WelcomeMessage                 string `json:"welcome_message"`
	NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
	SearchByPhone                  int    `json:"search_by_phone"`
	RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
	SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
	InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
	RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
	ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
	CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
	ApiAddr                        string `json:"api_addr"`
	ApiAddrJw                      string `json:"api_addr_jw"`
	WebAddr                        string `json:"web_addr"`
	WebAddrJw                      string `json:"web_addr_jw"`
	WsAddr                         string `json:"ws_addr"`
	WsAddrJw                       string `json:"ws_addr_jw"`
	WssAddr                        string `json:"wss_addr"`
	WssAddrJw                      string `json:"wss_addr_jw"`
	SocketAddr                     string `json:"socket_addr"`
	SocketAddrJw                   string `json:"socket_addr_jw"`
}

type managerAppModule struct {
	SID    string `json:"sid"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Status int    `json:"status"` // 模块状态 1.可选 0.不可选 2.选中不可编辑
}
