FROM golang:1.25-bookworm@sha256:c4bc0741e3c79c0e2d47ca2505a06f5f2a44682ada94e1dba251a3854e60c2bd AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY gen-code-insights/go.mod gen-code-insights/go.sum ./
RUN go mod download

# Copy the rest of the Go source code
COPY gen-code-insights/ ./

# Build the Go application
RUN CGO_ENABLED=0 go build -o /gen-code-insights .

FROM ghcr.io/safedep/vet:v1.12.18

RUN apk --no-cache add curl

COPY --from=builder /gen-code-insights /

COPY pipe /
COPY LICENSE.txt pipe.yml README.md /

RUN chmod a+x /*.sh

ENTRYPOINT ["/pipe.sh"]
