# Local Setup
There are 2 ways you can run Veriflow locally. Either use docker-compose which spins up both Veriflow and DynamoDB or run them seperately

### Using docker-compose
This will spin up the veriflow container with a dyanmodb. 
```
# Build the GO container
CGO_ENABLED=0 GOOS=linux go build -o main -v cmd/main.go
docker-compose build

# Start Veriflow and local DynamoDB
docker-compose up -d

# execute the script to create the tables required by Veriflow
./scripts/create_local_tables.sh
```

You can test if Veriflow is running by doing a curl request `curl http://127.0.0.1:8080/health`

### Run from source
#### Setting up DynamoDB locally
Download DynamoDB local jar from https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html#DynamoDBLocal.DownloadingAndRunning.title and extract the zip.

```
# Start DynamoDB
java -Xms256m -Xmx2048m -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb
```

You can check if dynamodb is runnning at http://localhost:8000 by using AWS CLI `aws dynamodb list-tables --endpoint-url http://localhost:8000`

Now let's create 3 tables needed for Veriflow `VeriflowUsers` and  `VeriflowRequests` and `VeriflowWebAuthnSessions`. 
Execute the script below to create them. 
```
./scripts/create_local_tables.sh
```

#### Running Veriflow
2 Ways to run Veriflow. You can either use SAM which will simulate your AWS using `make build && sam local start-api -p 8080` or run directly from source. 
```
export LOCAL=true
go mod tidy
go run cmd/main.go
```
You can test if Veriflow is running by doing a curl request `curl http://127.0.0.1:8080/health`

### Ngrok setup to expose your API endpoints
In order to use Slack slash commands, you need a public facing endpoint. For local environment you can use ngrok. 
- Head on to https://dashboard.ngrok.com/signup to sign up and get your auth token. 
- `brew install ngrok/ngrok/ngrok`
- `ngrok config add-authtoken 1y6pVfrk2lTnXfLyYIHFy2JPdYV_6AfmDBhXqYdKwr2u4mKPy`
- `ngrok http http://localhost:8080`

Now your application is running and copy the ngrok URL. 

