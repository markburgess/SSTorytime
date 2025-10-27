#!/bin/bash
# Docker startup script for SSTorytime

set -e

echo "🔥 Starting SSTorytime Docker..."
echo ""

# Build with BuildKit
echo "📦 Building Docker images..."
DOCKER_BUILDKIT=1 docker compose build

echo ""
echo "🚀 Starting services..."
docker compose up -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 3

echo ""
echo "✅ Services started!"
echo ""
echo "📊 PostgreSQL: localhost:5432"
echo "🌐 Web UI:     http://localhost:8080"
echo ""
echo "📚 To load example data:"
echo "  docker compose --profile tools up -d cli"
echo "  docker compose exec cli ./N4L examples/tutorial.n4l"
echo ""
echo "📋 To view logs:"
echo "  docker compose logs -f server"
echo ""
echo "🛑 To stop:"
echo "  docker compose down"
echo ""
