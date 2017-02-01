TAGNAME:="master"


build:
	bash tools/deploy.sh build $(TAGNAME)

release: build
	bash tools/deploy.sh release $(TAGNAME)

deploy: release
	bash tools/deploy.sh upload $(TAGNAME)

docker-image:
	docker build -t smileoss/argoos docker

