package amazon

import (
	"fmt"
	"time"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Packers is the struct used when returning information on running packer instances
type Packers struct {
	ToKill []ToKill
	ToSave []ToSave
}

// ToKill is the struct used when returning information about long running packer instances
type ToKill struct {
	ID string
	Sg string
	Kp string
}

// ToSave is the struct used when returning information about short running packer instances
type ToSave struct {
	ID string
	Sg string
	Kp string
}

// GetPackerInstances searches EC2 for instances which use a specified tag.
// If instances have been running for more than 45 minutes they are placed in the ToKill struct.
// If instances have been running for less than 45 minutes they are placed in the ToSave struct.
// Instances marked ToSave, will not have any of their resource removed.
func GetPackerInstances(svc *ec2.EC2, tagKey string, tagValue string) (Packers) {

	var filters []*ec2.Filter

	filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:" + tagKey),
			Values: []*string{
				aws.String(tagValue),
			},
		})
	filters = append(filters, &ec2.Filter{
		Name: aws.String("instance-state-name"),
		Values: []*string{
			aws.String("running"),
			aws.String("pending"),
		},
	})

	resp, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: filters,
	})

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	var packers Packers
	now := time.Now()
	then := now.Add(-45 * time.Minute)

	if len(resp.Reservations) == 0 {
		return packers
	}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if then.After(*inst.LaunchTime) {
				packers.ToKill = append(packers.ToKill, ToKill{
					ID: *inst.InstanceId,
					Sg: *inst.SecurityGroups[0].GroupId,
					Kp: *inst.KeyName,
				})
			} else {
				packers.ToSave = append(packers.ToSave, ToSave{
					ID: *inst.InstanceId,
					Sg: *inst.SecurityGroups[0].GroupId,
					Kp: *inst.KeyName,
				})
			}
		}
	}
	return packers
}

// GetPackerKeyPairs returns a list of all EC2 KeyPair names which start with 'packer_'.
func GetPackerKeyPairs(svc *ec2.EC2) (p []string) {

	params := &ec2.DescribeKeyPairsInput{
		Filters: []*ec2.Filter{{}}}

	resp, err := svc.DescribeKeyPairs(params)

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	for _, key := range resp.KeyPairs {
		if strings.HasPrefix(*key.KeyName, "packer_") {
			p = append(p, *key.KeyName)
		}
	}
	return p
}

// GetPackerSecurityGroups returns a list of all VPC security group IDs if their 'Name' starts with 'packer'.
func GetPackerSecurityGroups(svc *ec2.EC2) (p []string) {

	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{{}}}

	resp, err := svc.DescribeSecurityGroups(params)

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	for _, sg := range resp.SecurityGroups {
		if strings.HasPrefix(*sg.GroupName, "packer") {
			p = append(p, *sg.GroupId)
		}
	}
	return p
}

// DeletePackerSecurityGroup deletes the VPC security group provide as input.
func DeletePackerSecurityGroup(svc *ec2.EC2, sg string) () {

	params := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(sg),
	}

	fmt.Printf("Deleting security group %v\n", sg)

	_, err := svc.DeleteSecurityGroup(params)

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

// DeletePackerKeyPair deletes the EC2 KeyPair provide as input.
func DeletePackerKeyPair(svc *ec2.EC2, k string) () {

	params := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(k),
	}

	fmt.Printf("Deleting key pair %v\n", k)

	_, err := svc.DeleteKeyPair(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

// TerminatePackerInstances terminates the specified EC2 instances.
// It then uses the checkTerminatedInstances function to poll and check their state reaches 'terminated'.
func TerminatePackerInstances(svc *ec2.EC2, ids []string) (string, error) {

	params := &ec2.TerminateInstancesInput{
		InstanceIds: make([]*string, len(ids)),
	}

	for i := range params.InstanceIds {
		params.InstanceIds[i] = aws.String(ids[i])
	}

	resp, err := svc.TerminateInstances(params)

	if err != nil {
		return "", err
	}

	for _, i := range resp.TerminatingInstances {
		fmt.Printf("Instance %s is now in state %s\n", *i.InstanceId, *i.CurrentState.Name)
	}

	_, checkErr := checkTerminatedInstances(svc, ids)

	if err != nil {
		return "", checkErr
	}

	return "", nil
}

func checkTerminatedInstances(svc *ec2.EC2, ids []string) (string, error){

	params := &ec2.DescribeInstanceStatusInput{
		InstanceIds: make([]*string, len(ids)),
		IncludeAllInstances: aws.Bool(true),
	}

	for i := range params.InstanceIds {
		params.InstanceIds[i] = aws.String(ids[i])
	}

	for i := 1; i <= 20; i++ {
		fmt.Println("Polling AWS to confirm instance termination...")
		resp, err := svc.DescribeInstanceStatus(params)

		if err != nil {
			return "", err
		}

		for idx, inst := range resp.InstanceStatuses {
			if *inst.InstanceState.Name == "terminated" {
				ids = append(ids[:idx], ids[idx+1:]...)
			}
		}
		if len(ids) == 0 {
			return "", nil
		}
		time.Sleep(30 * time.Second)
	}
	return "", fmt.Errorf("Unable to successfully confirm termination of instances %s", ids)
}
