# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
# was /go/src/github.com/tarikjn/goblin-proxy
WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

# Build the command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o proxy

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary to the production image from the builder stage.
COPY --from=builder /src/proxy /proxy

# temporary, will be set later
ENV PROXY_ENV prod
ENV PORT 8080

EXPOSE 8080

# Perform any further action as an unprivileged user.
#USER nobody:nobody

# Run the web service on container startup.
ENTRYPOINT ["/proxy"]

# check:
# - https://medium.com/@pierreprinetti/the-go-1-11-dockerfile-a3218319d191
# - https://www.callicoder.com/docker-golang-image-container-example/
# - https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
