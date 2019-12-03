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
	"bufio"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/transport/tcp"
	"github.com/go-netty/go-netty/utils"
	"github.com/rokumoe/redisgo"
	"os"
	"strings"
)

func main() {
	// new bootstrap
	var bootstrap = netty.NewBootstrap()

	// setup client pipeline initializer.
	bootstrap.ClientInitializer(func(channel netty.Channel) {
		channel.Pipeline().
			AddLast(&simpleRedisCodec{}, &simpleRedisConsole{})
	})

	// setup bootstrap & startup server.
	bootstrap.Transport(tcp.New())

	// connect to redis server
	fmt.Println("connecting redis server ...")

	_, err := bootstrap.Connect("tcp://192.168.212.212:6379", nil)
	utils.Assert(err)

	// wait quit signal to stop the bootstrap.
	bootstrap.RunForever(os.Interrupt, os.Kill)
}

type simpleRedisConsole struct {
	respChan chan *redisgo.Resp
}

func (s *simpleRedisConsole) HandleActive(ctx netty.ActiveContext) {
	fmt.Println("connected")

	go s.runConsole(ctx)
}

func (s *simpleRedisConsole) HandleRead(ctx netty.InboundContext, message netty.Message) {
	s.respChan <- message.(*redisgo.Resp)
}

func (s *simpleRedisConsole) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Println("disconnected", ex)

	ctx.HandleInactive(ex)
}

func (s *simpleRedisConsole) runConsole(ctx netty.HandlerContext) {

	s.respChan = make(chan *redisgo.Resp, 1)

	fmt.Print(ctx.Channel().RemoteAddr(), ">")

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {

		if text := strings.TrimSpace(stdin.Text()); len(text) > 0 {

			switch text {
			case "exit":
				fmt.Println("exited")
				os.Exit(0)
				return
			case "help":
				fmt.Println("help information")
			default:
				inputs := strings.Split(text, " ")

				// build command.
				var cmds = make([]redisgo.Value, 0, len(inputs))
				for _, v := range inputs {
					cmds = append(cmds, redisgo.BlukString(v))
				}

				// send redis command.
				ctx.Write(cmds)

				// print response
				resp := <-s.respChan
				if resp.Null {
					fmt.Println("(empty)")
				} else {
					fmt.Println(resp.Data)
				}
			}
		} else {
			fmt.Println("please input redis command")
		}

		fmt.Print(ctx.Channel().RemoteAddr(), ">")
	}
}
