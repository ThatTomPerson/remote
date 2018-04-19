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
	"os"
	"os/exec"
	"strings"

	"github.com/ThatTomPerson/remote/internal/app/scout"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := scout.New()

		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}

		s, err := srv.Service(env)
		if err != nil {
			return err
		}
		t, err := s.Tasks()
		if err != nil {
			return err
		}

		ids, err := t.InstanceIds()
		if err != nil {
			return err
		}

		i, err := srv.Instance(ids[0])
		if err != nil {
			return err
		}

		taskArn := *t.Tasks[0].TaskDefinitionArn
		ipAddress := *i.PublicIpAddress

		user := viper.GetString("username")
		address := user + "@" + ipAddress
		command := "sh -c \"bash || sh\""
		if len(args) > 0 {
			command = strings.Join(args, " ")
		}

		taskName := taskArn[strings.Index(taskArn, "/"):]
		taskRevision := taskName[strings.Index(taskName, ":")+1:]

		command = fmt.Sprintf(
			"sudo docker exec -it $(sudo docker ps | grep ecs-%s-%s-php | awk '{print $1}' | head -n1) env TERM=screen %s",
			env,
			taskRevision,
			command,
		)

		log.Println(address)
		log.Println(command)

		child := exec.Command("ssh", address, "-t", command)

		child.Stdout = os.Stdout
		child.Stdin = os.Stdin
		child.Stderr = os.Stderr

		child.Run()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().StringP("environment", "e", "develop", "The Environment to remote into")
}
