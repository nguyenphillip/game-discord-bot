# game-discord-bot
Discord Bot made using Discordgo library used to start, stop, and list AWS EC2 resources.
## Build
```
go build -o ./bin ./cmd/...
```

## Generate AES Key
```
./bin/generate_key
```

## Running the bot
Run the bot with below commands using either the parameters or setting environment variables:
- `VBOT_TOKEN`: Discord Bot Token
- `VBOT_GUILD_ID`: Discord Guild ID
- `VBOT_AES_KEY`: AES Key used for encrypting and decrypting
- `DATABASE_URL`: Database connection string
```
./bin/bot --token <token> --guild <id> --db <connection_url> --key <key> 
```