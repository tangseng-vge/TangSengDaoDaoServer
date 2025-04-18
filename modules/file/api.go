package file

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	utils "github.com/TangSengDaoDao/TangSengDaoDaoServer/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/config"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/log"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/util"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// File 文件操作
type File struct {
	ctx *config.Context
	log.Log
	service IService
	cn      *AppConfig
}

type AppConfig struct {
	ctx *config.Context
	log.Log
	appConfigDB *appConfigDB
}

// New New
func New(ctx *config.Context) *File {
	return &File{
		ctx:     ctx,
		Log:     log.NewTLog("File"),
		service: NewService(ctx),
	}
}

// Route 路由
func (f *File) Route(r *wkhttp.WKHttp) {
	api := r.Group("/v1/file")
	{ // 文件上传
		// api.POST("/upload/*path", f.upload)
		// 组合图片
		api.POST("/compose/*path", f.makeImageCompose)
		// 获取文件
		api.GET("/preview/*path", f.getFile)
	}
	auth := r.Group("/v1/file", f.ctx.AuthMiddleware(r))
	{
		//获取上传文件地址
		auth.GET("/upload", f.getFilePath)
		//上传文件
		auth.POST("/upload", f.uploadFile)
	}
}

func (f *File) makeImageCompose(c *wkhttp.Context) {
	var imageURLs []string
	if err := c.BindJSON(&imageURLs); err != nil {
		f.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if len(imageURLs) <= 0 {
		c.ResponseError(errors.New("图片不能为空！"))
		return
	}
	if len(imageURLs) > 9 {
		c.ResponseError(errors.New("图片数量不能大于9！"))
		return
	}
	uploadPath := c.Param("path")
	// 下载并组合图片
	resultMap, err := f.service.DownloadAndMakeCompose(uploadPath, imageURLs)
	if err != nil {
		f.Error("组合图片失败！", zap.String("uploadPath", uploadPath), zap.Any("imageURLs", imageURLs), zap.Error(err))
		c.ResponseError(errors.New("组合图片失败！"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path": resultMap["fid"].(string),
	})
}

// 获取上传文件地址
func (f *File) getFilePath(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	uploadPath := c.Query("path")
	fileType := c.Query("type")
	err := f.checkReq(Type(fileType), uploadPath)
	if err != nil {
		c.ResponseError(err)
		return
	}
	var path string
	//  发送图片会显示真实ip地址
	appConfigM, err := f.cn.appConfigDB.query()
	if err != nil {
		f.Error("读取上传配置失败！", zap.Error(err))
		c.ResponseError(errors.New("读取上传配置失败！"))
		return
	}
	if appConfigM == nil {
		f.Error("读取上传配置失败1！", zap.Error(err))
		c.ResponseError(errors.New("读取上传配置失败1！"))
	}
	// 获取当前客户IP
	ip := utils.GetClientPublicIP(c.Request)
	area := utils.GetInstance().GetArea(ip)
	var BASEURL = ""
	if "CN" != area {
		BASEURL = appConfigM.ApiAddr
	} else {
		BASEURL = appConfigM.ApiAddrJw
	}

	if Type(fileType) == TypeMomentCover {
		// 动态封面
		path = fmt.Sprintf("%s/file/upload?type=%s&path=/%s.png", BASEURL, fileType, loginUID)
	} else if Type(fileType) == TypeSticker {
		// 自定义表情
		path = fmt.Sprintf("%s/file/upload?type=%s&path=/%s/%s.gif", BASEURL, fileType, loginUID, util.GenerUUID())
	} else if Type(fileType) == TypeWorkplaceBanner {
		// 工作台横幅
		path = fmt.Sprintf("%s/file/upload?type=%s&path=/workplace/banner/%s", BASEURL, fileType, path)
	} else if Type(fileType) == TypeWorkplaceAppIcon {
		// 工作台appIcon
		path = fmt.Sprintf("%s/file/upload?type=%s&path=/workplace/appicon/%s", BASEURL, fileType, path)
	} else {
		path = fmt.Sprintf("%s/file/upload?type=%s&path=%s", BASEURL, fileType, uploadPath)
	}
	c.Response(map[string]string{
		"url": path,
	})
}

// 上传文件
func (f *File) uploadFile(c *wkhttp.Context) {
	uploadPath := c.Query("path")
	fileType := c.Query("type")
	signature := c.Query("signature") // 是否返回签名
	var signatureInt int64 = 0
	if signature != "" {
		signatureInt, _ = strconv.ParseInt(signature, 10, 64)
	}
	contentType := c.DefaultPostForm("contenttype", "application/octet-stream")
	err := f.checkReq(Type(fileType), uploadPath)
	if err != nil {
		c.ResponseError(err)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		f.Error("读取文件失败！", zap.Error(err))
		c.ResponseError(errors.New("读取文件失败！"))
		return
	}
	path := uploadPath
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}
	var sign []byte
	if signatureInt == 1 {
		// bytes, err := ioutil.ReadAll(file)
		// if err != nil {
		// 	f.Error("读取文件错误", zap.Error(err))
		// 	c.ResponseError(errors.New("读取文件错误"))
		// 	return
		// }
		h := sha512.New()
		_, err := io.Copy(h, file)
		if err != nil {
			f.Error("签名复制文件错误", zap.Error(err))
			c.ResponseError(errors.New("签名复制文件错误"))
			return
		}
		sign = h.Sum(nil)
		//	sign = sha512.Sum512(bytes)

	}
	_, err = f.service.UploadFile(fmt.Sprintf("%s%s", fileType, path), contentType, func(w io.Writer) error {
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			f.Error("设置文件偏移量错误", zap.Error(err))
			return err
		}
		_, err = io.Copy(w, file)
		return err
	})

	defer file.Close()
	if err != nil {
		f.Error("上传文件失败！", zap.Error(err))
		c.ResponseError(errors.New("上传文件失败！"))
		return
	}
	if signatureInt == 1 {
		encoded := base64.StdEncoding.EncodeToString(sign[:])
		fmt.Print("编码文件", encoded)
		c.Response(map[string]interface{}{
			"path":   fmt.Sprintf("file/preview/%s%s", fileType, path),
			"sha512": encoded,
		})
	} else {
		c.Response(map[string]string{
			"path": fmt.Sprintf("file/preview/%s%s", fileType, path),
		})
	}
}

// 获取文件
func (f *File) getFile(c *wkhttp.Context) {
	ph := c.Param("path")
	if ph == "" {
		c.Response(errors.New("访问路径不能为空"))
		return
	}
	filename := c.Query("filename")
	if filename == "" {
		paths := strings.Split(ph, "/")
		if len(paths) > 0 {
			filename = paths[len(paths)-1]
		}
	}
	downloadURL, err := f.service.DownloadURL(ph, filename)
	if err != nil {
		c.ResponseError(err)
		return
	}
	c.Redirect(http.StatusFound, downloadURL)
}

func (f *File) checkReq(fileType Type, path string) error {
	if fileType == "" {
		return errors.New("文件类型不能为空")
	}
	if path == "" && fileType != TypeMomentCover && fileType != TypeSticker {
		return errors.New("上传路径不能为空")
	}
	if fileType != TypeChat && fileType != TypeMoment && fileType != TypeMomentCover && fileType != TypeSticker && fileType != TypeReport && fileType != TypeChatBg && fileType != TypeCommon && fileType != TypeDownload {
		return errors.New("文件类型错误")
	}
	return nil
}
