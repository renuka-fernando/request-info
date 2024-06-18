# Request Info - Sample Backend Service

Sample backend service which returns request details to the caller.

Open API Specification: [request_info_openapi.yaml](request_info_openapi.yaml)
https://app.swaggerhub.com/apis-docs/renuka-fernando/request-info/2.0.0

## 1. Test Sample Service

```sh
docker run --rm -p 8080:8080 -e "NAME=Service A" renukafernando/request-info:latest -addr :8080 -pretty -logH -logB -statusCode 200 -delayMs 1000
```

- Set `NAME` environment variable to set the name of the service, which will out as a response
- Set `-read-envs` argument if it is required to out environment variables set in container
- set `-pretty` argument if it is required to out prettified JSON
- set `-delayMs` argument to delay response for a given time in milliseconds
- set `-statusCode` argument to set status code of the response
- set `-logH` argument to log request headers
- set `-logB` argument to log request body
- set `-addr` argument to set the address to listen to
- set `-https` argument to enable HTTPS
- set `-key` argument to set the path to the key file
- set `-cert` argument to set the path to the cert file
- set `-mtls` argument to enable mTLS
- set `-ca` argument to set the path to the CA file
- set `-disable-access-logs` argument to disable access logs

Resource Usage
```yaml
resources:
  limits:
    cpu: "5m"
    memory: "10Mi"
  requests:
    cpu: "2m"
    memory: "5Mi"
```

Override statusCode and responseTime with the following query parameters.
- `statusCode` - HTTP status code to respond
- `delayMs` - Time to wait (ms) before responding to request

Get request info:
```sh
curl 'http://localhost:8080/hello/world?delayMs=2000:5000&statusCode=201&pretty=true' -i
```

Empty response:
```sh
curl 'http://localhost:8080/empty?delayMs=1000&statusCode=500' -i
```

Echo response:
```sh
curl 'http://localhost:8080/echo?statusCode=403' -d 'hello world!' -i
```

Set response data:

```sh
curl localhost:8080/req-info/response -d 'hello world' -i
curl localhost:8080/foo -i

curl localhost:8080/req-info/response -X DELETE
```


### 1.1. Basic

Running backend service.
```sh
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A renukafernando/request-info:latest
```

Sending request to backend service.
```sh
curl http://localhost:8080/hello/world
```

```json
{"Name":"Service A","Request":{"Method":"GET","Header":{"Accept":["*/*"],"User-Agent":["curl/7.54.0"]},"URL":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/hello/world","RawPath":"","ForceQuery":false,"RawQuery":"","Fragment":"","RawFragment":""},"ContentLength":0,"TransferEncoding":null,"Host":"localhost:8080","Form":null,"PostForm":null,"MultipartForm":null,"Trailer":null,"RemoteAddr":"172.17.0.1:55922","RequestURI":"/hello/world","TLS":null}}
```

Remove running service
```sh
docker rm -f service-A
```

### 1.2. Print ENV Variables

Running backend service.
```sh
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A renukafernando/request-info:latest -read-envs
```

Sending request to backend service.
```sh
curl http://localhost:8080/hello/world
```

```json
{"Name":"Service A","Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","NAME=Service A","HOME=/root"],"Request":{"Method":"GET","Header":{"Accept":["*/*"],"User-Agent":["curl/7.54.0"]},"URL":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/hello/world","RawPath":"","ForceQuery":false,"RawQuery":"","Fragment":"","RawFragment":""},"ContentLength":0,"TransferEncoding":null,"Host":"localhost:8080","Form":null,"PostForm":null,"MultipartForm":null,"Trailer":null,"RemoteAddr":"172.17.0.1:35306","RequestURI":"/hello/world","TLS":null}}
```

Remove running service
```sh
docker rm -f service-A
```

### 1.3. Pretty JSON Output

Running backend service.
```sh
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A renukafernando/request-info:latest -pretty
```

Sending request to backend service.
```sh
curl http://localhost:8080/hello/world
```

```json
{
    "Name": "Service A",
    "Request": {
        "Method": "GET",
        "Header": {
            "Accept": [
                "*/*"
            ],
            "User-Agent": [
                "curl/7.54.0"
            ]
        },
        "URL": {
            "Scheme": "",
            "Opaque": "",
            "User": null,
            "Host": "",
            "Path": "/hello/world",
            "RawPath": "",
            "ForceQuery": false,
            "RawQuery": "",
            "Fragment": "",
            "RawFragment": ""
        },
        "ContentLength": 0,
        "TransferEncoding": null,
        "Host": "localhost:8080",
        "Form": null,
        "PostForm": null,
        "MultipartForm": null,
        "Trailer": null,
        "RemoteAddr": "172.17.0.1:44222",
        "RequestURI": "/hello/world",
        "TLS": null
    }
}
```

Remove running service
```sh
docker rm -f service-A
```

### 1.4. HTTP Service

Generate Certs

```sh
./gen-certs.sh
```

Running backend service.

#### 1.5.1. Using Docker Image

```sh
docker run --rm -p 8443:8443 -e "NAME=Service A" -v ./certs:/certs  renukafernando/request-info:latest -pretty -logH -logB -addr :8443 -https -key /certs/server.key -cert /certs/server.crt
```

Test the service

```sh
curl https://localhost:8443/foo \
    --cacert certs/server.crt \
    --key certs/client.key \
    --cert certs/client.crt \
    -v
```

#### 1.5.2. Using Go Source Code

```sh
go run main.go -pretty -logH -logB -addr :8443 -https -key ./certs/server.key -cert ./certs/server.crt
```

Test the service
```sh
curl https://localhost:8443/foo \
    --cacert certs/server.crt \
    -v
```

### 1.5. mTLS Service

Generate Certs

```sh
./gen-certs.sh
```

Running backend service.

#### 1.5.1. Using Docker Image

```sh
docker run --rm -p 8443:8443 -e "NAME=Service A" -v ./certs:/certs  renukafernando/request-info:latest -pretty -logH -logB -addr :8443 -https -key /certs/server.key -cert /certs/server.crt -mtls -ca /certs/client.crt
```

Test the service

```sh
curl https://localhost:8443/foo \
    --cacert certs/server.crt \
    --key certs/client.key \
    --cert certs/client.crt \
    -v
```

#### 1.5.2. Using Go Source Code

```sh
go run main.go -pretty -logH -logB -addr :8443 -https -key ./certs/server.key -cert ./certs/server.crt -mtls -ca ./certs/client.crt
```

Test the service

```sh
curl https://localhost:8443/foo \
    --cacert certs/server.crt \
    --key certs/client.key \
    --cert certs/client.crt \
    -v
```


## 2. Build From Source

Execute the following command to build the Docker image.

```sh
./build-docker.sh
```

## 3. Deploy in Choreo

1. Create a Service component in Choreo.
2. Give name and description to the service.
3. Select this as the GitHub account. Select the repository as `request-info` and branch as `main`.
4. Set build preset as `Dockerfile`.
5. Set Dockerfile path as `DockerfileFullBuild`.
6. Set Docker context as `/`.
