#!/bin/bash
# Quick start script for SSTorytime Docker

set -e

echo "🔥 Starting SSTorytime..."
echo ""

# Start services
docker-compose up -d

echo ""
echo "✅ Services started!"
echo ""
echo "📊 PostgreSQL: localhost:5432"
echo "🌐 Web UI:     http://localhost:8080"
echo ""
echo "To load example data:"
echo "  docker-compose --profile tools up -d cli"
echo "  docker-compose exec cli ./N4L examples/tutorial.n4l"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f server"
echo ""
echo "To stop:"
echo "  docker-compose down"
echo ""
