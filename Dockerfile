# --- Build Stage ---
# Use a specific Go version for reproducible builds.
# Using a slim-bullseye base can be faster than Alpine due to glibc vs musl differences.
FROM golang:1.24-bullseye AS buildstage

# Set the working directory
WORKDIR /app

# 1. Copy only the files needed to download dependencies.
# This layer is only invalidated if go.mod or go.sum changes.
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy the rest of your application source code.
# This layer is invalidated if any .go file changes, but dependencies remain cached.
COPY . .

# 3. Build the application.
# This layer is only invalidated if source code changes, not on documentation/config changes.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -ldflags "-w -s -X 'main.version=${VERSION}'" -o /kmagent ./cmd/kmagent/...
    # Using ./cmd/kmagent/... is more robust than specifying main.go
    # The -a flag forces a rebuild of all packages, which can prevent stale cache issues.
    # The -w -s flags strip debug info, making the binary smaller.

# --- Final Stage ---
# Use a distroless image for a minimal and secure final container.
# It contains only your app and its runtime dependencies, nothing else.
FROM gcr.io/distroless/static-debian11

# Copy the static binary from the build stage.
COPY --from=buildstage /kmagent /kmagent

# Copy the configuration file.
COPY ./configs/docker-col-config.yaml /config.yaml


# Define the entrypoint for the container.
ENTRYPOINT ["/kmagent", "--docker-mode", "--config", "/config.yaml", "start"]