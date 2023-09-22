# rest-events

Simple rest service for events

## DB

`make shared-service-setup-db` - Create database and user


## Description

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
