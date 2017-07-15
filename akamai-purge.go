/*
 * Copyright 2017 Akamai Technologies, Inc
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

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type CCUv3PurgeBody struct {
	Objects []string `json:"objects""`
}

type CCUv3Purge struct {
	PurgeID          string `json:"purgeId"`
	EstimatedSeconds int    `json:"estimatedSeconds"`
	HTTPStatus       int    `json:"httpStatus"`
	Detail           string `json:"detail"`
	SupportID        string `json:"supportId"`
}

func main() {
	setCliTemplates()

	_, in_cli := os.LookupEnv("AKAMAI_CLI")

	app_name := "akamai-purge"
	if in_cli {
		app_name = "akamai purge"
	}

	app := cli.NewApp()
	app.Name = app_name
	app.HelpName = app_name
	app.Usage = "A CLI for Purge"
	app.Description = "Purge Content from the Edge. URLs/CPCodes may be specified as a list of arguments, or piped in via STDIN"
	app.Version = "0.1.0"
	app.Copyright = "Copyright (C) Akamai Technologies, Inc"
	app.Authors = []cli.Author{
		{
			Name:  "Davey Shafik",
			Email: "dshafik@akamai.com",
		},
		{
			Name:  "Akamai Developer",
			Email: "https://developer.akamai.com",
		},
	}

	dir, _ := homedir.Dir()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "edgerc",
			Usage:  "Location of the credentials file",
			Value:  dir,
			EnvVar: "AKAMAI_EDGERC",
		},
		cli.StringFlag{
			Name:   "section",
			Usage:  "Section of the credentials file",
			Value:  "default",
			EnvVar: "AKAMAI_EDGERC_SECTION",
		},
	}

	cmdFlags := []cli.Flag{
		cli.BoolFlag{
			Name:  "cpcode",
			Usage: "Purge by CPCode instead (beta)",
		},
		cli.BoolFlag{
			Name:  "production",
			Usage: "(default)",
		},
		cli.BoolFlag{
			Name: "staging",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "invalidate",
			Usage:     "Invalidate content",
			ArgsUsage: "[URL...] or [CP Codes...]",
			Action:    cmdInvalidate,
			Flags:     cmdFlags,
		},
		{
			Name:      "delete",
			Usage:     "Delete content",
			ArgsUsage: "[URL...] or [CP Codes...]",
			Action:    cmdDelete,
			Flags:     cmdFlags,
		},
		{
			Name:   "list",
			Usage:  "List commands",
			Action: cmdList,
		},
	}

	app.Run(os.Args)
}

func cmdInvalidate(c *cli.Context) error {
	return purge("invalidate", c)
}

func cmdDelete(c *cli.Context) error {
	return purge("delete", c)
}

func cmdList(c *cli.Context) {
	for _, command := range c.App.Commands {
		fmt.Println(command.HelpName + "  ")
	}
}

func purge(purgeType string, c *cli.Context) error {
	fmt.Print("Purging...")
	purgeBy := "url"

	network := "production"
	if c.IsSet("staging") {
		network = "staging"
	}

	if c.IsSet("cpcode") {
		purgeBy = "cpcode"
	}

	config, err := edgegrid.Init("", c.GlobalString("section"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	url := fmt.Sprintf(
		"/ccu/v3/%s/%s/%s",
		purgeType,
		purgeBy,
		network,
	)

	body := &CCUv3PurgeBody{}

	if c.Args().Present() {
		for _, object := range c.Args() {
			body.Objects = append(body.Objects, object)
		}
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			url := s.Text()
			if url == "" {
				continue
			}
			body.Objects = append(body.Objects, url)
		}

		if len(body.Objects) == 0 {
			fmt.Println("... [" + color.RedString("FAIL") + "]")
			return cli.NewExitError("You must specify at least one URL to purge.", 1)
		}
	}

	req, err := client.NewJSONRequest(config, "POST", url, body)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	res, err := client.Do(config, req)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if client.IsError(res) {
		fmt.Println("... [" + color.RedString("FAIL") + "]")
		return cli.NewExitError(client.NewAPIError(res).Error(), 1)
	}

	purge := &CCUv3Purge{}
	if err = client.BodyJSON(res, purge); err != nil {
		fmt.Println("... [" + color.RedString("FAIL") + "]")
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Println("... [" + color.GreenString("OK") + "] (URLs: " + color.BlueString("%d", len(body.Objects)) + ", ETA: " + color.BlueString("%d seconds", purge.EstimatedSeconds) + ")")

	return nil
}

func setCliTemplates() {
	cli.AppHelpTemplate = "" +
		color.YellowString("Usage: \n") +
		color.BlueString("	 {{if .UsageText}}"+
			"{{.UsageText}}"+
			"{{else}}"+
			"{{.HelpName}} "+
			"{{if .VisibleFlags}}[global flags]{{end}}"+
			"{{if .Commands}} command [command flags]{{end}} "+
			"{{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}"+
			"\n\n{{end}}") +

		"{{if .Description}}\n\n" +
		color.YellowString("Description:\n") +
		"   {{.Description}}" +
		"\n\n{{end}}" +

		"{{if .VisibleCommands}}" +
		color.YellowString("Built-In Commands:\n") +
		"{{range .VisibleCategories}}" +
		"{{if .Name}}" +
		"\n{{.Name}}\n" +
		"{{end}}" +
		"{{range .VisibleCommands}}" +
		`   {{join .Names ", "}}{{"\n"}}` +
		"{{end}}" +
		"{{end}}" +
		"\n{{end}}" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Global Flags:\n") +
		"{{range $index, $option := .VisibleFlags}}" +
		"{{if $index}}\n{{end}}" +
		"   {{$option}}" +
		"{{end}}" +
		"\n\n{{end}}" +

		"{{if len .Authors}}" +
		color.YellowString("Author{{with $length := len .Authors}}{{if ne 1 $length}}s{{end}}{{end}}:\n") +
		"{{range $index, $author := .Authors}}{{if $index}}\n{{end}}" +
		"   {{$author}}" +
		"{{end}}" +
		"\n\n{{end}}" +

		"{{if .Copyright}}" +
		color.YellowString("Copyright:\n") +
		"   {{.Copyright}}" +
		"{{end}}\n"

	cli.CommandHelpTemplate = "" +
		color.YellowString("Name: \n") +
		"   {{.HelpName}} - {{.Usage}}\n\n" +

		color.YellowString("Usage: \n") +
		color.BlueString("   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}\n\n") +

		"{{if .Category}}" +
		color.YellowString("Type: \n") +
		"   {{.Category}}\n\n{{end}}" +

		"{{if .Description}}" +
		color.YellowString("Description: \n") +
		"   {{.Description}}\n\n{{end}}" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Flags: \n") +
		"{{range .VisibleFlags}}   {{.}}\n{{end}}{{end}}"

	cli.SubcommandHelpTemplate = "" +
		color.YellowString("Name: \n") +
		"   {{.HelpName}} - {{.Usage}}\n\n" +

		color.YellowString("Usage: \n") +
		color.BlueString("   {{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}\n\n") +

		color.YellowString("Commands:\n") +
		"{{range .VisibleCategories}}" +
		"{{if .Name}}" +
		"{{.Name}}:" +
		"{{end}}" +
		"{{range .VisibleCommands}}" +
		`{{join .Names ", "}}{{"\t"}}{{.Usage}}` +
		"{{end}}\n\n" +
		"{{end}}" +

		"{{if .VisibleFlags}}" +
		color.YellowString("Flags:\n") +
		"{{range .VisibleFlags}}{{.}}\n{{end}}{{end}}"
}
