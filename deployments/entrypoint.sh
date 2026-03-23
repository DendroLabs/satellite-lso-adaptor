#!/bin/sh
set -e

# Start the Python orbital engine in the background
python3 /app/python/server.py &
ORBITAL_PID=$!

# Wait briefly for the orbital service to start
sleep 1

# Start the Go LSO adaptor
exec /usr/local/bin/adaptor
