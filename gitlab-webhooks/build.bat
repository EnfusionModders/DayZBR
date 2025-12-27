@echo off

docker build -t keganhollern/dayzbr-gitlab-webhooks:latest .
pause
docker push keganhollern/dayzbr-gitlab-webhooks:latest
pause