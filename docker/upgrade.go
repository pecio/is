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
package docker

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func Upgrade(containerName string, pullOnly bool) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		return err
	}

	for _, cnt := range containers {
		var name string

		if containerName != "" {
			found := false
			for _, name := range cnt.Names {
				if containerName == strings.TrimLeft(name, "/") {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if cnt.State != "running" {
				return errors.New("this will only work with running containers")
			}
			name = containerName
		} else {
			name = strings.TrimLeft(cnt.Names[0], "/")
		}

		log.Printf("Upgrading %v", name)
		err = doUpgrade(cli, cnt, pullOnly)
		if err != nil {
			return err
		}
	}

	return nil
}

func doUpgrade(cli *client.Client, cnt types.Container, pullOnly bool) error {
	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, cnt.Image, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	buf := bufio.NewScanner(reader)
	// Read 1st line and ignore
	buf.Scan()
	// Read second line, might be:
	// {"status":"Digest:..."} --> Already latest
	// {"status":"Pulling..."} --> Pulling new version
	buf.Scan()

	match, err := regexp.MatchString("{\"status\":\"Digest: .*", buf.Text())
	if err != nil {
		return err
	}

	if match {
		log.Printf("No new image for %v\n", cnt.Image)
		return nil
	}

	log.Printf("Pulling new image for %v", cnt.Image)

	for buf.Scan() {
	}
	log.Printf("Pulled %v", cnt.Image)

	if pullOnly {
		return nil
	}

	currentContainer, err := cli.ContainerInspect(ctx, cnt.ID)
	if err != nil {
		return err
	}

	log.Printf("Recreating %v", currentContainer.Name)
	err = cli.ContainerStop(ctx, os.Args[1], container.StopOptions{})
	if err != nil {
		return err
	}

	log.Printf("Stopped %v", currentContainer.Name)

	err = cli.ContainerRemove(ctx, os.Args[1], container.RemoveOptions{RemoveVolumes: false})
	if err != nil {
		return err
	}

	log.Printf("Removed %v", currentContainer.Name)

	// We have to build the NetworkConfig struct
	createResponse, err := cli.ContainerCreate(ctx, currentContainer.Config, currentContainer.HostConfig, &network.NetworkingConfig{EndpointsConfig: currentContainer.NetworkSettings.Networks}, nil, currentContainer.Name)
	if err != nil {
		return err
	}

	log.Printf("Created %v with ID %v", currentContainer.Name, createResponse.ID)

	err = cli.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	log.Printf("Started %v", currentContainer.Name)

	return nil
}
