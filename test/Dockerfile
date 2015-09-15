FROM gliderlabs/alpine:3.2
RUN apk update && apk add bash curl ca-certificates && rm -rf /var/cache/apk/*
ADD cbc-benchmark bin/

CMD ["cbc-benchmark"]
