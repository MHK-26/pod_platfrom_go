# deployments/docker/db-init/Dockerfile
FROM golang:1.20-alpine

# Install PostgreSQL client and migration tool
RUN apk add --no-cache postgresql-client
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy migration scripts
COPY scripts/migrations /app/scripts/migrations
COPY scripts/seed /app/scripts/seed

# Copy initialization script
COPY deployments/docker/db-init/init.sh /app/init.sh
RUN chmod +x /app/init.sh

# Run the script
CMD ["/app/init.sh"]