# Browse Together API

## Development
- Redis : `docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest`
- Start Backend Server : `ENV=dev go run main.go`


## Release

`aws ecr get-login-password --region eu-west-1 --profile padmin | docker login --username AWS --password-stdin 871274668106.dkr.ecr.eu-west-1.amazonaws.com`

`docker build . --platform=linux/amd64 -t 871274668106.dkr.ecr.eu-west-1.amazonaws.com/browse-together-repo:0.0.1-alpha`

`docker push 871274668106.dkr.ecr.eu-west-1.amazonaws.com/browse-together-repo:0.0.1-alpha`

