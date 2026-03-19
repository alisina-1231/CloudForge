
# CloudForge ☁️

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/AWS-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white" alt="AWS">
  <img src="https://img.shields.io/badge/EC2-FF9900?style=for-the-badge&logo=amazon-ec2&logoColor=white" alt="EC2">
  <img src="https://img.shields.io/badge/S3-569A31?style=for-the-badge&logo=amazon-s3&logoColor=white" alt="S3">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>Infrastructure as Code, Simplified.</b><br>
  A production-ready Go toolkit for automating AWS infrastructure deployment with intelligent resource management and automated teardown.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#aws-resources">AWS Resources</a> •
  <a href="#api-reference">API Reference</a> •
  <a href="#examples">Examples</a>
</p>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [AWS Resources](#aws-resources)
- [API Reference](#api-reference)
- [Error Handling](#error-handling)
- [Cleanup Strategy](#cleanup-strategy)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

CloudForge is a robust infrastructure automation tool written in Go that demonstrates enterprise-grade patterns for AWS resource management. Built with the AWS SDK for Go v2, it provides:

- **Declarative Infrastructure**: Define your stack once, deploy consistently
- **Intelligent Resource Tracking**: Automatic dependency mapping and cleanup
- **Production-Ready Patterns**: Context-aware operations, structured logging, and graceful error handling
- **Cost Control**: One-keypress complete infrastructure destruction

Currently supports **AWS** (EC2, S3, VPC, IAM) with Azure and GCP support planned.

---

## ✨ Features

### Core Capabilities

| Feature | Implementation | Status |
|---------|---------------|--------|
| **VPC Automation** | Custom VPC with subnets, IGW, route tables | ✅ Complete |
| **EC2 Deployment** | Ubuntu 18.04 LTS with cloud-init | ✅ Complete |
| **Security Groups** | SSH (port 22) ingress rules | ✅ Complete |
| **Elastic IPs** | Static public IP allocation | ✅ Complete |
| **S3 Buckets** | Object storage with presigned URLs | ✅ Complete |
| **Key Management** | SSH key pair import and injection | ✅ Complete |
| **Cloud-Init** | Automated software provisioning | ✅ Complete |
| **Resource Tagging** | Consistent naming with haikunator | ✅ Complete |

### Advanced Features

- **Dependency-Aware Cleanup**: Resources deleted in correct order (EIP → Instance → Security Group → Subnet → Route Table → IGW → VPC)
- **Retry Logic**: Exponential backoff for API rate limits and eventual consistency
- **Context Propagation**: Full context.Context support for cancellation and timeouts
- **Structured Errors**: AWS Smithy error handling with detailed error codes
- **Waiters**: Native AWS waiters for resource state transitions

---

## 🏗️ Architecture

### Design Patterns

```
┌─────────────────────────────────────────────────────────────────┐
│                         CloudForge                                │
│                    (Factory Pattern)                              │
├─────────────────────────────────────────────────────────────────┤
│  VirtualMachineFactory        │  StorageFactory                  │
│  ├── createVPC()              │  ├── createBucket()              │
│  ├── createSubnet()           │  ├── uploadObjects()             │
│  ├── createInternetGateway()  │  └── generatePresignedURLs()   │
│  ├── createRouteTable()       │                                  │
│  ├── createSecurityGroup()    │                                  │
│  ├── importKeyPair()          │                                  │
│  ├── createInstance()         │                                  │
│  └── createElasticIP()        │                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  AWS SDK Go v2  │
                    │  ├── EC2        │
                    │  ├── S3         │
                    │  └── IAM        │
                    └─────────────────┘
```

### Resource Lifecycle

```
Create Phase:                    Destroy Phase:
─────────────                    ─────────────
1. VPC                           1. Release Elastic IP
2. Subnet                        2. Terminate Instance (wait)
3. Internet Gateway              3. Delete Key Pair
4. Route Table                   4. Delete Security Group
5. Security Group                5. Delete Subnet
6. Key Pair                      6. Delete Route Table
7. EC2 Instance                  7. Detach & Delete IGW
8. Elastic IP                    8. Delete VPC
```

---

## 📁 Project Structure

```
cloudforge/
├── 📂 blobs/                      # Sample files for S3 upload
│   ├── image1.png
│   ├── document.pdf
│   └── ...
│
├── 📂 cloud-init/
│   └── init.yml                   # Cloud-init configuration
│
├── 📂 cmd/                        # Application entry points
│   ├── 📂 compute/
│   │   └── main.go               # VM deployment CLI
│   └── 📂 storage/
│       └── main.go               # S3 storage CLI
│
├── 📂 pkg/                        # Library code
│   ├── 📂 helpers/
│   │   └── helpers.go            # Shared utilities
│   │       ├── BuildAWSConfig()  # AWS configuration
│   │       ├── MustGetenv()      # Environment validation
│   │       ├── HandleErr()       # Error handling
│   │       └── WaitForResource() # Polling utilities
│   │
│   └── 📂 mgmt/                  # AWS resource management
│       ├── compute.go            # EC2/VPC automation
│       │   ├── VirtualMachineFactory
│       │   ├── VirtualMachineStack
│       │   ├── CreateVirtualMachineStack()
│       │   └── DestroyVirtualMachineStack()
│       │
│       └── storage.go            # S3 automation
│           ├── StorageFactory
│           ├── StorageStack
│           ├── CreateStorageStack()
│           └── DestroyStorageStack()
│
├── 📄 .env.example               # Environment template
├── 📄 go.mod                     # Go module definition
├── 📄 go.sum                     # Dependency checksums
└── 📄 README.md                  # This file
```

---

## 🛠️ Prerequisites

### Required

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Runtime and build |
| AWS CLI | 2.x+ | Credential management |
| Git | Any | Version control |

### AWS Setup

1. **Create AWS Account**: https://aws.amazon.com/free/

2. **Configure Credentials**:
   ```bash
   # Option 1: AWS CLI (Recommended)
   aws configure
   
   # Option 2: Environment variables
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   
   # Option 3: IAM Role (EC2/ECS/Lambda)
   # No configuration needed, uses instance metadata
   ```

3. **Required IAM Permissions**:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "ec2:RunInstances",
           "ec2:TerminateInstances",
           "ec2:DescribeInstances",
           "ec2:CreateVpc",
           "ec2:DeleteVpc",
           "ec2:CreateSubnet",
           "ec2:DeleteSubnet",
           "ec2:CreateInternetGateway",
           "ec2:DeleteInternetGateway",
           "ec2:AttachInternetGateway",
           "ec2:DetachInternetGateway",
           "ec2:CreateRouteTable",
           "ec2:DeleteRouteTable",
           "ec2:CreateRoute",
           "ec2:AssociateRouteTable",
           "ec2:CreateSecurityGroup",
           "ec2:DeleteSecurityGroup",
           "ec2:AuthorizeSecurityGroupIngress",
           "ec2:ImportKeyPair",
           "ec2:DeleteKeyPair",
           "ec2:AllocateAddress",
           "ec2:ReleaseAddress",
           "ec2:AssociateAddress",
           "ec2:DescribeAddresses",
           "s3:CreateBucket",
           "s3:DeleteBucket",
           "s3:PutObject",
           "s3:GetObject",
           "s3:DeleteObject",
           "s3:ListBucket"
         ],
         "Resource": "*"
       }
     ]
   }
   ```

4. **Generate SSH Key** (if not exists):
   ```bash
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/aws_vm_key
   # This creates:
   # - ~/.ssh/aws_vm_key (private key)
   # - ~/.ssh/aws_vm_key.pub (public key - used by CloudForge)
   ```

---

## 📦 Installation

### From Source

```bash
# Clone repository
git clone https://github.com/yourusername/cloudforge.git
cd cloudforge

# Initialize Go module
go mod init github.com/yourusername/cloudforge 2>/dev/null || true

# Download dependencies
go mod tidy

# Verify installation
go build ./...
```

### Dependencies

```go
// go.mod
require (
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/aws/aws-sdk-go-v2/config v1.26.1
    github.com/aws/aws-sdk-go-v2/service/ec2 v1.141.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.47.5
    github.com/aws/smithy-go v1.19.0
    github.com/joho/godotenv v1.5.1
    github.com/mitchellh/go-homedir v1.1.0
    github.com/yelinaung/go-haikunator v0.0.0-20150320034127-1249c5f72177
)
```

---

## 🔐 Configuration

### Environment Variables

Create `.env` file in project root:

```bash
# AWS Configuration
AWS_REGION=us-east-1                    # AWS region for deployment
AWS_ACCESS_KEY_ID=AKIA...               # IAM user access key
AWS_SECRET_ACCESS_KEY=secret...       # IAM user secret key

# SSH Configuration
SSH_PUBLIC_KEY_PATH=~/.ssh/id_rsa.pub # Path to public key for EC2

# Optional: Custom AMI (default: Ubuntu 18.04 LTS)
AWS_AMI_ID=ami-0c55b159cbfafe1f0      # Override default AMI

# Optional: Instance type (default: t3.medium)
AWS_INSTANCE_TYPE=t3.medium           # EC2 instance size
```

### Region-Specific AMIs

| Region | Ubuntu 18.04 LTS AMI | Description |
|--------|---------------------|-------------|
| us-east-1 | ami-0c55b159cbfafe1f0 | N. Virginia |
| us-east-2 | ami-0e83be366243f524a | Ohio |
| us-west-1 | ami-08d9a394ac1c2994c | N. California |
| us-west-2 | ami-0e34e7e9c8b2a1c7d | Oregon |
| eu-west-1 | ami-0a8e758f5e9d09d6f | Ireland |

> **Note**: Update AMI IDs periodically as AWS retires old images. Find latest at: https://cloud-images.ubuntu.com/locator/ec2/

---

## 🚀 Usage

### Compute Stack (EC2)

Deploy a complete VPC with EC2 instance:

```bash
cd cmd/compute
go run main.go
```

**Expected Output:**
```text
Starting to build AWS resources...
Building an AWS VPC named "black-disk"...
Building an AWS Subnet named "black-disk-subnet"...
Building an AWS Internet Gateway named "black-disk-igw"...
Building an AWS Route Table named "black-disk-rt"...
Building an AWS Security Group named "black-disk-sg"...
Importing SSH Key Pair named "black-disk-key"...
Building an AWS EC2 Instance named "black-disk-vm"...
Allocating Elastic IP for the instance...

✅ Infrastructure Ready!

Connect with: `ssh -i /home/user/.ssh/aws_vm_key ubuntu@52.70.180.140`

Instance Details:
  - Instance ID: i-0abcd1234efgh5678i
  - Public IP: 52.70.180.140
  - Private IP: 10.0.0.123
  - VPC: vpc-0ce17946e383c1088
  - Subnet: subnet-0a1b2c3d4e5f6789a

Cloud-init Status:
  - Packages: nginx, golang
  - Custom script: Applied

Press enter to delete the infrastructure.
```

**Connect to Instance:**
```bash
ssh -i ~/.ssh/aws_vm_key ubuntu@52.70.180.140

# Verify cloud-init
curl http://localhost
# Output: "hello world"

# Check installed packages
nginx -v
go version
```

### Storage Stack (S3)

Deploy S3 bucket with file upload:

```bash
cd cmd/storage
go run main.go
```

**Expected Output:**
```text
Starting to build AWS resources...
Building an AWS S3 Bucket named "winter-sun-bucket"...
Creating a new container "jd-imgs" in the S3 Bucket...
Reading all files ./blobs...
Uploading file "image1.png" to s3://winter-sun-bucket/jd-imgs/image1.png...
Uploading file "document.pdf" to s3://winter-sun-bucket/jd-imgs/document.pdf...

✅ Storage Ready!

Generating readonly links to blobs that expire in 2 hours...
https://winter-sun-bucket.s3.amazonaws.com/jd-imgs/image1.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...
https://winter-sun-bucket.s3.amazonaws.com/jd-imgs/document.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...

Bucket Details:
  - Name: winter-sun-bucket
  - ARN: arn:aws:s3:::winter-sun-bucket
  - Region: us-east-1
  - Objects: 2
  - Total Size: 2.4 MB

Press enter to delete the infrastructure.
```

---

## 🏗️ AWS Resources

### Compute Resources

#### VirtualMachineFactory

Main factory for EC2 infrastructure orchestration.

```go
type VirtualMachineFactory struct {
    cfg           aws.Config
    ec2Client     *ec2.Client
    sshPubKeyPath string
}

func NewVirtualMachineFactory(cfg aws.Config, sshPubKeyPath string) *VirtualMachineFactory
```

**Methods:**

| Method | Description | AWS API Calls |
|--------|-------------|---------------|
| `CreateVirtualMachineStack(ctx, location)` | Creates full EC2 stack | CreateVpc, CreateSubnet, CreateIGW, CreateRouteTable, CreateSecurityGroup, ImportKeyPair, RunInstances, AllocateAddress |
| `DestroyVirtualMachineStack(ctx, stack)` | Destroys stack in dependency order | ReleaseAddress, TerminateInstances, DeleteKeyPair, DeleteSecurityGroup, DeleteSubnet, DeleteRouteTable, DetachIGW, DeleteIGW, DeleteVpc |

#### VirtualMachineStack

Tracks all created resources for cleanup.

```go
type VirtualMachineStack struct {
    Location         string              // AWS region
    name             string              // Resource name prefix (haikunate)
    sshKeyPath       string              // Expanded SSH key path
    VPC              types.Vpc           // AWS VPC
    Subnet           types.Subnet        // AWS Subnet
    SecurityGroup    types.SecurityGroup // Security group with SSH rule
    Instance         types.Instance      // EC2 instance
    KeyPair          types.KeyPairInfo   // Imported SSH key
    PublicIP         types.Address       // Elastic IP
    InternetGateway  *types.InternetGateway
    RouteTable       *types.RouteTable
}
```

### Storage Resources

#### StorageFactory

Factory for S3 bucket management.

```go
type StorageFactory struct {
    cfg    aws.Config
    region string
}

func NewStorageFactory(cfg aws.Config, region string) *StorageFactory
```

**Methods:**

| Method | Description |
|--------|-------------|
| `CreateStorageStack(ctx, location)` | Creates S3 bucket |
| `DestroyStorageStack(ctx, stack)` | Deletes bucket and all objects |

#### StorageStack

```go
type StorageStack struct {
    Location   string
    name       string
    BucketName string
    Region     string
    cfg        aws.Config
    s3Client   *s3.Client
}
```

---

## 📚 API Reference

### Helper Functions

```go
// pkg/helpers/helpers.go

// BuildAWSConfig loads AWS configuration with region
func BuildAWSConfig(ctx context.Context, region string) aws.Config

// MustGetenv retrieves environment variable or exits
func MustGetenv(key string) string

// HandleErr panics on error with AWS error details
func HandleErr(err error)

// HandleErrWithResult returns result or panics on error
func HandleErrWithResult[T any](result T, err error) T

// WaitForResource polls until condition met
func WaitForResource(ctx context.Context, checkFunc func() (bool, error), timeout time.Duration)
```

### Error Handling

CloudForge uses structured error handling with AWS Smithy:

```go
import "github.com/aws/smithy-go"

// Check for specific AWS errors
var apiErr smithy.Error
if errors.As(err, &apiErr) {
    switch apiErr.ErrorCode() {
    case "DependencyViolation":
        // Resource has dependencies, retry after delay
    case "InvalidParameterValue":
        // Invalid input, check parameters
    case "UnauthorizedOperation":
        // IAM permissions issue
    }
}
```

---

## 🧹 Cleanup Strategy

### Dependency Order

Resources must be deleted in specific order to avoid `DependencyViolation` errors:

```
1. Elastic IP (ReleaseAddress)
   └── Attached to EC2 instance
   
2. EC2 Instance (TerminateInstances)
   └── Uses: Security Group, Subnet, Key Pair
   
3. Key Pair (DeleteKeyPair)
   └── Referenced by instance (now terminated)
   
4. Security Group (DeleteSecurityGroup)
   └── Referenced by instance (now terminated)
   
5. Subnet (DeleteSubnet)
   └── Uses: Route Table association
   
6. Route Table (DeleteRouteTable)
   └── Uses: IGW route, Subnet association
   
7. Internet Gateway (DetachInternetGateway → DeleteInternetGateway)
   └── Attached to VPC
   
8. VPC (DeleteVpc)
   └── All dependencies removed
```

### Retry Logic

```go
// Exponential backoff for VPC deletion
maxRetries := 10
for i := 0; i < maxRetries; i++ {
    _, err := ec2Client.DeleteVpc(ctx, input)
    if err == nil {
        break // Success
    }
    
    var apiErr smithy.Error
    if errors.As(err, &apiErr) && apiErr.ErrorCode() == "DependencyViolation" {
        waitTime := time.Duration(i+1) * 5 * time.Second
        time.Sleep(waitTime) // Backoff: 5s, 10s, 15s...
        continue
    }
    break // Non-retryable error
}
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test -v ./pkg/mgmt/...

# Run with race detection
go test -race ./...
```

### Integration Tests

```bash
# Set test environment
export AWS_REGION=us-east-1
export SSH_PUBLIC_KEY_PATH=~/.ssh/id_rsa.pub

# Run integration test (creates real resources)
go test -tags=integration ./...

# Cleanup after tests (always run)
go test -tags=integration -run=Cleanup ./...
```

### Manual Testing Checklist

- [ ] VPC created with correct CIDR (10.0.0.0/16)
- [ ] Subnet created with correct CIDR (10.0.0.0/24)
- [ ] Internet Gateway attached to VPC
- [ ] Route table has route to IGW (0.0.0.0/0)
- [ ] Security Group allows SSH (port 22)
- [ ] Key pair imported successfully
- [ ] EC2 instance reaches `running` state
- [ ] Elastic IP associated with instance
- [ ] SSH connection successful
- [ ] Cloud-init completed (nginx installed)
- [ ] S3 bucket created in correct region
- [ ] Files uploaded to `jd-imgs/` prefix
- [ ] Presigned URLs accessible (2-hour expiry)
- [ ] Cleanup removes all resources
- [ ] No `DependencyViolation` errors during cleanup

---

## 🔧 Troubleshooting

### Common Issues

#### 1. `DependencyViolation` on VPC Delete

**Symptom:**
```
api error DependencyViolation: The vpc 'vpc-xxx' has dependencies and cannot be deleted.
```

**Cause:** Resources still attached to VPC (IGW, subnets, etc.)

**Solution:** 
- Check AWS Console for remaining resources
- Ensure cleanup runs in correct order
- Increase retry attempts in `DestroyVirtualMachineStack`

#### 2. `InvalidAMIID.NotFound`

**Symptom:**
```
Error: InvalidAMIID.NotFound: The image id '[ami-xxx]' does not exist
```

**Cause:** AMI ID is region-specific or retired

**Solution:**
```bash
# Find current Ubuntu 18.04 AMI for your region
aws ec2 describe-images \
  --owners 099720109477 \
  --filters "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*" \
  --query 'Images[0].ImageId' \
  --region your-region
```

#### 3. `UnauthorizedOperation`

**Symptom:**
```
Error: UnauthorizedOperation: You are not authorized to perform this operation
```

**Cause:** IAM user/role lacks required permissions

**Solution:** Attach policy from [Prerequisites](#prerequisites) section

#### 4. `Resource.AlreadyExists`

**Symptom:**
```
Error: InvalidKeyPair.Duplicate: The keypair 'xxx' already exists
```

**Cause:** Previous run didn't clean up key pair

**Solution:**
```bash
# List and delete existing key pairs
aws ec2 describe-key-pairs --query 'KeyPairs[*].KeyName'
aws ec2 delete-key-pair --key-name your-key-name
```

#### 5. SSH Connection Refused

**Symptom:**
```
ssh: connect to host x.x.x.x port 22: Connection refused
```

**Causes & Solutions:**
- **Security Group**: Verify port 22 is open to your IP
- **Instance not ready**: Wait for status checks to pass
- **Wrong key**: Ensure using private key (not .pub)
- **Wrong user**: Use `ubuntu` for Ubuntu AMIs

```bash
# Check instance status
aws ec2 describe-instance-status --instance-id i-xxx

# Check security group rules
aws ec2 describe-security-groups --group-ids sg-xxx
```

### Debug Mode

Enable verbose logging:

```go
import "github.com/aws/smithy-go/logging"

cfg, err := config.LoadDefaultConfig(ctx,
    config.WithRegion(region),
    config.WithClientLogMode(aws.LogRetries|aws.LogRequest|aws.LogResponse),
    config.WithLogger(logging.NewStandardLogger(os.Stdout)),
)
```

---

## 🗺️ Roadmap

### Phase 1: AWS Foundation ✅
- [x] EC2 with VPC networking
- [x] S3 with presigned URLs
- [x] Automated cleanup with dependency ordering
- [x] Cloud-init support

### Phase 2: AWS Advanced (In Progress)
- [ ] Auto Scaling Groups
- [ ] Application Load Balancer
- [ ] RDS database instances
- [ ] EKS (Kubernetes) clusters
- [ ] CloudWatch monitoring
- [ ] IAM role management

### Phase 3: Multi-Cloud
- [ ] Azure Virtual Machines
- [ ] Azure Blob Storage
- [ ] GCP Compute Engine
- [ ] GCP Cloud Storage
- [ ] Cross-cloud migration tools

### Phase 4: Platform
- [ ] REST API for infrastructure management
- [ ] Web UI for resource visualization
- [ ] Terraform provider
- [ ] GitHub Actions integration
- [ ] Cost estimation and optimization

---

## 🤝 Contributing

We welcome contributions! Please follow these guidelines:

### Development Setup

```bash
# Fork and clone
git clone https://github.com/yourusername/cloudforge.git
cd cloudforge

# Create branch
git checkout -b feature/your-feature

# Make changes
# ...

# Test
go test ./...
go vet ./...
gofmt -w .

# Commit
git commit -m "feat: add new feature"

# Push
git push origin feature/your-feature
```

### Code Standards

- **Go Version**: 1.21+
- **Formatting**: `gofmt -w .`
- **Linting**: `golangci-lint run`
- **Testing**: >80% coverage for new code
- **Documentation**: Update README for API changes

### Commit Message Format

```
type(scope): subject

body

footer
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(aws): add support for multiple security group rules

- Allow configuring multiple ingress/egress rules
- Maintain backward compatibility with single rule

Closes #123
```

---

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

```
MIT License

Copyright (c) 2024 [Your Name]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## 🙏 Acknowledgments

- **AWS SDK for Go v2**: https://github.com/aws/aws-sdk-go-v2
- **go-haikunator**: Heroku-style memorable names
- **Packt Publishing**: Original Go for DevOps patterns
- **Ubuntu Cloud Images**: https://cloud-images.ubuntu.com/

---

## 📞 Support

| Channel | Link |
|---------|------|
| 🐛 Issues | https://github.com/yourusername/cloudforge/issues |
| 💬 Discussions | https://github.com/yourusername/cloudforge/discussions |
| 📧 Email | your.email@example.com |
| 🐦 Twitter | @yourhandle |

---

<p align="center">
  <sub>Built with ❤️ using <b>Go</b> and <b>AWS</b></sub>
</p> 
