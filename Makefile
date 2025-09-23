
clear:
	docker rm -f $$(docker ps -q) && docker-compose -f compose-onlykafka-arm.yml up -d

wsclient:
	go run examples/wsclient/main.go

http:
	curl 127.0.0.1:4242/stats?token=BTC
