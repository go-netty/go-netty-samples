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
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/transport/tcp"
	"github.com/go-netty/go-netty/utils"
)

func main() {

	// new bootstrap
	var bootstrap = netty.NewBootstrap().Transport(tcp.New())

	// setup client pipeline initializer.
	bootstrap.ClientInitializer(func(channel netty.Channel) {
		channel.Pipeline().
			AddLast(&simpleRedisCodec{}, &simpleRedisConsole{})
	})

	// connect to redis server
	fmt.Println("connecting redis server ...")

	ch, err := bootstrap.Connect("tcp://192.168.212.212:6379", nil)
	utils.Assert(err)

	select {
	case <-ch.Context().Done():
	case <-bootstrap.Context().Done():
	}

	fmt.Println("exited")
}
