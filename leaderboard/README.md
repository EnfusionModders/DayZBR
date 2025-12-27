# Leaderboard

DayZBR Leaderboards Web Service

## Developer Notes

After each match, the aggregate statistics of the match were pushed to this service.
This service processes that data and updates player statistics and leaderboards accordingly.

It exposes the following API endpoints for leaderboard interactions:

- `GET /data/:apikey/ping` - Health check endpoint.
- `POST /data/:apikey/matchsubmit` - Submit match data for processing and leaderboard updates.
- `POST /data/steaminfo` - Get Steam information for a specific player.
- `POST /data/player` - Get player leaderboard information.
- `POST /data/match` - Get match information.
- `POST /data/global` - Get global leaderboard information.
- `POST /data/rank` - Get ranking information.

MongoDB, a Discord Bot, a Steam API Key, and a [seq](https://datalust.co/) logging service are required for this service to function properly.
