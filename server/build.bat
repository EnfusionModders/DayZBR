@echo off

docker build -t keganhollern/dayzbr-server-service:latest .
pause
docker push keganhollern/dayzbr-server-service:latest
pause