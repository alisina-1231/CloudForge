package mgmt
import (
    "context"
    "encoding/base64"
    "errors"
    "fmt"
    "io/ioutil"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/aws/smithy-go"   // ✅ keep ONLY ONE
    "github.com/mitchellh/go-homedir"
    "github.com/yelinaung/go-haikunator"

    . "github.com/alisina-1231/aws-vm/pkg/helpers"
)

var (
	haiku = haikunator.New(time.Now().UnixNano())
)

type VirtualMachineFactory struct {
	cfg           aws.Config
	ec2Client     *ec2.Client
	sshPubKeyPath string
}

// VirtualMachineStack now tracks the Internet Gateway and Route Table explicitly
type VirtualMachineStack struct {
	Location         string
	name             string
	sshKeyPath       string
	VPC              types.Vpc
	Subnet           types.Subnet
	SecurityGroup    types.SecurityGroup
	Instance         types.Instance
	KeyPair          types.KeyPairInfo
	PublicIP         types.Address
	InternetGateway  *types.InternetGateway
	RouteTable       *types.RouteTable // Track Route Table for deletion
}

// NewVirtualMachineFactory instantiates an AWS EC2 factory
func NewVirtualMachineFactory(cfg aws.Config, sshPubKeyPath string) *VirtualMachineFactory {
	return &VirtualMachineFactory{
		cfg:           cfg,
		ec2Client:     ec2.NewFromConfig(cfg),
		sshPubKeyPath: sshPubKeyPath,
	}
}

// CreateVirtualMachineStack creates an EC2 instance with VPC, subnet, security group
func (vmf *VirtualMachineFactory) CreateVirtualMachineStack(ctx context.Context, location string) *VirtualMachineStack {
	stack := &VirtualMachineStack{
		Location:   location,
		name:       haiku.Haikunate(),
		sshKeyPath: HandleErrWithResult(homedir.Expand(vmf.sshPubKeyPath)),
	}

	// Create VPC
	stack.VPC = vmf.createVPC(ctx, stack.name)
	// Create Subnet
	stack.Subnet = vmf.createSubnet(ctx, stack)
	// Create Internet Gateway and attach to VPC
	igw := vmf.createInternetGateway(ctx, stack.name)
	stack.InternetGateway = &igw
	vmf.attachInternetGateway(ctx, stack)
	// Create Route Table and route to IGW (now tracked in stack)
	stack.RouteTable = vmf.createRouteTable(ctx, stack)
	// Create Security Group
	stack.SecurityGroup = vmf.createSecurityGroup(ctx, stack)
	// Import Key Pair
	stack.KeyPair = vmf.importKeyPair(ctx, stack)
	// Create Instance
	stack.Instance = vmf.createInstance(ctx, stack)
	// Allocate and associate Elastic IP
	stack.PublicIP = vmf.createElasticIP(ctx, stack)
	
	return stack
}

// DestroyVirtualMachineStack terminates instance and deletes resources in correct order
func (vmf *VirtualMachineFactory) DestroyVirtualMachineStack(ctx context.Context, vmStack *VirtualMachineStack) {
	fmt.Printf("Deleting AWS resources for %s...\n", vmStack.name)
	
	// 1. Release Elastic IP first (it's attached to instance)
	if vmStack.PublicIP.AllocationId != nil {
		fmt.Printf("Releasing Elastic IP %s...\n", *vmStack.PublicIP.AllocationId)
		_, err := vmf.ec2Client.ReleaseAddress(ctx, &ec2.ReleaseAddressInput{
			AllocationId: vmStack.PublicIP.AllocationId,
		})
		if err != nil {
			fmt.Printf("Warning: failed to release Elastic IP: %v\n", err)
		}
	}

	// 2. Terminate Instance and wait for it to be fully terminated
	if vmStack.Instance.InstanceId != nil {
		fmt.Printf("Terminating instance %s...\n", *vmStack.Instance.InstanceId)
		_, err := vmf.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			InstanceIds: []string{*vmStack.Instance.InstanceId},
		})
		HandleErr(err)
		
		// Wait for termination - this is crucial before deleting network resources
		fmt.Printf("Waiting for instance to terminate...\n")
		waiter := ec2.NewInstanceTerminatedWaiter(vmf.ec2Client)
		err = waiter.Wait(ctx, &ec2.DescribeInstancesInput{
			InstanceIds: []string{*vmStack.Instance.InstanceId},
		}, 5*time.Minute)
		if err != nil {
			fmt.Printf("Warning: error waiting for termination: %v\n", err)
		}
	}

	// 3. Delete Key Pair
	if vmStack.KeyPair.KeyName != nil {
		fmt.Printf("Deleting key pair %s...\n", *vmStack.KeyPair.KeyName)
		_, err := vmf.ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{
			KeyName: vmStack.KeyPair.KeyName,
		})
		if err != nil {
			fmt.Printf("Warning: failed to delete key pair: %v\n", err)
		}
	}

	// 4. Delete Security Group (only after instance is gone)
	if vmStack.SecurityGroup.GroupId != nil {
		fmt.Printf("Deleting security group %s...\n", *vmStack.SecurityGroup.GroupId)
		// Retry logic for security group deletion (might have eventual consistency issues)
		for i := 0; i < 5; i++ {
			_, err := vmf.ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
				GroupId: vmStack.SecurityGroup.GroupId,
			})
			if err == nil {
				break
			}
			if i < 4 {
				fmt.Printf("Retry deleting security group (%d): %v\n", i+1, err)
				time.Sleep(5 * time.Second)
			} else {
				fmt.Printf("Warning: failed to delete security group: %v\n", err)
			}
		}
	}

	// 5. Delete Subnet (only after instance and security group are gone)
	if vmStack.Subnet.SubnetId != nil {
		fmt.Printf("Deleting subnet %s...\n", *vmStack.Subnet.SubnetId)
		_, err := vmf.ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: vmStack.Subnet.SubnetId,
		})
		if err != nil {
			fmt.Printf("Warning: failed to delete subnet: %v\n", err)
		}
	}

	// 6. Delete Route Table (must be deleted before IGW and VPC)
	if vmStack.RouteTable != nil && vmStack.RouteTable.RouteTableId != nil {
		fmt.Printf("Deleting route table %s...\n", *vmStack.RouteTable.RouteTableId)
		_, err := vmf.ec2Client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: vmStack.RouteTable.RouteTableId,
		})
		if err != nil {
			fmt.Printf("Warning: failed to delete route table: %v\n", err)
		}
	}

	// 7. Detach and Delete Internet Gateway
	if vmStack.InternetGateway != nil && vmStack.InternetGateway.InternetGatewayId != nil {
		igwId := *vmStack.InternetGateway.InternetGatewayId
		vpcId := *vmStack.VPC.VpcId
		
		fmt.Printf("Detaching Internet Gateway %s from VPC %s...\n", igwId, vpcId)
		_, err := vmf.ec2Client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &igwId,
			VpcId:             &vpcId,
		})
		if err != nil {
			fmt.Printf("Warning: failed to detach IGW: %v\n", err)
		} else {
			// Wait a moment for detachment to complete
			time.Sleep(2 * time.Second)
			
			fmt.Printf("Deleting Internet Gateway %s...\n", igwId)
			_, err = vmf.ec2Client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
				InternetGatewayId: &igwId,
			})
			if err != nil {
				fmt.Printf("Warning: failed to delete IGW: %v\n", err)
			}
		}
	}

	// 8. Delete VPC (must be last, after all dependencies are removed)
	if vmStack.VPC.VpcId != nil {
		vpcId := *vmStack.VPC.VpcId
		fmt.Printf("Deleting VPC %s...\n", vpcId)
		
		// Retry logic for VPC deletion with exponential backoff
		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			_, err := vmf.ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
				VpcId: &vpcId,
			})
			if err == nil {
				fmt.Printf("✅ VPC deleted successfully\n")
				break
			}
			
			// Check if it's a dependency violation using smithy error checking
			
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				if apiErr.ErrorCode() == "DependencyViolation" {
					if i < maxRetries-1 {
						waitTime := time.Duration(i+1) * 5 * time.Second
						fmt.Printf("Retry deleting VPC (%d/%d): dependencies still exist, waiting %v...\n", 
							i+1, maxRetries, waitTime)
						time.Sleep(waitTime)
						continue
					}
				}
			}
			
			fmt.Printf("Error deleting VPC: %v\n", err)
			break
		}
	}
	
	fmt.Println("✅ Cleanup completed.")
}

func (vmf *VirtualMachineFactory) createVPC(ctx context.Context, name string) types.Vpc {
	input := &ec2.CreateVpcInput{
		CidrBlock:         aws.String("10.0.0.0/16"),
		TagSpecifications: vmf.createTags("vpc", name),
	}
	
	res := HandleErrWithResult(vmf.ec2Client.CreateVpc(ctx, input))
	
	// Wait for VPC to be available
	waiter := ec2.NewVpcAvailableWaiter(vmf.ec2Client)
	err := waiter.Wait(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{*res.Vpc.VpcId},
	}, 2*time.Minute)
	HandleErr(err)
	
	return *res.Vpc
}

func (vmf *VirtualMachineFactory) createSubnet(ctx context.Context, vmStack *VirtualMachineStack) types.Subnet {
	name := vmStack.name + "-subnet"
	
	input := &ec2.CreateSubnetInput{
		VpcId:            vmStack.VPC.VpcId,
		CidrBlock:        aws.String("10.0.0.0/24"),
		AvailabilityZone: aws.String(vmStack.Location + "a"),
		TagSpecifications: vmf.createTags("subnet", name),
	}
	
	res := HandleErrWithResult(vmf.ec2Client.CreateSubnet(ctx, input))
	return *res.Subnet
}

func (vmf *VirtualMachineFactory) createInternetGateway(ctx context.Context, name string) types.InternetGateway {
	igwName := name + "-igw"
	
	input := &ec2.CreateInternetGatewayInput{
		TagSpecifications: vmf.createTags("internet-gateway", igwName),
	}
	
	res := HandleErrWithResult(vmf.ec2Client.CreateInternetGateway(ctx, input))
	return *res.InternetGateway
}

func (vmf *VirtualMachineFactory) attachInternetGateway(ctx context.Context, vmStack *VirtualMachineStack) {
	_, err := vmf.ec2Client.AttachInternetGateway(ctx, &ec2.AttachInternetGatewayInput{
		InternetGatewayId: vmStack.InternetGateway.InternetGatewayId,
		VpcId:             vmStack.VPC.VpcId,
	})
	HandleErr(err)
}

// createRouteTable now returns the RouteTable for tracking
func (vmf *VirtualMachineFactory) createRouteTable(ctx context.Context, vmStack *VirtualMachineStack) *types.RouteTable {
	name := vmStack.name + "-rt"
	
	// Create route table
	rtRes := HandleErrWithResult(vmf.ec2Client.CreateRouteTable(ctx, &ec2.CreateRouteTableInput{
		VpcId:             vmStack.VPC.VpcId,
		TagSpecifications: vmf.createTags("route-table", name),
	}))
	
	// Create route to IGW
	_, err := vmf.ec2Client.CreateRoute(ctx, &ec2.CreateRouteInput{
		RouteTableId:         rtRes.RouteTable.RouteTableId,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            vmStack.InternetGateway.InternetGatewayId,
	})
	HandleErr(err)
	
	// Associate with subnet
	_, err = vmf.ec2Client.AssociateRouteTable(ctx, &ec2.AssociateRouteTableInput{
		RouteTableId: rtRes.RouteTable.RouteTableId,
		SubnetId:     vmStack.Subnet.SubnetId,
	})
	HandleErr(err)
	
	return rtRes.RouteTable
}

func (vmf *VirtualMachineFactory) createSecurityGroup(ctx context.Context, vmStack *VirtualMachineStack) types.SecurityGroup {
	name := vmStack.name + "-sg"
	
	input := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(name),
		Description: aws.String("Allow SSH access"),
		VpcId:       vmStack.VPC.VpcId,
		TagSpecifications: vmf.createTags("security-group", name),
	}
	
	res := HandleErrWithResult(vmf.ec2Client.CreateSecurityGroup(ctx, input))
	
	// Add SSH rule
	_, err := vmf.ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: res.GroupId,
		IpPermissions: []types.IpPermission{
			{
				IpProtocol: aws.String("tcp"),
				FromPort:   aws.Int32(22),
				ToPort:     aws.Int32(22),
				IpRanges: []types.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("Allow SSH from anywhere"),
					},
				},
			},
		},
	})
	HandleErr(err)
	
	return types.SecurityGroup{
		GroupId:   res.GroupId,
		GroupName: aws.String(name),
	}
}

func (vmf *VirtualMachineFactory) importKeyPair(ctx context.Context, vmStack *VirtualMachineStack) types.KeyPairInfo {
	keyName := vmStack.name + "-key"
	
	sshKeyData := HandleErrWithResult(ioutil.ReadFile(vmStack.sshKeyPath))
	
	input := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: sshKeyData,
		TagSpecifications: vmf.createTags("key-pair", keyName),
	}
	
	res := HandleErrWithResult(vmf.ec2Client.ImportKeyPair(ctx, input))
	return types.KeyPairInfo{
		KeyName:        res.KeyName,
		KeyFingerprint: res.KeyFingerprint,
	}
}

func (vmf *VirtualMachineFactory) createInstance(ctx context.Context, vmStack *VirtualMachineStack) types.Instance {
	name := vmStack.name + "-vm"
	
	// Read cloud-init
	cloudInitContent := HandleErrWithResult(ioutil.ReadFile("/home/ali/aws-vm/cloud-init/init.yml"))
	b64EncodedInit := base64.StdEncoding.EncodeToString(cloudInitContent)
	
	// Get AMI for Ubuntu 18.04 - you may need to adjust this for your region
	// This is a common Ubuntu 18.04 AMI for us-east-1
	amiId := "ami-0199fa5fada510433"
	
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(amiId),
		//InstanceType: "t3.micro",
		InstanceType: types.InstanceTypeT3Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		KeyName:      vmStack.KeyPair.KeyName,
		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				SubnetId:                 vmStack.Subnet.SubnetId,
				Groups:                   []string{*vmStack.SecurityGroup.GroupId},
				AssociatePublicIpAddress: aws.Bool(false),
				DeviceIndex:              aws.Int32(0),
			},
		},
		UserData: aws.String(b64EncodedInit),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(name)},
				},
			},
		},
	}
	
	res := HandleErrWithResult(vmf.ec2Client.RunInstances(ctx, input))
	
	// Wait for instance to be running
	waiter := ec2.NewInstanceRunningWaiter(vmf.ec2Client)
	err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{*res.Instances[0].InstanceId},
	}, 5*time.Minute)
	HandleErr(err)
	
	return res.Instances[0]
}

func (vmf *VirtualMachineFactory) createElasticIP(ctx context.Context, vmStack *VirtualMachineStack) types.Address {
	// Allocate address
	allocRes := HandleErrWithResult(vmf.ec2Client.AllocateAddress(ctx, &ec2.AllocateAddressInput{
		Domain: types.DomainTypeVpc,
	}))
	
	// Associate with instance
	_, err := vmf.ec2Client.AssociateAddress(ctx, &ec2.AssociateAddressInput{
		InstanceId:   vmStack.Instance.InstanceId,
		AllocationId: allocRes.AllocationId,
	})
	HandleErr(err)
	
	// Get the public IP
	describeRes := HandleErrWithResult(vmf.ec2Client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{
		AllocationIds: []string{*allocRes.AllocationId},
	}))
	
	return describeRes.Addresses[0]
}

func (vmf *VirtualMachineFactory) createTags(resourceType, name string) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: types.ResourceType(resourceType),
			Tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String(name)},
			},
		},
	}
}
