package reverse_proxy

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	core_ssh "github.com/ttlv/frp_adapter/app/action/reverse_proxy/core/ssh"
	"github.com/ttlv/frp_adapter/app/helpers"
	"github.com/ttlv/frp_adapter/frps_action/frps_fetch"
	"github.com/ttlv/frp_adapter/model"
	"golang.org/x/crypto/ssh"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	UpGrader websocket.Upgrader
}

func NewHandlers(readBufferSize, writeBufferSize int) Handlers {
	return Handlers{UpGrader: websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}}
}

// Frp反向代理,建立ssh长连接
func (handler *Handlers) ReverseProxy(c *gin.Context) {
	// 解析url,获取url中的nodemaintenances_name
	var (
		nodemaintenances_name = strings.TrimSpace(c.Param("nm_name"))
		sCols                 = strings.TrimSpace(c.Param("cols"))
		sRows                 = strings.TrimSpace(c.Param("rows"))
		frpcs                 []model.FrpServer
		err                   error
		wsConn                *websocket.Conn
		sshClient             *ssh.Client
		sshConn               *core_ssh.SshConn
		cols, rows            int
		quitChan              = make(chan bool, 3)
		logBuff               = new(bytes.Buffer)
	)
	// 建立websocket连接
	if wsConn, err = handler.UpGrader.Upgrade(c.Writer, c.Request, nil); err != nil {
		// websocket连接失败 handler
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("Websocket create connection failed, err is: %v", err))
		return
	}
	defer wsConn.Close()
	// 根据nodemaintenances_name获取unique_id通过Frps映射到相应的frpc,拿到frpc的相关数据
	if frpcs, err = frps_fetch.FetchFromFrps(); err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("can't get data from frps, err is %v", err))
		return
	}
	for _, frpc := range frpcs {
		// 找到本次post请求需要进行反向代理的节点
		if frpc.UniqueID == strings.Split(nodemaintenances_name, "-")[1] {
			// 如果frpc的节点状态是离线则无法进行反向代理
			// 确认frpc节点当前是online状态之后再继续进行后续的操作
			if frpc.Status == model.FrpOffline {
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("frp client %v is offline now", frpc.UniqueID))
				return
			}
			// 当rows或者是cols有一个为空则设置一个默认值,cols用以设置web终端的列高，rows用以设置web终端的宽
			if sRows == "" || sCols == "" {
				sRows = "88"
				sCols = "33"
			}
			// 确保 sCols和sRows是合法值
			if cols, err = strconv.Atoi(sCols); err != nil {
				helpers.WsErrorHandle(wsConn, err)
				helpers.RenderFailureJSON(c, http.StatusBadRequest, "the value of the cols is invalid")
				return
			}
			if rows, err = strconv.Atoi(sRows); err != nil {
				helpers.WsErrorHandle(wsConn, err)
				helpers.RenderFailureJSON(c, http.StatusBadRequest, "the value of the cols is invalid")
				return
			}
			// 创建ssh client客户端
			if sshClient, err = core_ssh.NewSshClient(frpc.PublicIpAddress, frpc.Port); err != nil {
				helpers.WsErrorHandle(wsConn, err)
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("ssh client create failed, err is: %v", err))
				return
			}
			defer sshClient.Close()
			// 创建ssh连接
			if sshConn, err = core_ssh.NewSshConn(cols, rows, sshClient); err != nil {
				helpers.WsErrorHandle(wsConn, err)
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("new ssh conection failed, err is: %v", err))
				return
			}
			defer sshConn.Close()
			// 启动goroutine接受来自web端的websocket client传入的数据
			go sshConn.ReceiveWsMsg(wsConn, logBuff, quitChan)
			// 启动goroutine将frpc所在机器的二进制流stdout写入websocket中回传给前端web
			go sshConn.SendComboOutput(wsConn, quitChan)
			// 启动携程保持shh session会话
			go sshConn.SessionWait(quitChan)

			<-quitChan

			logrus.Info("websocket finished")
		}
	}
}

func (handler *Handlers) ReverseProxySshCommand(c *gin.Context) {
	var (
		nm        = c.PostForm("nm_name")
		cmd       = c.PostForm("cmd")
		frpcs     []model.FrpServer
		err       error
		sshClient *ssh.Client
		session   *ssh.Session
	)
	// 根据nodemaintenances_name获取unique_id通过Frps映射到相应的frpc,拿到frpc的相关数据
	if frpcs, err = frps_fetch.FetchFromFrps(); err != nil {
		helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("can't get data from frps, err is %v", err))
		return
	}
	for _, frpc := range frpcs {
		// 找到本次post请求需要进行反向代理的节点
		if frpc.UniqueID == strings.Split(nm, "-")[1] {
			// 如果frpc的节点状态是离线则无法进行反向代理
			// 确认frpc节点当前是online状态之后再继续进行后续的操作
			if frpc.Status == model.FrpOffline {
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("frp client %v is offline now", frpc.UniqueID))
				return
			}
			// 创建ssh client客户端
			if sshClient, err = core_ssh.NewSshClient(frpc.PublicIpAddress, frpc.Port); err != nil {
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("ssh client create failed, err is: %v", err))
				return
			}
			defer sshClient.Close()
			// create session
			if session, err = sshClient.NewSession(); err != nil {
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("ssh client new session failed, err is: %v", err))
				return
			}
			defer session.Close()
			// execute shell command

			result, err := session.Output(cmd)
			if err != nil {
				helpers.RenderFailureJSON(c, http.StatusBadRequest, fmt.Sprintf("execute shell command %v failed, err is: %v", cmd, err))
				return
			}
			helpers.RenderSuccessJSON(c, http.StatusOK, fmt.Sprintf("execute shell command %v successfully,the result is: %v", cmd, strings.TrimSpace(string(result))))
			return
		}
	}
}
