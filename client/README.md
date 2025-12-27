# Client

DayZBR Client Service

## Developer Notes

This service is responsible for handling client API requests.
So all players will be querying this service to interact with the backend
for DayZ Battle Royale.

It requires MongoDB, a Discord Bot, and a [seq](https://datalust.co/) logging service.

It exposes the following API endpoints for client interactions:

- `GET /client/ping` - Health check endpoint.
- `POST /client/start` - Player login / session start. Runs at game launch.
- `POST /client/matchmake` - Matchmaking request.
- `POST /client/player` - Get player information.
- `POST /client/match` - Get game/match information.
- `POST /client/server` - Get server information.
