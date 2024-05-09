
deploy:
	go build .
	docker build -t dx2wlog2emoncms .
	docker run -d --rm --network=host --name dx2wlog2emoncms dx2wlog2emoncms