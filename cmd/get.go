// Copyright (c) arkade author(s) 2020. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"strings"

	"github.com/alexellis/arkade/pkg/env"
	"github.com/alexellis/arkade/pkg/get"
	"github.com/spf13/cobra"
)

func MakeGet() *cobra.Command {
	tools := get.MakeTools()

	var command = &cobra.Command{
		Use:          "get",
		Short:        "Get a tool from its upstream download location",
		Example:      `  arkade get`,
		SilenceUsage: true,
	}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println(arkadeGet)
			return nil
		}
		var tool *get.Tool

		if len(args) == 1 {
			for _, t := range tools {
				if t.Name == args[0] {
					tool = &t
					break
				}
			}
		}
		if tool == nil {
			return fmt.Errorf("cannot get tool: %s", args[0])
		}

		fmt.Printf("Downloading %s\n", tool.Name)

		arch, os := env.GetClientArch()
		version := ""

		downloadOut, err := get.Download(tool, strings.ToLower(os), strings.ToLower(arch), version)
		if err != nil {
			return err
		}

		fmt.Println(downloadOut)

		return nil
	}
	return command
}

const arkadeGet = `You can get various tools with arkade:
  arkade get kubectl
  arkade get faas-cli
`
