/*
 * Copyright 2018. Akamai Technologies, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"os"

	akamai "github.com/akamai/cli-common-golang"
)

const (
	VERSION = "1.0.1"
)

func main() {
	akamai.CreateApp(
		"purge",
		"A CLI for Purge",
		"Purge Content from the Edge. URLs/CPCodes/Tags may be specified as a list of arguments, or piped in via STDIN",
		VERSION,
		"ccu",
		commandLocator,
	)

	akamai.App.Run(os.Args)
}
