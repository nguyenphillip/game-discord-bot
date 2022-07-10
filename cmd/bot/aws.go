package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
)

func createEC2Client(args map[string]string) *ec2.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(args["region"]),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(args["aws_access_key_id"], args["aws_secret_access_key"], "")))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	return ec2.NewFromConfig(cfg)
}

// EC2DescribeInstancesAPI defines the interface for the DescribeInstances function.
// We use this interface to test the function using a mocked service.
type EC2DescribeInstancesAPI interface {
	DescribeInstances(ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// GetInstances retrieves information about your Amazon Elastic Compute Cloud (Amazon EC2) instances.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a DescribeInstancesOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to DescribeInstances.
func GetInstances(c context.Context, api EC2DescribeInstancesAPI, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return api.DescribeInstances(c, input)
}

func DescribeInstancesCmd(args map[string]string, instanceID string) ([]map[string]interface{}, error) {

	client := createEC2Client(args)

	input := &ec2.DescribeInstancesInput{}
	if instanceID != "" {
		input = &ec2.DescribeInstancesInput{
			InstanceIds: []string{
				instanceID,
			},
		}
	}

	result, err := GetInstances(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error retrieving information about your Amazon EC2 instances:")
		fmt.Println(err)
		return nil, err
	}

	var returnInstances []map[string]interface{}

	for _, r := range result.Reservations {

		// fmt.Println("Reservation ID: " + *r.ReservationId)
		// fmt.Println("Instance IDs:")
		for _, i := range r.Instances {
			instance := make(map[string]interface{})
			for _, t := range i.Tags {
				if *t.Key == "Name" {
					instance["NAME"] = *t.Value
				}
			}

			instance["IP"] = ""
			if i.PublicIpAddress != nil {
				instance["IP"] = *i.PublicIpAddress
			}
			instance["ID"] = *i.InstanceId
			instance["STATUS"] = strings.ToUpper(string(i.State.Name))

			returnInstances = append(returnInstances, instance)
		}

		// fmt.Println("")
	}
	return returnInstances, nil

}

// EC2StartInstancesAPI defines the interface for the StartInstances function.
// We use this interface to test the function using a mocked service.
type EC2StartInstancesAPI interface {
	StartInstances(ctx context.Context,
		params *ec2.StartInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
}

// StartInstance starts an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a StartInstancesOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to StartInstances.
func StartInstance(c context.Context, api EC2StartInstancesAPI, input *ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error) {
	resp, err := api.StartInstances(c, input)

	var apiErr smithy.APIError
	f := false
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
		fmt.Println("User has permission to start an instance.")
		input.DryRun = &f
		return api.StartInstances(c, input)
	}

	return resp, err
}

func StartInstancesCmd(args map[string]string, instanceID string) error {

	if instanceID == "" {
		fmt.Println("You must supply an instance ID (-i INSTANCE-ID")
		return errors.New("error instance ID must not be empty")
	}

	client := createEC2Client(args)

	t := true
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{
			instanceID,
		},
		DryRun: &t,
	}

	_, err := StartInstance(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error starting the instance")
		fmt.Println(err)
		return err
	}

	fmt.Println("Started instance with ID " + instanceID)
	return nil
}

// EC2StopInstancesAPI defines the interface for the StopInstances function.
// We use this interface to test the function using a mocked service.
type EC2StopInstancesAPI interface {
	StopInstances(ctx context.Context,
		params *ec2.StopInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
}

// StopInstance stops an Amazon Elastic Compute Cloud (Amazon EC2) instance.
// Inputs:
//     c is the context of the method call, which includes the AWS Region.
//     api is the interface that defines the method call.
//     input defines the input arguments to the service call.
// Output:
//     If success, a StopInstancesOutput object containing the result of the service call and nil.
//     Otherwise, nil and an error from the call to StopInstances.
func StopInstance(c context.Context, api EC2StopInstancesAPI, input *ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error) {
	resp, err := api.StopInstances(c, input)

	var apiErr smithy.APIError
	f := false
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DryRunOperation" {
		fmt.Println("User has permission to stop instances.")
		input.DryRun = &f
		return api.StopInstances(c, input)
	}

	return resp, err
}

func StopInstancesCmd(args map[string]string, instanceID string) error {
	if instanceID == "" {
		fmt.Println("You must supply an instance ID (-i INSTANCE-ID")
		return errors.New("error instance ID must not be empty")
	}

	client := createEC2Client(args)
	t := true
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{
			instanceID,
		},
		DryRun: &t,
	}

	_, err := StopInstance(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error stopping the instance")
		fmt.Println(err)
		return err
	}

	fmt.Println("Stopped instance with ID " + instanceID)
	return nil
}
