cat > README.md << 'ENDOFFILE'
# CloudForge ☁️

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/AWS-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white" alt="AWS">
  <img src="https://img.shields.io/badge/Azure-0089D6?style=for-the-badge&logo=microsoftazure&logoColor=white" alt="Azure">
  <img src="https://img.shields.io/badge/GCP-4285F4?style=for-the-badge&logo=googlecloud&logoColor=white" alt="GCP">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License">
</p>

<p align="center">
  <b>Write once, deploy anywhere.</b><br>
  A unified Go toolkit for provisioning cloud infrastructure across AWS, Azure, and GCP with identical workflows.
</p>

---

## 🚀 Quick Start

Deploy a VM + Storage stack in under 5 minutes:

```bash
# Clone the repository
git clone https://github.com/yourusername/cloudforge.git
cd cloudforge

# Set up your environment
cp .env.example .env
# Edit .env with your cloud credentials

# Deploy to AWS
cd cmd/compute
go run main.go

# Or deploy storage
cd ../storage
go run main.go
