dev_redis:
	QUEUE=REDIS STORAGE=REDIS go run main.go  

dev:
	go run main.go  

start_redis:
    docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest

release:
	aws ecr get-login-password --region $(aws_region) | docker login --username AWS --password-stdin $(aws_account).dkr.ecr.$(aws_region).amazonaws.com
	docker build . --platform=linux/amd64 -t $(aws_account).dkr.ecr.$(aws_region).amazonaws.com/$(aws_repo):$(version)
	docker push $(aws_account).dkr.ecr.$(aws_region).amazonaws.com/$(aws_repo):$(version)

fly:
	flyctl deploy