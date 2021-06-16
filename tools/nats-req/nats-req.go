// Copyright 2012-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/ascii85"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NOTE: Can test with demo servers.
// nats-req -s demo.nats.io <subject> <msg>
// nats-req -s demo.nats.io:4443 <subject> <msg> (TLS version)

func usage() {
	log.Printf("Usage: nats-req [-s server] [-creds file] <subject> <msg>\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	var userCreds = flag.String("creds", "", "User Credentials File")
	var showHelp = flag.Bool("h", false, "Show help message")
	var stdout = flag.Bool("o", false, "Write raw response to stdout")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	args := flag.Args()
	if len(args) < 2 {
		showUsageAndExit(1)
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Sample Requestor")}

	// Use UserCredentials
	if *userCreds != "" {
		opts = append(opts, nats.UserCredentials(*userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	subj := args[0]

	if len(args) > 1 {
		for _, arg := range args[1:] {
			text, payload, err := prepare(arg)
			if err != nil {
				log.Fatal(err)
			}
			msg, err := nc.Request(subj, payload, 2*time.Second)
			if err != nil {
				if nc.LastError() != nil {
					log.Fatalf("%v for request", nc.LastError())
				}
				log.Fatalf("%v for request", err)
			}
			logPub(*stdout, text, subj, payload, msg)
		}
	} else {
		payload, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		msg, err := nc.Request(subj, payload, 2*time.Second)
		if err != nil {
			if nc.LastError() != nil {
				log.Fatalf("%v for request", nc.LastError())
			}
			log.Fatalf("%v for request", err)
		}
		logPub(*stdout, false, subj, payload, msg)
	}

}

func logPub(stdout, text bool, subj string, payload []byte, msg *nats.Msg) {
	if !text {
		payload = []byte(hex.EncodeToString(payload))
	}
	log.Printf("Published [%s] : '%s'", subj, payload)
	if stdout {
		log.Printf("Received  [%v] : %d bytes", msg.Subject, len(msg.Data))
		_, _ = io.Copy(os.Stdout, bytes.NewReader(msg.Data))
	} else {
		log.Printf("Received  [%v] : '%s'", msg.Subject, string(msg.Data))
	}
}

const (
	hexPrefix = "hex:"
	b64Prefix = "b64:"
	a85Prefix = "a85:"
)

func prepare(arg string) (bool, []byte, error) {
	if strings.HasPrefix(arg, "@") {
		bs, err := os.ReadFile(strings.TrimPrefix(arg, "@"))
		return false, bs, err
	}
	if strings.HasPrefix(arg, hexPrefix) {
		bs, err := hex.DecodeString(strings.TrimPrefix(arg, hexPrefix))
		return false, bs, err
	}
	if strings.HasPrefix(arg, b64Prefix) {
		bs, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(arg, b64Prefix))
		return false, bs, err
	}
	if strings.HasPrefix(arg, a85Prefix) {
		var bs []byte
		_, err := ascii85.NewDecoder(strings.NewReader(strings.TrimPrefix(arg, a85Prefix))).Read(bs)
		return false, bs, err
	}
	return true, []byte(arg), nil
}
