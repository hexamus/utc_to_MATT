# Start from the latest golang base image
FROM golang:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY ./utcar/ .

#test
RUN go mod init utcar
RUN go get -d -v

# Build the Go app
RUN go build -o main .

EXPOSE 12300

# Command to run the executable
CMD ["./main"]
