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
	"fmt"
	"net/http"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-transport/websocket"
	"github.com/go-netty/go-netty/codec/format"
	"github.com/go-netty/go-netty/codec/frame"
)

var ManagerInst = NewManager()

func main() {

	// index page.
	websocket.DefaultOptions.ServeMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(indexHtml)
	})

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// read websocket message
			AddLast(frame.PacketCodec(128)).
			// decode bytes to map[string]interface{}
			AddLast(format.JSONCodec(true, false)).
			// session recorder.
			AddLast(ManagerInst).
			// chat handler.
			AddLast(chatHandler{})
	}

	// setup bootstrap & startup server.
	netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(websocket.New())).
		Listen("0.0.0.0:8080/chat").Sync()
}

type chatHandler struct{}

func (chatHandler) HandleActive(ctx netty.ActiveContext) {
	type wsTransport interface {
		Route() string
		Header() http.Header
	}

	if wst, ok := ctx.Channel().Transport().(wsTransport); ok {
		fmt.Printf("child connection from: %s, route: %s, Websocket-Key: %s, User-Agent: %s\n",
			ctx.Channel().RemoteAddr(), wst.Route(), wst.Header().Get("Sec-Websocket-Key"), wst.Header().Get("User-Agent"))
	}

	ctx.HandleActive()
}

func (chatHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {

	fmt.Printf("received child message from: %s, %v\n", ctx.Channel().RemoteAddr(), message)

	if cmd, ok := message.(map[string]interface{}); ok {
		cmd["id"] = ctx.Channel().ID()
	}

	ManagerInst.Broadcast(message)
}

func (chatHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Printf("child connection closed: %s %s\n", ctx.Channel().RemoteAddr(), ex.Error())
	ctx.HandleInactive(ex)
}
