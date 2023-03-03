# tracing-test
- This source code is a small sample about how to use OpenTelemetry combine Jaeger and ElasticSearch to trace data in service of distribute system.
- In this example, I use ```Echo``` to make some APIs. 
- Tracing data is handled by ```Otel``` (SDK of ```OpenTelemetry```), it use http handler to integrate in middleware of ```Echo``` to get data tracing.


#### Step 1: Pull code from github to local:
``` git clone https://github.com/namphamtoday/tracing-test.git ```

#### Step 2: Init ElasticSearch and Kibana Container:
``` docker compose -f docker-compose-elasticsearch.yaml -d ```

#### Step 3: Init Jaeger components container:

``` docker compose up -d ```

After all components are run, you can access this url to view Jaeger UI:<br>
```http://localhost:16686/```

#### Step 4: After you init all components, you can work with example code:

Run ``` go mod tidy ``` , if you get error while download packages, you can run:

```
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
go get go.opentelemetry.io/otel/exporters/jaeger
go get go.opentelemetry.io/otel/sdk/resource
go get go.opentelemetry.io/otel/sdk/trace
```

Then you run ```go run main.go``` to run program <br>
There are some cURL to call example api, you can import to Postman and call it:

Example GET method: <br>
```
curl --location 'http://localhost:8080/hello?name=Thinh&class=9A1&school=Nguyen%20Du'
```

Example POST method:<br>
```
curl --location 'http://localhost:8080/foo' \
--header 'Content-Type: application/json' \
--data '{
    "name": "Huong",
    "class": "12A6",
    "school": "Nguyen Dinh Chieu"
}'
```

Some images of tracing results:
![image](https://user-images.githubusercontent.com/63083419/222696845-967354f5-f719-4dd6-89fb-0e8dc84d0838.png) 

![image](https://user-images.githubusercontent.com/63083419/222696985-91a921ea-afee-4871-b75f-85c2ecee87e9.png)

![image](https://user-images.githubusercontent.com/63083419/222697259-733e2d6e-9af7-4415-bf8b-95f1f3eb51e5.png)




### Reference:
https://www.jaegertracing.io/ <br>
https://opentelemetry.io/ <br>
https://echo.labstack.com/ <br>
 


