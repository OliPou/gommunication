# Gommunication Service API: Unified Email and SMS Management Platform

The Gommunication Service API is a robust RESTful service that provides a unified interface for managing both email and SMS communications. It enables applications to send, track, and manage email and text messages through a secure, Kong and his apiKey.

This service offers comprehensive communication management capabilities including email composition with HTML and plain text support, SMS delivery with status tracking, and webhook integration for delivery receipts. The platform is built with a focus on reliability and scalability, using PostgreSQL for persistent storage and providing detailed status tracking for all communications.

## Repository Structure
```
.
├── auth/                      # Authentication related components
├── docs/                      # API documentation and Swagger specifications
│   ├── swagger.json          # OpenAPI/Swagger JSON specification
│   └── swagger.yaml          # OpenAPI/Swagger YAML specification
├── email/                    # Email service implementation
│   ├── handlers.go           # HTTP handlers for email endpoints
│   ├── services.go           # Email business logic implementation
│   └── models.go             # Email data structures
├── internal/                 # Internal application code
│   ├── common/               # Shared utilities and validators
│   └── database/            # Database access layer and generated SQL code
├── sql/                     # SQL definitions and migrations
│   ├── queries/             # SQL query definitions
│   └── schema/              # Database schema migrations
└── textmessage/            # SMS service implementation
    ├── handlers.go          # HTTP handlers for SMS endpoints
    ├── services.go          # SMS business logic implementation
    └── models.go            # SMS data structures
```

## Usage Instructions
### Prerequisites
- PostgreSQL database server
- Go 1.x or higher
- API key for authentication

### Installation
```bash
# Clone the repository
git clone <repository-url>

# Install dependencies
go mod download

# Run database migrations
goose postgres postgres://psqladmin:yourPasword@localhost:5432/gommunication up

# Build the application
go build -o gommunication
```

### Quick Start
1. Start the service:
```bash
./gommunication
```

2. Send an email:
```bash
curl -X POST http://localhost:8081/email/send \
  -H "x-application-id: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "emailSubject": "Test Email",
    "emailText": "Hello World",
    "html": "<h1>Hello World</h1>",
    "recipientEmail": "recipient@example.com",
    "recipientName": "John Doe",
    "senderEmail": "sender@example.com",
    "senderName": "Jane Smith",
    "userName": "service-user"
  }'
```

3. Send an SMS:
```bash
curl -X POST http://localhost:8081/text-message/send \
  -H "x-application-id: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "+1234567890",
    "sender": "ServiceName",
    "text": "Hello from the messaging service",
    "userName": "service-user"
  }'
```

### More Detailed Examples
#### Retrieving Email History
```bash
curl -X GET http://localhost:8081/email \
  -H "x-application-id: your-api-key"
```

#### Retrieving SMS History
```bash
curl -X GET http://localhost:8081/text-message \
  -H "x-application-id: your-api-key"
```

### Troubleshooting
#### Common Issues
1. Authentication Failures
   - Error: "401 Unauthorized"
   - Solution: Verify your API key is correctly set in the x-application-id header

2. Database Connection Issues
   - Error: "failed to connect to database"
   - Check database credentials and connection string
   - Verify PostgreSQL service is running

3. Message Delivery Failures
   - Check the message status using GET endpoints
   - Verify recipient information is correctly formatted
   - Review server logs for detailed error messages

## Data Flow
The service processes communication requests through a multi-stage pipeline, handling authentication, validation, and delivery through external providers while maintaining persistent storage of all communications.

```ascii
Client Request → Authentication → Validation → Business Logic → External Provider
     ↑                                            ↓
     └────────────── Database Storage ────────────┘
```

Key Component Interactions:
1. Authentication middleware validates API keys for all requests
2. Request validation ensures all required fields are present and properly formatted
3. Business logic layer processes the request and prepares it for the external provider
4. External provider integration handles actual message delivery
5. Database layer maintains persistent record of all communications
6. Webhook endpoints receive and process delivery status updates
7. Response handlers format and return appropriate status information to clients


## DI Schema
```

                            +-----------------+
                            |     main.go     |
                            +-----------------+
                                     |
                            calls run() → init app
                                     |
                                     v
                      +-----------------------------+
                      |  di.BuildDependencies()      |
                      +-----------------------------+
                                     |
         +---------------------------+---------------------------+
         |                           |                           |
         v                           v                           v
+------------------+   +--------------------------+   +--------------------------+
| config.DB (SQL)  |   | SendGridEmailSender      |   | Env vars (VONAGE_*)      |
+------------------+   +--------------------------+   +--------------------------+
         |                          |                           |
         v                          v                           v
+-----------------------------+  +--------------------------+   |
| database.New(config.DB)     |  | sendgrid.NewSendClient() |   |
| → dbQueries                 |  +--------------------------+   |
+-----------------------------+                                 |
         |                                                      |
         v                                                      v
+--------------------------------------+     +------------------------------------------+
| email.ApiConfig                      |     | textmessage.ApiConfig                    |
|  - DB: dbQueries                     |     |  - DB: dbQueries                         |
|  - EmailSender: SendGridEmailSender  |     |  - ApiKey, Secret, URLs from env vars    |
+--------------------------------------+     +------------------------------------------+
         |                                                      |
         +------------------------+-----------------------------+
                                  |
                                  v
                    +------------------------------+
                    |   AppDependencies struct     |
                    |   - APICfg                   |
                    |   - APICfgTextMessage        |
                    |   - ...                      |
                    +------------------------------+
                                  |
                                  v
                    +------------------------------+
                    | routers.SetupRouter(deps)    |
                    |  (handler injection)         |
                    +------------------------------+
                                  |
                                  v
                        +------------------+
                        |   Gin Router     |
                        +------------------+
```