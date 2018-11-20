// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "cfa",
	Short: "Cloudflare Access. Connect to a site behind Cloudflare Access",
	Run: func(cmd *cobra.Command, args []string) {
		conn(cmd, args)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func conn(cmd *cobra.Command, args []string) {
	var authChan chan string
	if len(args) != 1 {
		log.Fatal("Invalid number of arguments. Only provide a URL")
	}

	connUrl := args[0]
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(connUrl)
	if err != nil {
		log.Fatalf("Could not fetch connection url: %s", err)
	}
	if resp.StatusCode != 302 {
		log.Fatalf("Site not on cloudflare access")
	}
	redirectLocation := resp.Header.Get("location")
	redirectURL, err := url.Parse(redirectLocation)
	if err != nil {
		log.Fatalf("Could not parse redirect url: %s", err)
	}

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Could not open listener: %s", err)
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r)
			authChan <- "yo"
		})
		if err != nil {
			panic(err)
		}
		http.Serve(listener, nil)
	}()

	fmt.Println(listener.Addr().String())
	redirectURL.Path = "/cdn-cgi/access/login/" + listener.Addr().String()

	fmt.Println(redirectURL.String())
	//
	err = exec.Command("open", redirectURL.String()).Run()
	if err != nil {
		log.Fatalf("Could not open redirect url: %s", err)
	}
	fmt.Println(<-authChan)
	fmt.Println("YO")
}
