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
	"os"
	"strings"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-samples/redis_cli/redisgo"
)

type simpleRedisConsole struct {
	respChan chan *redisgo.Resp
}

func (s *simpleRedisConsole) HandleActive(ctx netty.ActiveContext) {
	fmt.Println("connected")

	go s.attachConsole(ctx)
	ctx.HandleActive()
}

func (s *simpleRedisConsole) HandleRead(ctx netty.InboundContext, message netty.Message) {
	s.respChan <- message.(*redisgo.Resp)
}

func (s *simpleRedisConsole) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Println("disconnected", ex)

	ctx.HandleInactive(ex)
}

func (s *simpleRedisConsole) attachConsole(ctx netty.HandlerContext) {

	s.respChan = make(chan *redisgo.Resp, 1)

	fmt.Print(ctx.Channel().RemoteAddr(), ">")

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {

		if text := strings.TrimSpace(stdin.Text()); len(text) > 0 {

			switch text {
			case "exit":
				ctx.Close(fmt.Errorf("user exit"))
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

				switch {
				case resp.Null:
					fmt.Println("(empty)")
				case len(resp.Data) > 0:
					fmt.Println(resp.Data)
				case len(resp.Array) > 0:
					for index, rsp := range resp.Array {
						fmt.Println(fmt.Sprintf(`%d) "%s"`, index+1, rsp.Data))
					}
				}
			}
		} else {
			fmt.Println("please input redis command")
		}

		fmt.Print(ctx.Channel().RemoteAddr(), ">")
	}
}
