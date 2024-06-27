package lightsailP

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/lormars/octohunter/internal/logger"
)

type proxyLocation struct {
	proxy    string
	location string
}

var ipProxyMap = make(map[string]*proxyLocation)

func GetIP(proxy, location string) string {
	svc, instanceName := fconfig(proxy, location)
	// Get the instance's public IP address
	ip := getInstancePublicIP(svc, instanceName)
	ipProxyMap[ip] = &proxyLocation{proxy: proxy, location: location}
	logger.Infof("The public IP address of the instance %s is %s\n", instanceName, ip)
	return ip
}

func ReGetIp(ipPort string) string {
	ip := strings.Split(ipPort, ":")[0]
	name := ipProxyMap[ip].proxy
	location := ipProxyMap[ip].location
	svc, instanceName := fconfig(name, location)
	rebootInstance(svc, instanceName)
	waitForInstance(svc, instanceName, "running")
	newIP := getInstancePublicIP(svc, instanceName)
	return newIP

}

func fconfig(proxy, location string) (*lightsail.Client, string) {
	// Load the default AWS configuration (credentials from IAM role)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(location))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an AWS Lightsail client
	svc := lightsail.NewFromConfig(cfg)

	// Define the instance name
	instanceName := proxy
	return svc, instanceName
}

func rebootInstance(svc *lightsail.Client, instanceName string) {
	// Stop the instance
	stopInstance(svc, instanceName)
	waitForInstance(svc, instanceName, "stopped")
	// Start the instance
	startInstance(svc, instanceName)
}

func stopInstance(svc *lightsail.Client, instanceName string) {
	input := &lightsail.StopInstanceInput{
		InstanceName: aws.String(instanceName),
	}

	_, err := svc.StopInstance(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to stop instance, %v", err)
	}
	fmt.Printf("Stopped instance %s\n", instanceName)
}

func startInstance(svc *lightsail.Client, instanceName string) {
	input := &lightsail.StartInstanceInput{
		InstanceName: aws.String(instanceName),
	}

	_, err := svc.StartInstance(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to start instance, %v", err)
	}
	fmt.Printf("Started instance %s\n", instanceName)
}
func waitForInstance(svc *lightsail.Client, instanceName, state string) {
	for {
		input := &lightsail.GetInstanceInput{
			InstanceName: aws.String(instanceName),
		}
		result, err := svc.GetInstance(context.TODO(), input)
		if err != nil {
			log.Fatalf("failed to get instance state, %v", err)
		}

		if *result.Instance.State.Name == state {
			//fmt.Printf("Instance %s is running\n", instanceName)
			return
		}

		//fmt.Printf("Waiting for instance %s to be running\n", instanceName)
		time.Sleep(3 * time.Second)
	}
}

func getInstancePublicIP(svc *lightsail.Client, instanceName string) string {
	input := &lightsail.GetInstanceInput{
		InstanceName: aws.String(instanceName),
	}

	result, err := svc.GetInstance(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to get instance details, %v", err)
	}

	return aws.ToString(result.Instance.PublicIpAddress)
}
