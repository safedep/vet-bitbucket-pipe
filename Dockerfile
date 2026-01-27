FROM golang:1.25.5 AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY gen-code-insights/go.mod gen-code-insights/go.sum ./
RUN go mod download

# Copy the rest of the Go source code
COPY gen-code-insights/ ./

# Build the Go application
RUN CGO_ENABLED=0 go build -o /gen-code-insights .

FROM ghcr.io/safedep/vet:v1.12.18

RUN apt update -y
RUN apt install curl -y

COPY --from=builder /gen-code-insights /

COPY pipe /
COPY LICENSE.txt pipe.yml README.md /

RUN chmod a+x /*.sh

ENTRYPOINT ["/pipe.sh"]
