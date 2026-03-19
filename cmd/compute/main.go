package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/alisina-1231/aws-vm/pkg/helpers"
	"github.com/alisina-1231/aws-vm/pkg/mgmt"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	region := helpers.MustGetenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	sshPubKeyPath := helpers.MustGetenv("SSH_PUBLIC_KEY_PATH")

	cfg := helpers.BuildAWSConfig(context.Background(), region)
	factory := mgmt.NewVirtualMachineFactory(cfg, sshPubKeyPath)

	fmt.Println("Starting to build AWS resources...")
	stack := factory.CreateVirtualMachineStack(context.Background(), region)

	var (
		ipAddress       = *stack.PublicIP.PublicIp
		sshIdentityPath = strings.TrimRight(sshPubKeyPath, ".pub")
	)
	fmt.Printf("Connect with: `ssh -i %s ubuntu@%s`\n\n", sshIdentityPath, ipAddress)
	fmt.Println("Press enter to delete the infrastructure.")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	factory.DestroyVirtualMachineStack(context.Background(), stack)
}
