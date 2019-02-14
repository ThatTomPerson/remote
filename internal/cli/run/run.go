package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apex/log"

	"github.com/thattomperson/remote/internal/cli/root"
	"github.com/thattomperson/remote/scout"
	"github.com/tj/kingpin"
)

type commands []string

func (i *commands) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func (i *commands) String() string {
	return strings.Join(*i, " ")
}

func (i *commands) IsCumulative() bool {
	return true
}

func Commands(s kingpin.Settings) (target *commands) {
	target = new(commands)
	s.SetValue((*commands)(target))
	return
}

func init() {
	cmd := root.Command("run", "Run a command in a new container.").Default()
	cmd.Example(`remote run acg bash`, "Run bash in the service called acg.")

	srv := scout.New()

	user := cmd.Flag("user", "User to ssh with").Short('u').String()
	environment := cmd.Flag("environment", "Enable debug mode.").Default("production").Short('e').String()
	project := cmd.Arg("project", "project").Required().String()
	command := Commands(cmd.Arg("command", "command to run").Default("bash"))

	cmd.Action(func(_ *kingpin.ParseContext) error {
		if *user == "" {
			*user = srv.DefaultUser()
		}

		serviceName := fmt.Sprintf("%s-%s-http", *project, *environment)

		log.Infof("looking for %s", serviceName)

		s, err := srv.Service(serviceName)
		if err != nil {
			return fmt.Errorf("can not find service %s: %v", serviceName, err)
		}

		log.Infof("found %s", *s.Service.TaskDefinition)

		td, err := srv.TaskDef(s.Service.TaskDefinition)
		if err != nil {
			log.Errorf("can not find task def %s: %v", serviceName, err)
			return err
		}

		t, err := s.Tasks()
		if err != nil {
			log.Errorf("can not find tasks for service %s: %v", serviceName, err)
			return err
		}

		ids, err := t.InstanceIds()
		if err != nil {
			log.Errorf("no instances running service %s: %v", serviceName, err)
			return err
		}

		i, err := srv.Instance(ids[0])
		if err != nil {
			log.Errorf("failed getting instance %s: %v", *ids[0], err)
			return err
		}

		// taskArn := *t.Tasks[0].TaskDefinitionArn
		address := fmt.Sprintf("%s@%s", *user, *i.PrivateIpAddress)
		log.Infof("ssh %s", address)

		def := td.ContainerDefinitions[0]

		envString := ""

		for _, e := range append(def.Environment, srv.Credentials()...) {
			envString += fmt.Sprintf(" -e %s=\"%s\"", *e.Name, *e.Value)
		}

		log.Debugf("docker run %s %s", *def.Image, command.String())

		cmd := fmt.Sprintf("sudo docker run --rm --init -it%s --ulimit nofile=8192 %s %s", envString, *def.Image, command.String())
		child := exec.Command("ssh", address, "-t", cmd)

		child.Stdout = os.Stdout
		child.Stdin = os.Stdin
		child.Stderr = os.Stderr

		log.Infof("Running %s", command.String())
		log.Debug(cmd)
		if err = child.Run(); err != nil {
			log.Error(err.Error())
		}
		return nil
	})
}
