FROM golang:1.13-alpine as builder

# Setup
RUN mkdir /work

# Add libraries
RUN apk add --no-cache git

# Copy & build
ADD . /work
RUN cd /work && CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -installsuffix nocgo .

# Copy into scratch container
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /work ./
ENTRYPOINT ["./key4-traefik-forward-auth"]
