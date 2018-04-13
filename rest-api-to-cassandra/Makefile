.PHONY: all
all: build

.PHONY: clean
clean:
	docker rmi -f ${REPO}/${IMAGE}:${TAG} || :

.PHONY: build
build:
	docker images
	docker build \
		--build-arg BUILD_DATE=${BUILD_DATE} \
		--build-arg VERSION=${VERSION} \
		--build-arg SERVICEPORT=${SERVICEPORT} \		
		--no-cache \
		--tag ${REPO}/${IMAGE}:${TAG} ${GIT_REPO_LINK}
	docker images | grep ${IMAGE}

.PHONY: push
push: clean
