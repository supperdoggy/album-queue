# Spotify Link Collector Bot

## Overview

This is a Telegram bot written in Golang that collects Spotify playlist, album, and song links, and adds them to a queue for download. The bot listens for user inputs, processes Spotify links, and queues them in a designated download system.

## Features

- Accepts Spotify links for playlists, albums, or songs.
- Automatically detects and validates Spotify URLs.
- Queues the links for download.
- Provides feedback to the user on whether the link was successfully added to the queue.
- Logs all incoming links and actions for auditing and debugging purposes.
  
## Prerequisites

- Golang 1.21 or above
- A Spotify API client with proper access tokens and permissions
- A Telegram bot token (generated from the BotFather in Telegram)
- Redis or any queueing system to manage download requests (optional)

## Installation

1. Clone the repository:

    ```bash
    git clone git@github.com:DigitalIndependence/album-queue.git
    cd album-queue
    ```

2. Install the required dependencies:

    ```bash
    go mod tidy
    ```

3. Set up environment variables:
    
    You can create a `.env` file in the project root directory for the following variables:
    
    ```env
    TELEGRAM_BOT_TOKEN=<your-telegram-bot-token>
    SPOTIFY_CLIENT_ID=<your-spotify-client-id>
    SPOTIFY_CLIENT_SECRET=<your-spotify-client-secret>
    REDIS_URL=<redis-url-if-used>
    QUEUE_NAME=<name-of-the-queue>
    ```

4. Run the bot:

    ```bash
    go run main.go
    ```

## Usage

1. Add the bot to your Telegram and start a chat.
2. Send any Spotify playlist, album, or song link.
3. The bot will validate the link, and if it's correct, add it to the download queue.
4. You will receive a confirmation message if the link was added to the queue successfully.

## Example Commands

- Send a Spotify song link:

    ```
    https://open.spotify.com/track/1234567890abcdefghij
    ```

- Send a Spotify album link:

    ```
    https://open.spotify.com/album/abcdefghij1234567890
    ```

- Send a Spotify playlist link:

    ```
    https://open.spotify.com/playlist/abcdefghij0987654321
    ```

## Queue System

This bot uses a queue to manage download requests. By default, it supports Redis as the queueing system. If you prefer another queue system, make sure to update the relevant code in the `queue.go` file.

## Logging

Logs are generated for each incoming Spotify link, including whether the link was valid and successfully added to the queue. Check the logs for debugging or to monitor usage.

## Spotify API

To interact with Spotify, the bot uses the [Spotify Web API](https://developer.spotify.com/documentation/web-api/). Ensure you have set up your Spotify Developer account and generated the appropriate API credentials. 

## Contributing

1. Fork the repository.
2. Create your feature branch: `git checkout -b feature/my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature/my-new-feature`
5. Submit a pull request!

## License

This project is licensed under the MIT License.
