package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ThatTomPerson/remote/scout"
	"gopkg.in/alecthomas/kingpin.v2"
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

var (
	version = "dev"
)

var (
	user        = kingpin.Flag("user", "User to ssh with").Default("ec2-user").Short('u').String()
	environment = kingpin.Flag("environment", "Enable debug mode.").Default("production").Short('e').String()
	project     = kingpin.Arg("project", "project").Required().String()
	command     = Commands(kingpin.Arg("command", "command to run").Default("bash"))
)

func run() error {
	kingpin.Version(version)
	kingpin.Parse()

	serviceName := fmt.Sprintf("%s-%s-http", *project, *environment)
	srv := scout.New()

	s, err := srv.Service(serviceName)
	if err != nil {
		return fmt.Errorf("can not find service %s: %v", serviceName, err)
	}

	log.Printf("finding %s\n", *s.Service.ServiceName)

	td, err := srv.TaskDef(s.Service.TaskDefinition)
	if err != nil {
		return fmt.Errorf("can not find task def %s: %v", serviceName, err)
	}

	t, err := s.Tasks()
	if err != nil {
		return fmt.Errorf("can not find tasks for service %s: %v", serviceName, err)
	}

	ids, err := t.InstanceIds()
	if err != nil {
		return fmt.Errorf("no instances running service %s: %v", serviceName, err)
	}

	i, err := srv.Instance(ids[0])
	if err != nil {
		return fmt.Errorf("failed getting instance %s: %v", *ids[0], err)
	}

	// taskArn := *t.Tasks[0].TaskDefinitionArn
	address := fmt.Sprintf("%s@%s", *user, *i.PrivateIpAddress)
	log.Printf("ssh %s", address)

	def := td.ContainerDefinitions[0]

	envString := ""

	for _, e := range def.Environment {
		envString += fmt.Sprintf(" -e %s=\"%s\"", *e.Name, *e.Value)
	}

	log.Printf("docker run %s %s", *def.Image, command.String())

	cmd := fmt.Sprintf("sudo docker run --rm -it%s %s %s", envString, *def.Image, command.String())
	child := exec.Command("ssh", address, "-t", cmd)

	child.Stdout = os.Stdout
	child.Stdin = os.Stdin
	child.Stderr = os.Stderr

	return child.Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
