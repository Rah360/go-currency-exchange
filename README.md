# go-currency-exchange

real-time currency exchange rate updates using Golang, WebSocket, and Redis

## Features

- Real-time exchange rate updates via WebSocket.
- API to update exchange rates.
- Use of Redis for storing and broadcasting exchange rates.
- Separate roles for admin and clients.

## Prerequisites

## Installation

Clone the repository to your local machine

Configuration
Ensure your Redis server is up and running.
currently it usages default configuration for redis

Running the Application
To run the server, execute:

```
go run main.go server.go client.go storage.go
```

The server will start listening for WebSocket connections on ws://localhost:3000/ws

### Endpoint: Update Currency Rate
To update exchange rates, send a POST request with the following JSON payload
Only admin can use this api

| Field            | Description                                          |
|------------------|------------------------------------------------------|
| **URL**          | `/update`                                        |
| **Method**       | `POST`                                                |
| **Headers**      | `Authorization: <TOKEN>`                     |
| **Payload** | `{"currency": "USD","rate":"200"}` |

HTTP requests on http://localhost:3000.
create a sample jwt token with secret key value "test_key" with role= admin

### Running client
once server is running you can run the client using 
```
 go run testClient/testClient.go
```
this dummy client will start receiving new currency updates via redis pubsub

