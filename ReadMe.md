# Ticket-Zetu-API

![Go Version](https://img.shields.io/badge/go-1.24%2B-blue)
![MySQL](https://img.shields.io/badge/MySQL-8.0%2B-orange)
![Redis](https://img.shields.io/badge/Redis-7.0%2B-red)

A secure Go RESTful API for event ticket management with full authentication and authorization.

## Features

- 🎫 Complete ticket lifecycle management
- 🔐 Authentication (Email + OAuth)
- 🛡️ Fine-grained role-based access control
- 📧 Secure email communications
- 📊 Interactive API documentation
- 🏗️ Modular architecture

## Prerequisites

### Development Tools
- Go 1.24+
- MySQL 8.0+
- Redis 7.0+
- Git
- Make

### Required Services
- Cloud storage account
- Email service provider
- OAuth provider credentials

## Installation

### 1. Clone the Repository
```bash
git clone https://github.com/glittering-role/ticket-zetu-api.git
cd ticket-zetu-api
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Configuration
Copy the example environment file:
```bash
cp .env.example .env
```

Then configure the environment variables in the `.env` file with your service credentials.

### 4. Database Setup
Initialize the database with:
```bash
make seed-roles
```

## Running the API

### Development Mode
```bash
make run
```
API will be available at: `http://localhost:8080`

### Production Build
```bash
make build
./ticket-zetu-api
```

## API Documentation
Interactive documentation is available at:
`http://localhost:8080/swagger/index.html`

## Security Best Practices

1. Never commit `.env` files
2. Use environment variables for all secrets
3. Rotate credentials regularly
4. Restrict database permissions
5. Enable HTTPS in production

## Contributing
Please see our [Contribution Guidelines](CONTRIBUTING.md) for details.

## License
[MIT License](LICENSE)
