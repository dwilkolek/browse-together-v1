.PHONY: dev_redis dev start_redis release fly dev_redis_8081

dev_redis:
	QUEUE=REDIS STORAGE=REDIS go run main.go  
dev_redis_8081:
	QUEUE=REDIS STORAGE=REDIS PORT=8081 go run main.go  

dev:
	go run main.go  

fly:
	flyctl deploy