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
	"bufio"
	"fmt"
	"os"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/ccu-v3"
	akamai "github.com/akamai/cli-common-golang"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func purge(purgeType string, c *cli.Context) error {
	akamai.StartSpinner("Purging...", fmt.Sprintf("Purging...... [%s]", color.GreenString("OK")))

	purgeByType := "url"

	network := "production"
	if c.IsSet("staging") {
		network = "staging"
	}

	if c.IsSet("cpcode") {
		purgeByType = "cpcode"
	}

	if c.IsSet("tag") {
		purgeByType = "tag"
	}

	config, err := akamai.GetEdgegridConfig(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	ccu.Init(config)

	objects := make([]string, 0)
	if c.Args().Present() {
		objects = c.Args()
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			url := s.Text()
			if url == "" {
				continue
			}
			objects = append(objects, url)
		}

		if len(objects) == 0 {
			akamai.StopSpinnerFail()
			return cli.NewExitError(color.RedString("You must specify at least one %s to purge.", purgeByType), 1)
		}
	}

	purge := ccu.NewPurge(objects)
	var res *ccu.PurgeResponse
	if purgeType == "invalidate" {
		res, err = purge.Invalidate((ccu.PurgeTypeValue)(purgeByType), (ccu.NetworkValue)(network))
	} else {
		res, err = purge.Delete((ccu.PurgeTypeValue)(purgeByType), (ccu.NetworkValue)(network))
	}

	if err != nil {
		akamai.StopSpinnerFail()
		return cli.NewExitError(err.Error(), 1)
	}

	akamai.StopSpinnerOk()
	var plural string
	if len(purge.Objects) > 0 {
		plural = "s"
	}

	fmt.Fprintf(
		akamai.App.ErrWriter,
		"Purged %s object%s (ETA: %s seconds)\n",
		color.BlueString("%d", len(purge.Objects)),
		plural,
		color.BlueString("%d", res.EstimatedSeconds),
	)

	return nil
}