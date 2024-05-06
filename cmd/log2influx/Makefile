
deploy:
	go build .
	docker build -t dx2wlog2influx .
	docker run -d --rm --network=host --name dx2wlog2influx dx2wlog2influx