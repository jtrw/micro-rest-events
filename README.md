# rest-events

Simple rest service for events

## DB

`make shared-service-setup-db` - Create database and user


## Description

## Overview

The Micro-Rest-Events service is designed to handle RESTful operations for managing events. It provides endpoints to create, update, and retrieve events, as well as manage event statuses and user interactions.

## Features

### 1. Create Single Event
Endpoint: `POST /api/v1/events`

Allows the creation of a single event with the following parameters:
- `type`: Type of event
- `user_id`: User identifier
- Optional: `caption`, `body`, `status`

Example:
```go
// Example request payload
{
    "type": "notification",
    "user_id": "12345",
    "caption": "New Notification",
    "body": "This is a notification message.",
    "status": "new"
}
```

### 2. Create Batch Events

Endpoint: POST `/api/v1/events/batch`

Enables the creation of multiple events for different users simultaneously by providing:

- `type`: Type of event
- `users`: Array of user identifiers
Example:

```
// Example request payload
{
    "type": "alert",
    "users": ["user1", "user2", "user3"]
}
```

### 3. Update Event Status
Endpoint: POST `/api/v1/events/{uuid}`

Updates the status of a specific event by UUID.

Example:
```
// Example request payload
{
    "status": "resolved"
}
```

### 4. Get Events by User ID
Endpoint: GET `/api/v1/events/users/{id}`

Retrieves events based on a specific user identifier.

### 5. Set Event as Seen
Endpoint: POST `/api/v1/events/{uuid}/seen`

Marks an event as "seen."

## Usage
To use the Micro-Rest-Events service, follow these steps:

### Prerequisites

1. GoLang environment set up.
1. Ensure required dependencies are installed (go get command might be needed).

### Configuration

Configure service settings:
1. Modify .env file or set environment variables (e.g., POSTGRES_DSN, LISTEN, etc.).
1. Ensure a PostgreSQL database connection is available.

### Running the Service

1. Build and run the application using the GoLang command: `go run main.go`
1. Service starts running on the specified address (default: localhost:8080).
1. Access the various endpoints to perform CRUD operations on events.

### Notes

1. Ensure proper error handling for all API calls.
1. Secure endpoints using authentication mechanisms if deploying in a production environment.
1. Monitor service logs for errors and warnings.


POST - `/events` Create new event
POST - `/events/{uuid}` Update event
GET - `/events/users/{id}` Get all events by user id
POST - `/events/{uuid}/seen/` Mark event as seen



### For feature 
POST - `/events/{uuid}/status/in_progress` Change status of event in progress
POST - `/events/{uuid}/status/done` Change status of event done
POST - `/events/{uuid}/status/error` Change status of event error

## Test
`go test -v app/backend -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html`
