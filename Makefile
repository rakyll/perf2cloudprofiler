build:
	GOOS=linux GOARCH=amd64 go build -o=./bin/perf2cloudprofiler

push:
	gsutil cp bin/* gs://jbd-releases