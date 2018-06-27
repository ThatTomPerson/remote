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
	"os"

	"github.com/ThatTomPerson/remote/internal/pkg/scout"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
		// prg, err := cmd.Flags().GetString("project")
		// if err != nil {
		// 	return err
		// }

		s, err := srv.Service(env)
		if err != nil {
			return fmt.Errorf("can not find service %s: %v", env, err)
		}
		t, err := s.Tasks()
		if err != nil {
			return fmt.Errorf("can not find tasks for service %s: %v", env, err)
		}

		ids, err := t.InstanceIds()
		if err != nil {
			return fmt.Errorf("no instances running service %s: %v", env, err)
		}

		i, err := srv.Instance(ids[0])
		if err != nil {
			return fmt.Errorf("failed getting instance %s: %v", *ids[0], err)
		}

		// taskArn := *t.Tasks[0].TaskDefinitionArn
		ipAddress := *i.PublicIpAddress

		// user :=
		// address := user + "@" + ipAddress
		// command := "sh"
		// if len(args) > 0 {
		// 	command = strings.Join(args, " ")
		// }

		// taskName := taskArn[strings.Index(taskArn, "/"):]
		// taskRevision := taskName[strings.Index(taskName, ":")+1:]

		sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			return fmt.Errorf("unable to open ssh-agent socket: %v", err)
		}

		// Create the Signer for this private key.

		a := agent.NewClient(sock)
		signers, err := a.Signers()
		if err != nil {
			return fmt.Errorf("unable to get signers: %v", err)
		}

		config := &ssh.ClientConfig{
			User: viper.GetString("username"),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signers...),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", ipAddress), config)
		if err != nil {
			log.Fatal("Failed to dial: ", err)
		}

		session, err := client.NewSession()
		if err != nil {
			log.Fatal("Failed to create session: ", err)
		}
		defer session.Close()

		// Once a Session is created, you can execute a single command on
		// the remote side using the Run method.

		session.Stdout = os.Stdout
		if err := session.Run("/usr/bin/whoami"); err != nil {
			log.Fatal("Failed to run: " + err.Error())
		}

		// command = fmt.Sprintf(
		// 	"sudo docker exec -it $(sudo docker ps | grep ecs-%s-%s-php | awk '{print $1}' | head -n1) env TERM=screen %s",
		// 	env,
		// 	taskRevision,
		// 	command,
		// )

		// log.Println(address)
		// log.Println(command)

		// child := exec.Command("ssh", address, "-t", command)

		// child.Stdout = os.Stdout
		// child.Stdin = os.Stdin
		// child.Stderr = os.Stderr

		// child.Run()

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
