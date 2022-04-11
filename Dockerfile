FROM registry-hz.rubikstack.com/ci/golang:1.17.6-b3da33d-20220119030217 AS builder

# Set necessary environmet variables needed for our image
#    CGO_ENABLED=0 \
ENV GOOS=linux \
    CGO_ENABLED=0 \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN make build

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/k8sync .

# Build a small image
FROM alpine
WORKDIR /app
COPY --from=builder /dist/k8sync /app

# Export necessary port
EXPOSE 8000
EXPOSE 8001

# Command to run
ENTRYPOINT ["/app/k8sync"]
CMD [""]

