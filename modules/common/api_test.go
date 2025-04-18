package common

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/util"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/testutil"
)

func TestAddVersion(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	model := &appVersionReq{
		AppVersion:  "1.0",
		OS:          "android",
		DownloadURL: "http://www.githubim.com/download/test.apk",
		IsForce:     1,
		UpdateDesc:  "发布新版本",
	}
	req, _ := http.NewRequest("POST", "/v1/common/appversion", bytes.NewReader([]byte(util.ToJson(model))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetNewVersion(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertAppVersion(&appVersionModel{
		AppVersion:  "1.0",
		OS:          "android",
		DownloadURL: "http://www.githubim.com",
		IsForce:     1,
		UpdateDesc:  "发布新版本",
	})
	assert.NoError(t, err)

	_, err = f.db.insertAppVersion(&appVersionModel{
		AppVersion:  "1.2",
		OS:          "android",
		DownloadURL: "http://www.githubim.com",
		IsForce:     1,
		UpdateDesc:  "发布新版本",
	})
	assert.NoError(t, err)

	f.Route(s.GetRoute())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/common/appversion/android/1.2", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"app_version":1.0`))
}

func TestGetAppConfig(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	err = f.appConfigDB.insert(&appConfigModel{
		WelcomeMessage:                 "欢迎使用唐僧叨叨",
		NewUserJoinSystemGroup:         1,
		RegisterInviteOn:               1,
		InviteSystemAccountJoinGroupOn: 1,
		SendWelcomeMessageOn:           1,
	})
	assert.NoError(t, err)
	//f.Route(s.GetRoute())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/common/appconfig", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"invite_system_account_join_group_on":1`))
}
