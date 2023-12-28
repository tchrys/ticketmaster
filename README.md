## About
This is a PoC project which implements an event registration web server similar to Ticketmaster. The main purpose of this project is to leverage an interesting architecture based on Redis and Golang and see how far it can go load-wise (I tested it only locally, but the results are quite remarkable). I was focused only on making the registration step as scalable as possible, so this project has only 4 endpoints and minimal business logic, but with some improvements it can act as a standalone microservice in a complex real-app architecture.
The registration flow consists of the following steps:
1. Someone creates an event, providing the capacity, registration start time and a registration deadline
2. Users have to request a token for an event, and once they get it they can call another endpoint to register for that event. This should prevent a DDoS attack (if a malicious actor tries to acquire a large share of tickets)
3. After the registration window ends, there is a 60s cool-down window which allows the web server to find the winners. After this window, users can call the API to find out the results (if they won a ticket or not)

## Scale test
The scale test has 3 parameters:
1. Capacity = how many tickets are available
2. DeadlineDelta = the length of the registration interval in seconds
3. DemandRatio = how many users are interested in getting tickets. If the capacity is 10k and the demand ratio is 20, it means that 200k users will register

Once the event is created, every user makes a request to get a token then another to register. The load is spread evenly. After the 60s cool-down period, they call the API again to find the result
I tested the app locally with various parameters to find its limits. It seems that it survived 100k concurrent users per minute without any error, which is impressive ! (To be precise, I set this parameters: Capacity = 10k, DemandRatio = 20, DeadlineDelta = 120s).

## Setup

1. Start the Redis instance using this command : `sudo docker run -p 6379:6379 -it -d --rm redis/redis-stack-server:latest`
2. `go mod tidy`
3. `go build`
4. `./ticketmaster`
5. If you want to use the default parameters when running scale tests, you have to increase the number of open file descriptors using `ulimit -n <number>` . When I tested the app, I used 64k.
6. Run scale tests using `go run scale_test/test.go`
