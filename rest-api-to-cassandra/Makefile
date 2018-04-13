.PHONY: all
all: build

.PHONY: clean
clean:
	docker container stop ${CONTAINER} || :
	docker container rm ${CONTAINER} || :
	docker rmi -f ${REPO}/${IMAGE}:${TAG} || :

.PHONY: build
build:
	docker build \
		--build-arg BUILD_DATE=${BUILD_DATE} \
		--build-arg VERSION=${VERSION} \
		--build-arg SERVICEPORT=${SERVICEPORT} \
		--build-arg CONTAINER=${CONTAINER} \
		--no-cache \
		--tag ${REPO}/${IMAGE}:${TAG} ${GIT_REPO_LINK}
	docker images | grep ${IMAGE}

.PHONY: push
push:
	docker image push ${REPO}/${IMAGE}:${TAG}

.PHONY: deploy
deploy:
	docker run --net vnet --name ${CONTAINER} -p 8080:8080 -d ${REPO}/${IMAGE}:${TAG}

.PHONY: logs
logs:
	docker logs --details ${CONTAINER}
