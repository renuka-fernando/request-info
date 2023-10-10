# Request Info - Sample Backend Service
Sample backend service which returns request details to the caller.

Open API Specification: [request_info_openapi.yaml](request_info_openapi.yaml)
https://app.swaggerhub.com/apis-docs/renuka-fernando/request-info/2.0.0

## 1. Test Sample Service

```sh
docker run --rm -p 8080:8080 -e "NAME=Service A" cakebakery/request-info:v2 -addr :8080 -pretty -logH -logB -statusCode 200 -delayMs 1000
```

- Set `NAME` environment variable to set the name of the service, which will out as a response
- Set `-read-envs` argument if it is required to out environment variables set in container
- set `-pretty` argument if it is required to out prettified JSON
- set `-delayMs` argument to delay response for a given time in milliseconds
- set `-statusCode` argument to set status code of the response

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
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A cakebakery/request-info:v2
```

Sending request to backend service.
```
curl http://localhost:8080/hello/world
```

```json
{"Name":"Service A","Request":{"Method":"GET","Header":{"Accept":["*/*"],"User-Agent":["curl/7.54.0"]},"URL":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/hello/world","RawPath":"","ForceQuery":false,"RawQuery":"","Fragment":"","RawFragment":""},"ContentLength":0,"TransferEncoding":null,"Host":"localhost:8080","Form":null,"PostForm":null,"MultipartForm":null,"Trailer":null,"RemoteAddr":"172.17.0.1:55922","RequestURI":"/hello/world","TLS":null}}
```

Remove running service
```
docker rm -f service-A
```

### 1.2. Print ENV Variables

Running backend service.
```sh
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A cakebakery/request-info:v2 -read-envs
```

Sending request to backend service.
```
curl http://localhost:8080/hello/world
```

```json
{"Name":"Service A","Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","NAME=Service A","HOME=/root"],"Request":{"Method":"GET","Header":{"Accept":["*/*"],"User-Agent":["curl/7.54.0"]},"URL":{"Scheme":"","Opaque":"","User":null,"Host":"","Path":"/hello/world","RawPath":"","ForceQuery":false,"RawQuery":"","Fragment":"","RawFragment":""},"ContentLength":0,"TransferEncoding":null,"Host":"localhost:8080","Form":null,"PostForm":null,"MultipartForm":null,"Trailer":null,"RemoteAddr":"172.17.0.1:35306","RequestURI":"/hello/world","TLS":null}}
```

Remove running service
```
docker rm -f service-A
```

### 1.3. Pretty JSON Output

Running backend service.
```sh
docker run -d -p 8080:8080 -e "NAME=Service A" --name service-A cakebakery/request-info:v2 -pretty
```

Sending request to backend service.
```
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
```
docker rm -f service-A
```

## 2. Build From Source
run `./build-docker.sh`

## 3. Deploy in Choreo

1. Create a Service component in Choreo.
2. Give name and description to the service.
3. Select this as the GitHub account. Select the repository as `request-info` and branch as `main`.
4. Set build preset as `Dockerfile`.
5. Set Dockerfile path as `DockerfileFullBuild`.
6. Set Docker context as `/`.
