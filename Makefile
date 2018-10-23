
FILES = $(shell find . -type f -name '*.go')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&a{} -> new(a)' -w $(FILES)

deploy:
ifndef tag
	$(error tag is not set)
endif

	docker build -t altipla/k8s-oom-monitor .
	docker tag altipla/k8s-oom-monitor altipla/k8s-oom-monitor:$(tag)
	docker tag altipla/k8s-oom-monitor altipla/k8s-oom-monitor:latest
	docker push altipla/k8s-oom-monitor:$(tag)
	docker push altipla/k8s-oom-monitor:latest
