# WISdoM OSS - Consumer Management Service
<p>
<img src="https://img.shields.io/github/go-mod/go-version/wisdom-oss/service-consumers?filename=src%2Fgo.mod&style=for-the-badge" alt="Go Lang Version"/>
<a href="openapi.yaml">
<img src="https://img.shields.io/badge/Schema%20Version-3.0.0-6BA539?style=for-the-badge&logo=OpenAPI%20Initiative" alt="Open
API Schema Version"/></a>
</p>

## Overview
This microservice is responsible for managing consumers and their associated 
data. 
It is a part of the WISdoM OSS project.
It uses the microservice template for the WISdoM OSS project.

## Using the service
The service is included in every WISdoM OSS deployment by default and does not 
require the user to do anything.

A documentation for the API can be found in the [openapi.yaml](openapi.yaml) file in the 
repository.

## Request Flow
The following diagram shows the request flow of the service.
```mermaid
sequenceDiagram
    actor U as User
    participant G as API Gateway
    participant S as Consumer Service
    participant D as Database
    
    U->>G: New Request
    activate G
    G->>G: Check authentication
    alt authentication failed 
        note over U,G: Authentication may fail due to<br/>invalid credentials or missing<br/>headers
        G->>U: Error Response 
    else authentication successful
        G->>S: Proxy request
        activate S
    end
    S-->S: Check authentication information for explicit group
    deactivate G
    activate D
    S-->D: Query consumer data
    deactivate D
    S-->S: Build response
    S->>G: response
    deactivate S
    G->>U: deliver response
    
```

## Development
### Prerequisites
- Go 1.20

### Important notices
- Since the service is usually deployed behind an API gateway which
    authenticates the user, the service does reject all requests which do not
    contain the `X-Authenticated-Groups` and `X-Authenticated-User` header.

    You need to set those headers manually when testing the service locally.