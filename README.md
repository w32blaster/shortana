# Shortana
Just a simple plain URL shortener managed by Telegram bot. Using bot you can add, delete
your short URLs and get a view statistics.

To install Shortana to your server, you need:

1. create a new Telegram Bot
2. create account at maxmind.com to download fresh GeoIP database (needed to get country and town by IP address)
3. buy a new domain, short one :)

## How to create a telegram bot
1. Create a new bot using @BotFather
2. Copy the API KEY
3. Start a new chat with freshly created new bot and type something
4. (optional, only for local development) if you want your local installation
    is accessible from the internet, the most convenient way is to use Ngrok.
    ```
     ./ngrok http http://localhost:8444 -region eu
    ```
    this command will expose your localhost to the public and print out the URL.
5. Open your console and call the command:
    ```
    curl -F "url=https://1.2.3.4:8443/<bot API Key>" https://api.telegram.org/bot<bot API key>/setWebhook
    ```
    where 1.2.3.4 is your actual IP or URL, where the backend is running,
    for the local development it is the Ngrock URL, such as https://blablabla.eu.ngrok.io 
6. now, if you print something in your bot, the request will be propagated to your locally running app

# Docker
Official Docker image can be found here: 
https://hub.docker.com/repository/docker/w32blaster/shortana

# Credits

This product includes GeoLite2 data created by MaxMind, available from
https://www.maxmind.com.