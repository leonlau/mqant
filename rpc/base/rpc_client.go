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
	"github.com/leonlau/mqant/v2/rpc/util"
	"github.com/leonlau/mqant/v2/utils/uuid"
	"time"
)

type RPCClient struct {
	app         module.App
	nats_client *NatsClient
}

func NewRPCClient(app module.App, session module.ServerSession) (mqrpc.RPCClient, error) {
	rpc_client := new(RPCClient)
	rpc_client.app = app
	nats_client, err := NewNatsClient(app, session)
	if err != nil {
		log.Errorf("Dial: %s", err)
		return nil, err
	}
	rpc_client.nats_client = nats_client
	return rpc_client, nil
}

func (c *RPCClient) Done() (err error) {
	if c.nats_client != nil {
		err = c.nats_client.Done()
	}
	return
}

func (c *RPCClient) CallArgs(_func string, ArgsType []string, args [][]byte) (interface{}, string) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:       *proto.String(_func),
		Reply:    *proto.Bool(true),
		Expired:  *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().Rpc.RpcExpired)).UnixNano()) / 1000000),
		Cid:      *proto.String(correlation_id),
		Args:     args,
		ArgsType: ArgsType,
	}

	callInfo := &mqrpc.CallInfo{
		RpcInfo: *rpcInfo,
	}
	callback := make(chan rpcpb.ResultInfo, 1)
	var err error
	//优先使用本地rpc
	//if c.local_client != nil {
	//	err = c.local_client.Call(*callInfo, callback)
	//} else
	err = c.nats_client.Call(*callInfo, callback)
	if err != nil {
		return nil, err.Error()
	}
	select {
	case resultInfo, ok := <-callback:
		if !ok {
			return nil, "client closed"
		}
		result, err := argsutil.Bytes2Args(c.app, resultInfo.ResultType, resultInfo.Result)
		if err != nil {
			return nil, err.Error()
		}
		return result, resultInfo.Error
	case <-time.After(time.Second * time.Duration(c.app.GetSettings().Rpc.RpcExpired)):
		close(callback)
		c.nats_client.Delete(rpcInfo.Cid)
		return nil, "deadline exceeded"
	}
}

func (c *RPCClient) CallNRArgs(_func string, ArgsType []string, args [][]byte) (err error) {
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &rpcpb.RPCInfo{
		Fn:       *proto.String(_func),
		Reply:    *proto.Bool(false),
		Expired:  *proto.Int64((time.Now().UTC().Add(time.Second * time.Duration(c.app.GetSettings().Rpc.RpcExpired)).UnixNano()) / 1000000),
		Cid:      *proto.String(correlation_id),
		Args:     args,
		ArgsType: ArgsType,
	}
	callInfo := &mqrpc.CallInfo{
		RpcInfo: *rpcInfo,
	}
	//优先使用本地rpc
	//if c.local_client != nil {
	//	err = c.local_client.CallNR(*callInfo)
	//} else
	return c.nats_client.CallNR(*callInfo)
}

/**
消息请求 需要回复
*/
func (c *RPCClient) Call(_func string, params ...interface{}) (interface{}, string) {
	var ArgsType []string = make([]string, len(params))
	var args [][]byte = make([][]byte, len(params))
	var span log.TraceSpan = nil
	for k, param := range params {
		var err error = nil
		ArgsType[k], args[k], err = argsutil.ArgsTypeAnd2Bytes(c.app, param)
		if err != nil {
			return nil, fmt.Sprintf("args[%d] error %s", k, err.Error())
		}
		switch v2 := param.(type) { //多选语句switch
		case log.TraceSpan:
			//如果参数是这个需要拷贝一份新的再传
			span = v2
		}
	}
	start := time.Now()
	r, errstr := c.CallArgs(_func, ArgsType, args)
	if c.app.GetSettings().Rpc.Log {
		log.TInfo(span, "RPC Call ServerId = %v Func = %v Elapsed = %v Result = %v ERROR = %v", c.nats_client.session.GetId(), _func, time.Since(start), r, errstr)
	}
	return r, errstr
}

/**
消息请求 不需要回复
*/
func (c *RPCClient) CallNR(_func string, params ...interface{}) (err error) {
	var ArgsType []string = make([]string, len(params))
	var args [][]byte = make([][]byte, len(params))
	var span log.TraceSpan = nil
	for k, param := range params {
		ArgsType[k], args[k], err = argsutil.ArgsTypeAnd2Bytes(c.app, param)
		if err != nil {
			return fmt.Errorf("args[%d] error %s", k, err.Error())
		}

		switch v2 := param.(type) { //多选语句switch
		case log.TraceSpan:
			//如果参数是这个需要拷贝一份新的再传
			span = v2
		}
	}
	start := time.Now()
	err = c.CallNRArgs(_func, ArgsType, args)
	if c.app.GetSettings().Rpc.Log {
		log.TInfo(span, "RPC CallNR ServerId = %v Func = %v Elapsed = %v ERROR = %v", c.nats_client.session.GetId(), _func, time.Since(start), err)
	}
	return err
}
