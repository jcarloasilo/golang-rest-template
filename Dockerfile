# Use the official Golang image as the base image
FROM golang:1.22.3-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Install make
RUN apk add --no-cache make

# Build the application using make
RUN make build

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application using make
CMD ["/app/main"]