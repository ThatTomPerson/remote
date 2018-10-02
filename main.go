package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThatTomPerson/remote/scout"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "dev"
)

var (
	app = kingpin.New("remote", "A command-line chat application.")

	runCommand  = app.Command("run", "Run a command in a new container").Default()
	user        = runCommand.Flag("user", "User to ssh with").Default("ec2-user").Short('u').String()
	environment = runCommand.Flag("environment", "Enable debug mode.").Default("production").Short('e').String()
	project     = runCommand.Arg("project", "project").Required().String()
	command     = Commands(runCommand.Arg("command", "command to run").Default("bash"))

	batchCommand     = app.Command("batch", "Run a command in all of the running containers for a service")
	batchUser        = batchCommand.Flag("user", "User to ssh with").Default("ec2-user").Short('u').String()
	batchEnvironment = batchCommand.Flag("environment", "Enable debug mode.").Default("production").Short('e').String()
	batchProject     = batchCommand.Arg("project", "project").Required().String()
	batchCommands    = Commands(batchCommand.Arg("command", "command to run").Default("bash"))
)

func run(ctx context.Context) error {
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

// func FindAll

func batch(ctx context.Context) error {

	// serviceName := fmt.Sprintf("%s-%s-http", *batchProject, *batchEnvironment)

	select {
	case <-time.After(time.Second * 2):
		// all := FindAll(serviceName)
		return nil
	case <-ctx.Done():
		return errors.New("Canceled")
	}

	// serviceName := fmt.Sprintf("%s-%s-http", *batchProject, *batchEnvironment)

	// srv := scout.New()
	// s, err := srv.Service(serviceName)
	// if err != nil {
	// 	return err
	// }

	// tasks, err := s.Tasks()
	// if err != nil {
	// 	return err
	// }

	// for _, t := range tasks.Tasks {
	// 	if err != nil {
	// 		log.Println(err)
	// 	}

	//  spew.Dump(res)
	// 	return nil
	// }

	return nil
}

func main() {
	kingpin.Version(version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	errs := make(chan error)

	go func() {
		switch kingpin.MustParse(app.Parse(os.Args[1:])) {
		case runCommand.FullCommand():
			errs <- run(ctx)

		case batchCommand.FullCommand():
			errs <- batch(ctx)
		}

		errs <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("Canceled")
	case err := <-errs:
		if err != nil {
			log.Println(err)
		}
	}
}
