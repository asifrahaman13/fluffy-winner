#!/bin/bash

# Function to kill the processes when the script exits
cleanup() {
  echo "Stopping Go and Next.js applications..."
  kill $(jobs -p)
}
trap cleanup EXIT

# Navigate to the directory containing your applications
cd "$(dirname "$0")"

# Start the Go application
echo "Starting Go application..."
go run main.go &

# Start the Next.js application
echo "Starting Next.js application..."
cd frontend/
bun run dev &

# Wait for all background processes to finish
wait
