/*
 * Copyright 2019 the go-netty project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/format"
	"github.com/go-netty/go-netty/codec/frame"
	"github.com/go-netty/go-netty/utils"
)

func main() {

	// setup child pipeline initializer.
	childInitializer := func(channel netty.Channel) {
		channel.Pipeline().
			AddLast(frame.LengthFieldCodec(binary.LittleEndian, 1024, 0, 2, 0, 2)).
			AddLast(format.TextCodec()).
			AddLast(EchoHandler{"Server"})
	}

	// setup client pipeline initializer.
	clientInitializer := func(channel netty.Channel) {
		channel.Pipeline().
			AddLast(frame.LengthFieldCodec(binary.LittleEndian, 1024, 0, 2, 0, 2)).
			AddLast(format.TextCodec()).
			AddLast(EchoHandler{"Client"})
	}

	// new bootstrap
	var bootstrap = netty.NewBootstrap(netty.WithChildInitializer(childInitializer), netty.WithClientInitializer(clientInitializer))

	// connect to the server after 1 second
	time.AfterFunc(time.Second, func() {
		_, err := bootstrap.Connect("127.0.0.1:6565", nil)
		utils.Assert(err)
	})

	// setup bootstrap & startup server.
	bootstrap.Listen("0.0.0.0:6565").Sync()
}

type EchoHandler struct {
	role string
}

func (l EchoHandler) HandleActive(ctx netty.ActiveContext) {
	fmt.Println(l.role, "->", "active:", ctx.Channel().RemoteAddr())

	ctx.Write("Hello I'm " + l.role)
	ctx.HandleActive()
}

func (l EchoHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
	fmt.Println(l.role, "->", "handle read:", message)
	ctx.HandleRead(message)
}

func (l EchoHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Println(l.role, "->", "inactive:", ctx.Channel().RemoteAddr(), ex)
	ctx.HandleInactive(ex)
}
