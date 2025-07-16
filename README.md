
# Vivy Auth Service

**Vivy** | Video + Voice + You - A lightweight, friendly, and versatile real-time communication application.

`vivy-chat` is the Chat Service within Vivy's microservices architecture. It handles user login, registration, token issuance, and authentication. The service is built with **Golang** and uses **gRPC** for high-performance communication with other components.

## Overview

`vivy-chat` is the core of Vivy's security system, ensuring that all users and requests are authenticated before accessing services like chat, voice call, or video call. 

### Key Features

- **Conversation Management**: Create, retrieve, and manage one-on-one and group conversations.
- **Rich Messaging**: Send and receive various message types, such as text and images.
- **Message History**: List messages within a conversation with support for pagination.
- **Message Status**: Track the read status of messages and mark them as read.
- **Real-time Updates**: Experience live updates with real-time message streaming and typing indicators.
- **User Sync**: Keep user information synchronized across services.

## Architecture & Technologies

`vivy-chat` is designed as a microservice and integrates into the Vivy ecosystem with the following stack:

- **Language**: Golang - High performance and concurrency support.
- **Communication**: gRPC - Fast and efficient inter-service communication.
- **Database**: PostgreSQL (stores user information), Redis (session and cache management).


## Installation & Deployment

### System Requirements
- **Golang**: Version 1.18+
- **PostgreSQL**: Version 13+ (user data storage)
- **Docker** (optional): For containerized deployment
- **protoc**: Protocol Buffers compiler for generating gRPC code
- **Redis**: Version 6+ (session and cache management)
