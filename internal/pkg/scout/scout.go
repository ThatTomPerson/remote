package scout

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// Service wraps ecs.Service to add fuctionality
type Service struct {
	scout   *Scout
	Service *ecs.Service
}

// Tasks wraps []*ecs.Task to add fuctionality
type Tasks struct {
	scout *Scout
	Tasks []*ecs.Task
}

func (s *Scout) TaskDef(def *string) (*ecs.TaskDefinition, error) {
	res, err := s.ECS.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: def,
	})

	if err != nil {
		return nil, err
	}

	return res.TaskDefinition, nil
}

// InstanceIds the instances that are running these tasks
func (t *Tasks) InstanceIds() ([]*string, error) {

	var arns []*string

	for _, task := range t.Tasks {
		arns = append(arns, task.ContainerInstanceArn)
	}

	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            t.scout.Cluster,
		ContainerInstances: arns,
	}

	output, err := t.scout.ECS.DescribeContainerInstances(input)
	if err != nil {
		return nil, err
	}

	var ids []*string

	for _, i := range output.ContainerInstances {
		ids = append(ids, i.Ec2InstanceId)
	}

	return ids, nil
}

// TaskArns arns of the tasks running in this service
func (s *Service) TaskArns() ([]*string, error) {
	input := &ecs.ListTasksInput{
		Cluster:     s.scout.Cluster,
		ServiceName: s.Service.ServiceName,
	}

	output, err := s.scout.ECS.ListTasks(input)
	if err != nil {
		return nil, err
	}

	return output.TaskArns, nil
}

// Tasks get the Tasks from the service
func (s *Service) Tasks() (*Tasks, error) {
	a, err := s.TaskArns()
	if err != nil {
		return nil, err
	}

	input := &ecs.DescribeTasksInput{
		Cluster: s.scout.Cluster,
		Tasks:   a,
	}

	output, err := s.scout.ECS.DescribeTasks(input)
	if err != nil {
		return nil, err
	}

	return &Tasks{
		scout: s.scout,
		Tasks: output.Tasks,
	}, nil
}

// Scout functions for discovering aws resources
type Scout struct {
	Cluster *string
	ECS     *ecs.ECS
	EC2     *ec2.EC2
}

// New return a new Scout
func New() *Scout {
	sess := session.Must(session.NewSession())
	config := &aws.Config{Region: aws.String("ap-southeast-2")}

	return &Scout{
		Cluster: aws.String("acg-development"),
		ECS:     ecs.New(sess, config),
		EC2:     ec2.New(sess, config),
	}
}

// Instances get instances from ids
func (s *Scout) Instances(ids []*string) ([]*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}
	output, err := s.EC2.DescribeInstances(input)
	if err != nil {
		return nil, err
	}

	var is []*ec2.Instance

	for _, r := range output.Reservations {
		for _, i := range r.Instances {
			is = append(is, i)
		}
	}
	return is, nil
}

// Instance just one instance
func (s *Scout) Instance(id *string) (*ec2.Instance, error) {
	is, err := s.Instances([]*string{id})
	if err != nil {
		return nil, err
	}

	if len(is) == 0 {
		return nil, fmt.Errorf("No instance found")
	}

	return is[0], nil
}

// Service find an ecs service
func (s *Scout) Service(name string) (*Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  s.Cluster,
		Services: []*string{&name},
	}

	output, err := s.ECS.DescribeServices(input)
	if err != nil {
		return nil, err
	}

	for _, v := range output.Services {
		return &Service{
			scout:   s,
			Service: v,
		}, nil
	}

	return nil, fmt.Errorf("No services found")
}
