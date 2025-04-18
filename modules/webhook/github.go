package webhook

import (
	"fmt"
	"io/ioutil"

	"github.com/tangseng-vge/TangSengDaoDaoServerLib/pkg/wkhttp"
)

func (w *Webhook) github(c *wkhttp.Context) {
	fmt.Println("github webhook-->", c.Params)

	result, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println("github-result-->", result)
}
