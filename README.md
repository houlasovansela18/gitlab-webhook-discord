# GitLab Webhook to Discord

This project provides a Go-based service that listens for GitLab webhook events and forwards them to a Discord channel in a structured and readable format. It supports multiple event types, including repository updates, push events, tag push events, and merge request events.

## Features

- Converts GitLab webhook events into Discord messages.
- Supports the following GitLab events:
  - Repository update events
  - Push events
  - Tag push events
  - Merge request events
- Easy to configure via Docker Compose and environment variables.

## Prerequisites

- Docker
- Docker Compose
- A Discord webhook URL
- A GitLab instance configured to send webhooks

## Environment Variables

Ensure you have the following environment variables set:

- `DISCORD_WEBHOOK_URL`: The Discord webhook URL where messages will be sent.

## Running the Application

You can run the application using Docker Compose.

## Prepare the `.env` file

Create a `.env` file in the root of your project and include the `DISCORD_WEBHOOK_URL`:

```bash
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/your-webhook-id/your-webhook-token
```
