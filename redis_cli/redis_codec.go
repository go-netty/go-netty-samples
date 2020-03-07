/*
 *  Copyright 2019 the go-netty project
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       https://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"bytes"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-samples/redis_cli/redisgo"
	"github.com/go-netty/go-netty/utils"
	"io"
)

type simpleRedisCodec struct {
	decoder *redisgo.Decoder
}

func (s *simpleRedisCodec) CodecName() string {
	return "simple-redis-codec"
}

func (s *simpleRedisCodec) HandleRead(ctx netty.InboundContext, message netty.Message) {

	// init decoder.
	if nil == s.decoder {
		s.decoder = redisgo.NewDecoder(message.(io.Reader), 10240)
	}

	// decode redis response.
	resp := &redisgo.Resp{}
	utils.Assert(s.decoder.Decode(resp))

	// fmt.Println("->", resp.String())
	// post response.
	ctx.HandleRead(resp)
}

func (s *simpleRedisCodec) HandleWrite(ctx netty.OutboundContext, message netty.Message) {

	switch v := message.(type) {
	case []redisgo.Value:
		// encode request.
		buffer := bytes.NewBuffer(nil)
		utils.Assert(redisgo.EncodeMulti(buffer, v...))

		// fmt.Println("<-", buffer.String())
		// post request.
		ctx.HandleWrite(buffer)
	default:
		utils.Assert(fmt.Errorf("%T is invalid message", message))
	}
}
