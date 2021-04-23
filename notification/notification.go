package notification

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"github.com/royeo/dingrobot"
	"gopkg.in/ini.v1"
	"log"
	"os"
)

func Notice(operating string,status_now string,c *gin.Context) {

	//获取地址模板
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Println("文件读取错误", err)
		os.Exit(1)
	}
	webHook := cfg.Section("webhook").Key("url")


	var str =map[string]interface{}{"uid":c.Request.FormValue("unique_id"),"status_now": status_now, "operating": operating,"status":c.Request.FormValue("status")}
	content := "####  节点{{operating}}  \n>  节点{{uid}}状态由{{status_now}} 变更为{{status}} ![screenshot](https://z3.ax1x.com/2021/04/17/c4XKds.png)\n> ###### 发布from  [谐云](http://www.harmonycloud.cn/overindex) \n"

	text :=mustache.Render(content,str)

	//生成dingding机器人通讯
	robot := dingrobot.NewRobot(webHook.String())

	atMobiles := []string{""}
	isAtAll := true
	title := "节点创建更新通知"
	err = robot.SendMarkdown(title,text, atMobiles, isAtAll)
	if err != nil {
		log.Fatal(err)
	}

}
