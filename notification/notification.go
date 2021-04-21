package notification

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Notice(operating string,status_now string,c *gin.Context) {
	//请求地址模板
	webHook := `https://oapi.dingtalk.com/robot/send?access_token=41d052ecba12f0ec45f4e68657e3dd0544b2b1ceb71d40b64240416881e7826b`
	/*content := `{"msgtype": "text",
		"text": {"content": "`+ msg + `"}
	}`*/
	str :="节点:"+ c.Request.FormValue("unique_id") + "状态由" + status_now+"变更为"+c.Request.FormValue("status")
	content :=`{ "msgtype": "markdown",
		"markdown": {
		"title":"节点创建更新通知",
			"text": "#### `+ operating +` \n>` + str + `![screenshot](https://z3.ax1x.com/2021/04/17/c4XKds.png)\n> ###### 发布from  [谐云](http://www.harmonycloud.cn/overindex) \n"
	}
	}`
	//创建一个请求
	req, err := http.NewRequest("POST", webHook, strings.NewReader(content))
	if err != nil {
		// handle error
	}

	client := &http.Client{}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)
	//关闭请求
	defer resp.Body.Close()

	if err != nil {
		// handle error
	}
}

func Notice1(operating string) {
	//请求地址模板
	webHook := `https://oapi.dingtalk.com/robot/send?access_token=41d052ecba12f0ec45f4e68657e3dd0544b2b1ceb71d40b64240416881e7826b`
	content := `{"msgtype": "text",
		"text": {"content": "`+ operating +"通知"+ `"}
	}`
	//content :=`{ "msgtype": "markdown",
	//	"markdown": {
	//	"title":"节点创建更新通知",
	//		"text": "#### `+ operating +` \n>![screenshot](https://z3.ax1x.com/2021/04/17/c4XKds.png)\n> ###### 发布from  [谐云](http://www.harmonycloud.cn/overindex) \n"
	//}
	//}`
	//创建一个请求
	req, err := http.NewRequest("POST", webHook, strings.NewReader(content))
	if err != nil {
		// handle error
	}

	client := &http.Client{}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)
	//关闭请求
	defer resp.Body.Close()

	if err != nil {
		// handle error
	}
}