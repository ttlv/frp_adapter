package notification

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"gopkg.in/ini.v1"
	"net/http"
	"os"
	"strings"
)

func Notice(operating string,status_now string,c *gin.Context) {

	//获取地址模板
	cfg, err := ini.Load("../config.ini")
	if err != nil {
		fmt.Println("文件读取错误", err)
		os.Exit(1)
	}
	webHook := cfg.Section("webhook").Key("url")


	var str =map[string]interface{}{"uid":c.Request.FormValue("unique_id"),"status_now": status_now, "operating": operating,"status":c.Request.FormValue("status")}
	content :=`{ "msgtype": "markdown",
		"markdown": {
		"title":"节点创建更新通知",
			"text": "####  节点{{operating}}  \n>  节点{{uid}}状态由{{status_now}} 变更为{{status}} ![screenshot](https://z3.ax1x.com/2021/04/17/c4XKds.png)\n> ###### 发布from  [谐云](http://www.harmonycloud.cn/overindex) \n"
	}
}`

	view :=mustache.Render(content,str)
	//创建一个请求
	req, err := http.NewRequest("POST", webHook.String(), strings.NewReader(view))
	if err != nil {
		// handle error
		fmt.Errorf(fmt.Sprintf("notice request failed:%v",err))
		return
	}

	client := &http.Client{}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Errorf(fmt.Sprintf("notice request failed:%v",err))
		return
	}//关闭请求
	defer resp.Body.Close()


}
