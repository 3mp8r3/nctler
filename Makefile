timestamp=$(shell date +%Y%m%d)
cur_path=$(shell cd ../ && pwd)
image_name=maoqide/nctler:dev-${timestamp}
help:
	@echo ''
	@echo 'Optional Commands: '
	@echo ''
	@echo '	make help/make usage'
	@echo '	make pkg-ensure'
	@echo '	make build-bin'
	@echo '	make build-image'
	@echo '	make build-onestep'
	@echo '	make build-and-push'
	@echo '	make push-image'
	@echo '	make run-container'
	@echo ''
usage: help
default: help
pkg-ensure:
	export GOPATH=${GOPATH}:${cur_path} && dep ensure -v
build-bin:
	sh ./build.sh
build-image:
	docker build -t ${image_name} .
push-image:
	docker push ${image_name}
build-onestep: build-bin build-image
build-and-push: build-onestep push-image
run-container:
	docker run -d --net host \
    --restart=always \
    -u root \
    --pid=host \
    --privileged \
    -e REDIS_ADDR=127.0.0.1:6379 \
    -e REDIS_DB=0 \
    -e DOCKER_ENDPOINT=unix:///var/run/docker.sock \
    -e DOCKER_API_VERSION=1.20 \
    -v /lib64/libdevmapper.so.1.02:/lib64/libdevmapper.so.1.02:ro \
    -v /lib64/libdevmapper-event.so.1.02:/lib64/libdevmapper-event.so.1.02:ro \
    -v /usr/bin/docker:/usr/bin/docker \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/lib64/libltdl.so.7:/usr/lib64/libltdl.so.7 \
    ${image_name}

