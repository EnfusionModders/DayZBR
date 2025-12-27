@echo off

docker build -t keganhollern/dayzbr-client-service:latest .
pause
docker push keganhollern/dayzbr-client-service:latest
pause