# albion-killgot

Discord bot for tracking kills, deaths, and battles in Albion Online. Written in Go.

## Quick start

### Download

Grab the latest binary from [Releases](https://github.com/phx7/albion-killgot/releases).

### Configure

Create a `.env` file next to the binary:

```env
DISCORD_TOKEN=your_bot_token_here
```

Get a token at https://discord.com/developers/applications — create an app, add a Bot, copy the token.

Required bot permissions: `Send Messages`, `Embed Links`  
Required scopes: `bot`, `applications.commands`

### Run

**Linux:**
```bash
chmod +x bot-linux-amd64
./bot-linux-amd64
```

**Windows:**  
Double-click `bot-windows-amd64.exe` or run from PowerShell.

### Run as a service (Linux)

```bash
sudo cp bot-linux-amd64 /opt/albion-killgot/bot
sudo cp .env /opt/albion-killgot/.env

sudo tee /etc/systemd/system/albion-killgot.service > /dev/null <<EOF
[Unit]
Description=Albion Killgot Discord Bot
After=network.target

[Service]
ExecStart=/opt/albion-killgot/bot
WorkingDirectory=/opt/albion-killgot
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now albion-killgot
sudo journalctl -fu albion-killgot
```

## Commands

| Command | Description |
|---------|-------------|
| `/track player` | Track a player |
| `/track guild` | Track a guild |
| `/track alliance` | Track an alliance |
| `/untrack player/guild/alliance` | Stop tracking |
| `/list` | Show all tracked entities |
| `/settings kills` | Configure kill notifications |
| `/settings deaths` | Configure death notifications |
| `/settings battles` | Configure battle notifications |
| `/settings juicy` | Configure high-value kill notifications |
| `/settings show` | Show current settings |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DISCORD_TOKEN` | required | Discord bot token |
| `DB_PATH` | `killbot.db` | Path to SQLite database |

## Building from source

```bash
go build -o bot ./cmd/bot/
```
