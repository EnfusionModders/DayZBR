@echo off

docker build -t keganhollern/dayzbr-leaderboard-service:latest .
pause
docker push keganhollern/dayzbr-leaderboard-service:latest
pause