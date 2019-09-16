// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package defaultrpc

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/leonlau/mqant/v2/log"
	"github.com/leonlau/mqant/v2/module"
	"github.com/leonlau/mqant/v2/rpc"
	"github.com/leonlau/mqant/v2/rpc/pb"
	"github.com/nats-io/nats.go"
	"runtime"
	"strings"
	"time"
)

type NatsServer struct {
	call_chan chan mqrpc.CallInfo
	addr      string
	app       module.App
	server    *RPCServer
	done      chan error
}

func setAddrs(addrs []string) []string {
	var cAddrs []string
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if !strings.HasPrefix(addr, "nats://") {
			addr = "nats://" + addr
		}
		cAddrs = append(cAddrs, addr)
	}
	if len(cAddrs) == 0 {
		cAddrs = []string{nats.DefaultURL}
	}
	return cAddrs
}

func NewNatsServer(app module.App, s *RPCServer) (*NatsServer, error) {
	server := new(NatsServer)
	server.server = s
	server.done = make(chan error)
	server.app = app
	server.addr = nats.NewInbox()
	go server.on_request_handle()
	return server, nil
}
func (s *NatsServer) Addr() string {
	return s.addr
}

/**
注销消息队列
*/
func (s *NatsServer) Shutdown() (err error) {
	s.done <- nil
	return
}

func (s *NatsServer) Callback(callinfo mqrpc.CallInfo) error {
	body, _ := s.MarshalResult(callinfo.Result)
	reply_to := callinfo.Props["reply_to"].(string)
	return s.app.Transport().Publish(reply_to, body)
}

/**
接收请求信息
*/
func (s *NatsServer) on_request_handle() error {
	defer func() {
		if r := recover(); r != nil {
			var rn = ""
			switch r.(type) {

			case string:
				rn = r.(string)
			case error:
				rn = r.(error).Error()
			}
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			errstr := string(buf[:l])
			log.Errorf("%s\n ----Stack----\n%s", rn, errstr)
		}
	}()
	subs, err := s.app.Transport().SubscribeSync(s.addr)
	if err != nil {
		return err
	}

	go func() {
		<-s.done
		subs.Unsubscribe()
	}()

	for {
		m, err := subs.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			//log.Warning("NatsServer error with '%v'",err)
			continue
		} else if err != nil {
			//log.Error("NatsServer error with '%v'",err)
			return err
		}

		rpcInfo, err := s.Unmarshal(m.Data)
		if err == nil {
			callInfo := &mqrpc.CallInfo{
				RpcInfo: *rpcInfo,
			}
			callInfo.Props = map[string]interface{}{
				"reply_to": rpcInfo.ReplyTo,
			}

			callInfo.Agent = s //设置代理为NatsServer

			s.server.Call(*callInfo)
		} else {
			fmt.Println("error ", err)
		}
	}

	return nil
}

func (s *NatsServer) Unmarshal(data []byte) (*rpcpb.RPCInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var rpcInfo rpcpb.RPCInfo
	err := proto.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}

	panic("bug")
}

// goroutine safe
func (s *NatsServer) MarshalResult(resultInfo rpcpb.ResultInfo) ([]byte, error) {
	//log.Error("",map2)
	b, err := proto.Marshal(&resultInfo)
	return b, err
}
