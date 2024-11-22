#!/bin/bash

# Start Web Service A
go run servA/main.go &
service_a_pid=$!

# Start Web Service B
go run servB/main.go &
service_b_pid=$!

# Wait for user to terminate both services
trap "kill $service_a_pid $service_b_pid" SIGINT
wait
