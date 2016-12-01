
build:
	go build -tags netgo -o argoos


docker-image:
	docker build -t smileoss/argoos docker

