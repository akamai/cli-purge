package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/fatih/color"
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
	_, in_cli := os.LookupEnv("AKAMAI_CLI")

	app_name := "akamai-purge"
	if in_cli {
		app_name = "akamai purge"
	}

	app := cli.NewApp()
	app.Name = app_name
	app.HelpName = app_name
	app.Usage = "Purge Content from the Edge"
	app.Version = "0.1.0"
	app.Copyright = "Copyright (C) Akamai Technologies, Inc"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "section",
			Usage: "Section of the credentials file",
			Value: "default",
		},
	}

	flags := []cli.Flag{
		cli.StringSliceFlag{
			Name:  "cpcode",
			Usage: "CPCode(s) to Purge",
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
			Action:    invalidate,
			Flags:     flags,
		},
		{
			Name:      "delete",
			Usage:     "Delete content",
			ArgsUsage: "[URL...] or [CP Codes...]",
			Action:    delete,
			Flags:     flags,
		},
	}

	app.Run(os.Args)
}

func invalidate(c *cli.Context) error {
	return purge("invalidate", c)
}

func delete(c *cli.Context) error {
	return purge("delete", c)
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

	client, err := edgegrid.NewClient(nil, &config)
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

	res, err := client.PostJSON(url, body)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	if res.IsError() {
		fmt.Println("... [" + color.RedString("FAIL") + "]")
		return cli.NewExitError(edgegrid.NewAPIError(res).Error(), 1)
	}

	purge := &CCUv3Purge{}
	if err = res.BodyJSON(purge); err != nil {
		fmt.Println("... [" + color.RedString("FAIL") + "]")
		return cli.NewExitError(err.Error(), 1)
	}

	fmt.Println("... [" + color.GreenString("OK") + "] (URLs: " + color.BlueString("%d", len(body.Objects)) + ", ETA: " + color.BlueString("%d seconds", purge.EstimatedSeconds) + ")")

	return nil
}
