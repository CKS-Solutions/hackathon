# ms-video

Microservice for video upload, processing, and download with S3 and SQS integration.

## Features

- **Video Upload**: Upload videos via multipart form data
- **Asynchronous Processing**: Videos are queued for background processing
- **Progress Tracking**: Monitor video processing progress
- **Video Listing**: List all user's videos with metadata
- **Video Download**: Download processed videos as ZIP files
- **JWT Authentication**: All endpoints require valid JWT token from ms-auth

## Architecture

This service follows hexagonal architecture (ports & adapters):

```
├── cmd/
│   ├── main.go           # Application entry point
│   ├── http/             # HTTP server
│   └── sqs/              # SQS consumer (worker)
├── internal/
│   ├── core/             # Business logic
│   │   ├── entities/     # Domain entities
│   │   ├── ports/        # Interfaces
│   │   └── usecases/     # Use cases
│   ├── adapters/
│   │   ├── driven/       # Output adapters (S3, SQS, DynamoDB, JWT)
│   │   └── driver/       # Input adapters (HTTP controllers, middleware)
│   └── infra/            # Infrastructure (AWS clients)
└── pkg/
    └── utils/            # Utilities
```

## Endpoints

### Health Check
- `GET /health` - No authentication required
- `GET /video/health` - No authentication required

### Video Operations (All require JWT authentication)
- `POST /video/upload` - Upload a new video
- `GET /video/list` - List all user's videos
- `GET /video/download?id={videoId}` - Download processed video

## Authentication

All protected endpoints require JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

The JWT token must be obtained from the ms-auth service.

## Upload Video

```bash
curl -X POST http://localhost:8080/video/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "video=@/path/to/video.mp4"
```

### Supported Video Formats
- MP4 (.mp4)
- AVI (.avi)
- MOV (.mov)
- MKV (.mkv)
- WebM (.webm)

### Constraints
- Maximum file size: 500MB
- Valid JWT token required

### Response
```json
{
  "video_id": "123e4567-e89b-12d3-a456-426614174000",
  "original_name": "video.mp4",
  "status": "pending",
  "message": "Video uploaded successfully and queued for processing"
}
```

## List Videos

```bash
curl -X GET http://localhost:8080/video/list \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Response
```json
{
  "videos": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "original_name": "video.mp4",
      "status": "completed",
      "progress_percent": 100,
      "file_size": 10485760,
      "created_at": "2026-02-23T10:00:00Z",
      "updated_at": "2026-02-23T10:05:00Z"
    }
  ]
}
```

### Video Status
- `pending`: Video uploaded, waiting for processing
- `processing`: Video is being processed
- `completed`: Video processing completed, ready for download
- `failed`: Video processing failed

## Download Video

```bash
curl -X GET "http://localhost:8080/video/download?id=VIDEO_ID" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Response
```json
{
  "presigned_url": "https://s3.amazonaws.com/...",
  "video_id": "123e4567-e89b-12d3-a456-426614174000",
  "file_name": "video.mp4.zip",
  "expires_in": 900
}
```

The presigned URL is valid for 15 minutes (900 seconds).

## Environment Variables

```bash
# AWS Configuration
AWS_REGION=us-east-1
STAGE=dev
AWS_ENDPOINT_URL=http://localstack:4566  # For local development

# SQS Configuration
VIDEO_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/video-processing-dev

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production

# Server Configuration
PORT=8080
```

## Running Locally

### With Docker Compose

```bash
cd packages/ms-video
docker-compose up
```

### Without Docker

1. Install dependencies:
```bash
cd app
go mod download
```

2. Set environment variables:
```bash
export AWS_REGION=us-east-1
export STAGE=dev
export JWT_SECRET=your-secret-key
export VIDEO_QUEUE_URL=your-sqs-queue-url
```

3. Run the service:
```bash
go run cmd/main.go
```

## AWS Resources Required

### S3 Bucket
- Bucket name: `video-system-{stage}` (e.g., `video-system-dev`)
- Folders:
  - `raw/` - Stores uploaded videos
  - `processed/` - Stores processed ZIP files

### DynamoDB Table
- Table name: `videos-{stage}` (e.g., `videos-dev`)
- Primary key: `id` (String)
- Global Secondary Index: `user_id-index`
  - Partition key: `user_id` (String)
  - Sort key: `created_at` (String)

### SQS Queue
- Queue name: `video-processing-{stage}` (e.g., `video-processing-dev`)
- Visibility timeout: 300 seconds (5 minutes)
- Recommended: Configure Dead Letter Queue for failed messages

## Video Processing

The worker (SQS consumer) performs the following steps:

1. Download raw video from S3
2. Update status to "processing" (10% progress)
3. Process the video (simulation: create ZIP file)
4. Update progress at various stages (30%, 60%, 80%)
5. Upload processed ZIP to S3
6. Update status to "completed" (100% progress)

The processed ZIP file contains:
- Original video file
- metadata.txt - Processing metadata
- README.txt - Information about the processed video

## Testing

### Unit Tests
```bash
cd app
go test ./...
```

### Endpoint Tests
Use the automated test script to test all endpoints:

```bash
# Run with default configuration
./test-endpoints.sh

# Or customize the configuration
BASE_URL=http://localhost:8080 \
MS_AUTH_URL=http://localhost:3000 \
TEST_EMAIL=user@example.com \
TEST_PASSWORD=yourpassword \
./test-endpoints.sh
```

The script will:
- Test the health check endpoint
- Obtain a JWT token from ms-auth
- Test video upload (with and without authentication)
- Test listing videos (with and without authentication)
- Test video download (with and without authentication)
- Test error cases (invalid video ID)

**Note:** Make sure both ms-auth and ms-video services are running, and that a user account exists in ms-auth.

### Integration Tests
Use the provided examples to test each endpoint manually or create automated integration tests.

## Development

### Adding New Video Formats
Update the `allowedExtensions` map in `upload_video_usecase.go`.

### Customizing Video Processing
Modify the `Execute` method in `process_video_usecase.go` to implement actual video processing logic (e.g., transcoding, compression, watermarking).

### Adjusting File Size Limits
Update the `MaxVideoSize` constant in `upload_video_usecase.go`.

## Monitoring

The service logs important events:
- Video uploads
- Processing progress
- Queue consumption
- Errors and failures

Integrate with CloudWatch or your logging solution to monitor service health and performance.

## Security

- All video operations require valid JWT authentication
- Users can only access their own videos
- Presigned URLs expire after 15 minutes
- S3 bucket should have proper IAM policies
- Use HTTPS in production

## Deployment

See the `infra/k8s` directory for Kubernetes deployment configurations and `infra/terraform` for infrastructure provisioning.
