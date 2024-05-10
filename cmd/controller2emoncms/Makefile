
deploy:
	go build .
	docker build -t controller2emoncms .
	docker run -d --rm --network=host --name controller2emoncms controller2emoncms