FROM alpine:3.19.1

RUN apk --no-cache add curl

COPY testMain /app/

WORKDIR app

ENTRYPOINT ["/app/testMain"]
