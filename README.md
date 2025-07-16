# Server Status Lambda Function

A Go-based AWS Lambda function that fetches and aggregates server status information for Dungeons & Dragons; Online, an
MMO by Standing Stone games.

> 📝 **Note**: This is a fan-driven project, not affiliated with Standing Stone Games, Daybreak Games, or Wizards of the
> Coast.
> See our [full disclaimer](DISCLAIMER.md) for more information.

## Features

- Fetches server information from a configurable datacenter URL
- Concurrent processing of multiple server status endpoints
- Returns sorted server information based on server order
- Handles XML parsing with charset support
- Implements worker pool pattern for efficient concurrent requests
- AWS Lambda compatible
- CORS enabled

## Prerequisites

- Go 1.24 or later
- AWS CLI configured with appropriate credentials
- AWS SAM CLI (optional, for local testing)

## Environment Variables

| Variable       | Description                                | Required |
|----------------|--------------------------------------------|----------|
| DATACENTER_URL | URL of the primary datacenter XML endpoint | Yes      |

## Building

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <project-directory>
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build for AWS Lambda:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
   zip function.zip bootstrap
   ```

## Deployment

### Manual Deployment

1. Create a new Lambda function in AWS Console:
    - Runtime: Custom runtime on Amazon Linux 2
    - Handler: bootstrap
    - Architecture: x86_64

2. Set the environment variable:
    - Key: DATACENTER_URL
    - Value: Your datacenter XML endpoint URL

3. Upload the function:
   ```bash
   aws lambda update-function-code \
       --function-name YOUR_FUNCTION_NAME \
       --zip-file fileb://function.zip
   ```

### Using AWS SAM

1. Create a `template.yaml`:
   ```yaml
   AWSTemplateFormatVersion: '2010-09-09'
   Transform: AWS::Serverless-2016-10-31
   Resources:
     ServerStatusFunction:
       Type: AWS::Serverless::Function
       Properties:
         CodeUri: .
         Handler: bootstrap
         Runtime: provided.al2
         Environment:
           Variables:
             DATACENTER_URL: your-datacenter-url
         Events:
           ApiEvent:
             Type: Api
             Properties:
               Path: /status
               Method: get
   ```

2. Deploy:
   ```bash
   sam build
   sam deploy --guided
   ```

## API Response Format

```json
{
  "servers": [
    {
      "name": "ServerName",
      "commonName": "Server Common Name",
      "status": true,
      "order": 1
    }
  ],
  "errors": [
    "Error message if any"
  ]
}
```

## Local Testing

To test locally with AWS SAM:

```bash
sam local start-api
```

Then access the endpoint at: http://localhost:3000/status

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details

# Disclaimer

This project is a fan-driven initiative and is not affiliated with, endorsed by, or connected to Standing Stone Games,
Daybreak Games, or Wizards of the Coast in any way.

All trademarks, properties, and copyrights belong to their respective owners. This is a free fan project created under
fair use.

Veteran Software operates independently and is not sponsored by, officially connected to, nor endorsed by any of the
aforementioned companies. Any trademarks, registered trademarks, product names, and company names or logos mentioned are
used for identification purposes only and remain the property of their respective owners.

This project is intended to support the gaming community and operates on a non-commercial basis.
