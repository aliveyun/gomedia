package rtmp

import (
	"fmt"
	"net"
	"net/url"
	"time"
	//neturl "net/url"
)

// Start ...
func (cc *RtmpClient) ConnectServer(rtmpUrl string) (err error) {

	u, err := url.Parse(rtmpUrl)
	if err != nil {
		panic(err)
	}
	host := u.Host
	if u.Port() == "" {
		host += ":1935"
	}
	//connect to remote rtmp server
	//c, err := net.Dial("tcp4", host)

	c, err := net.DialTimeout("tcp", host, time.Second*2)

	if err != nil {
		fmt.Println("connect failed", err)
		return err
	}

	cc.SetOutput(func(data []byte) error {
		_, err := c.Write(data)
		return err
	})
	isReady := make(chan struct{})
	//监听状态变化,STATE_RTMP_PUBLISH_START 状态通知推流
	cc.OnStateChange(func(newState RtmpState) {
		if newState == STATE_RTMP_PUBLISH_START {
			fmt.Println("ready for publish")
			close(isReady)
		}
	})

	cc.Start(rtmpUrl)
	go cc.Monitor(c)
	select {
	case <-isReady:
	case <-time.NewTimer(time.Millisecond * 20).C:
	}
	return nil
}
func (cc *RtmpClient) Monitor(c net.Conn) {
	defer func() {
		//client.Signals <- SignalStreamStop
	}()
	var err error

	buf := make([]byte, 4096)
	n := 0
	for err == nil {
		n, err = c.Read(buf)
		if err != nil {
			return
		}
		cc.Input(buf[:n])
	}
	fmt.Println(err)
}

// Close 关闭连接，并回调onClosed
func (c *RtmpClient) Close() error {
	c.conn.Close()
	// if c.onClosed != nil {
	// 	c.onClosed()
	// }
	return nil
}
