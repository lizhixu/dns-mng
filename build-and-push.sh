#!/bin/bash

# DNS Manager - Build and Push to Docker Hub
# Usage: ./build-and-push.sh [version]

set -e

VERSION=${1:-latest}
DOCKER_USERNAME="jacyli"
IMAGE_NAME="dns-mng"

echo "🔨 Building DNS Manager Docker images..."
echo "📦 Version: $VERSION"
echo ""

# Build backend image
echo "🔨 Building backend image..."
docker build -t ${DOCKER_USERNAME}/${IMAGE_NAME}:backend-${VERSION} ./backend
docker tag ${DOCKER_USERNAME}/${IMAGE_NAME}:backend-${VERSION} ${DOCKER_USERNAME}/${IMAGE_NAME}:backend

# Build frontend image
echo "🔨 Building frontend image..."
docker build -t ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend-${VERSION} ./frontend
docker tag ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend-${VERSION} ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend

echo ""
echo "✅ Build completed!"
echo ""
echo "📤 Pushing images to Docker Hub..."
echo ""

# Login to Docker Hub (if not already logged in)
echo "🔐 Please login to Docker Hub if prompted..."
docker login

# Push backend images
echo "📤 Pushing backend image..."
docker push ${DOCKER_USERNAME}/${IMAGE_NAME}:backend-${VERSION}
docker push ${DOCKER_USERNAME}/${IMAGE_NAME}:backend

# Push frontend images
echo "📤 Pushing frontend image..."
docker push ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend-${VERSION}
docker push ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend

echo ""
echo "✅ All images pushed successfully!"
echo ""
echo "📋 Images:"
echo "   Backend:  ${DOCKER_USERNAME}/${IMAGE_NAME}:backend-${VERSION}"
echo "   Backend:  ${DOCKER_USERNAME}/${IMAGE_NAME}:backend (latest)"
echo "   Frontend: ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend-${VERSION}"
echo "   Frontend: ${DOCKER_USERNAME}/${IMAGE_NAME}:frontend (latest)"
echo ""
echo "🚀 To use these images:"
echo "   docker-compose pull"
echo "   docker-compose up -d"
echo ""
