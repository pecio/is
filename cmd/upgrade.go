/*
Copyright © 2025 Raul Pedroche <pedroche@me.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"log"

	"github.com/pecio/is/docker"
	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade a container image to most recent with same tag",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if (all && containerName != "") || (!all && containerName == "") {
			log.Fatalln("Use either --container or --all, but not both")
		}
		err := docker.Upgrade(containerName, pullOnly)
		if err != nil {
			log.Fatalln(err)
		}
	},
}

var containerName string
var all bool
var pullOnly bool

func init() {
	rootCmd.AddCommand(upgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	upgradeCmd.Flags().StringVarP(&containerName, "container", "c", "", "Container name or id")
	upgradeCmd.Flags().BoolVarP(&all, "all", "a", false, "Upgrade all running containers")
	upgradeCmd.Flags().BoolVarP(&pullOnly, "pullonly", "p", false, "Only pull image, do not recreate container")
}
