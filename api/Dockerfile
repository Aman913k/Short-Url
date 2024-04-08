# Stage 1: Build the application
FROM golang:alpine as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

# Install dependencies (if any)
# RUN go mod download

# Build the application
RUN go build -o main .

# Stage 2: Create a minimal image to run the application
FROM alpine

# Create a non-root user
RUN adduser -S -D -H -h /app appuser
USER appuser

# Set the working directory in the container
WORKDIR /app

# Copy the binary built in the previous stage into the container
COPY --from=builder /build/main .

# Copy the .env file into the container
COPY .env .

# Expose the port on which the application will run
EXPOSE 3000

# Command to run the application
CMD ["./main"]   
