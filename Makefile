TAGNAME:="master"


build:
	bash tools/deploy.sh build $(TAGNAME)
	mv argoos docker

release: build
	bash tools/deploy.sh release $(TAGNAME)

deploy: release
	bash tools/deploy.sh upload $(TAGNAME)

docker-image:
	docker build --no-cache -t smilelab/argoos docker


