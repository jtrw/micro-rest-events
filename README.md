# rest-events


## DB

`make shared-service-setup-db` - Create database and user


## Description

POST - `/events` Create new event
POST - `/events/{uuid}/status/in_progress` Change status of event in progress
POST - `/events/{uuid}/status/done` Change status of event done
POST - `/events/{uuid}/status/error` Change status of event error

GET - `/events/users/{id}` Get all events by user id

POST - `/events/{uuid}/seen/` Mark event as seen   
