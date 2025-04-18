package file_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/file"
	"github.com/stretchr/testify/assert"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/config"
	"github.com/tangseng-vge/TangSengDaoDaoServerLib/testutil"
)

func TestOSSUpload(t *testing.T) {
	cfg := config.New()
	ctx := testutil.NewTestContext(cfg)
	cfg.OSS.Endpoint = "oss-cn-shanghai.aliyuncs.com"
	cfg.OSS.AccessKeyID = "xxxx"
	cfg.OSS.AccessKeySecret = "xxxxxx"

	service := file.NewServiceOSS(ctx)
	_, err := service.UploadFile("chat/zdd/fjj.txt", "*", func(writer io.Writer) error {
		_, err := writer.Write(bytes.NewBufferString("this is test content").Bytes())
		return err
	})
	assert.NoError(t, err)

}
