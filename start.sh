#!/bin/bash

# DNS Manager - Docker Deployment Script

set -e

echo "🚀 Starting DNS Manager..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env file and set your JWT_SECRET before running in production!"
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install it first."
    exit 1
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose up -d --build

# Wait for services to be healthy
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check service status
echo "📊 Service Status:"
docker-compose ps

echo ""
echo "✅ DNS Manager is running!"
echo ""
echo "📱 Access the application:"
echo "   Frontend: http://localhost"
echo "   Backend API: http://localhost:8080"
echo ""
echo "📝 Useful commands:"
echo "   View logs: docker-compose logs -f"
echo "   Stop services: docker-compose stop"
echo "   Restart services: docker-compose restart"
echo "   Remove services: docker-compose down"
echo ""
