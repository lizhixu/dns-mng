#!/bin/bash

# DNS Manager - Stop Script

set -e

echo "🛑 Stopping DNS Manager..."

docker-compose stop

echo "✅ Services stopped successfully!"
echo ""
echo "To start again: ./start.sh"
echo "To remove all containers: docker-compose down"
echo "To remove all data: docker-compose down -v"
