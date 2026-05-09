# albion-kill<img src="https://go.dev/blog/go-brand/Go-Logo/SVG/Go-Logo_Blue.svg" height="52" alt="go">t

Discord bot for tracking kills, deaths, and battles in Albion Online. Written in Go.

## Quick start

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/phx7/albion-killgot/main/install.sh | sudo bash
```

The script will download the latest binary, ask for your Discord bot token, and install it as a systemd service.

### Windows

Download `bot-windows-amd64.exe` from [Releases](https://github.com/phx7/albion-killgot/releases), create a `.env` file next to it:

```env
DISCORD_TOKEN=your_bot_token_here
```

Then run the `.exe`.

### Getting a bot token

1. Go to https://discord.com/developers/applications
2. Create an application → Bot → Reset Token → copy it
3. OAuth2 → URL Generator → scopes: `bot`, `applications.commands` → permissions: `Send Messages`, `Embed Links` → invite the bot to your server

## Commands

| Command | Description |
|---------|-------------|
| `/track player` | Track a player |
| `/track guild` | Track a guild |
| `/track alliance` | Track an alliance by ID |
| `/untrack player/guild/alliance` | Stop tracking |
| `/list` | Show all tracked entities |
| `/settings kills` | Configure kill notifications |
| `/settings deaths` | Configure death notifications |
| `/settings battles` | Configure battle notifications |
| `/settings juicy` | Configure high-value kill notifications |
| `/settings show` | Show current settings |
| `/settings permit` | Grant a role or user permission to use bot commands |
| `/settings revoke` | Revoke a role or user permission |
| `/test kill` | Send a test kill notification |
| `/test death` | Send a test death notification |

> `/settings` commands are restricted to the server owner. All other commands require the server owner or an explicitly permitted role/user.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DISCORD_TOKEN` | required | Discord bot token |
| `DB_PATH` | `killbot.db` | Path to SQLite database |

## Building from source

```bash
go build -o bot ./cmd/bot/
```
