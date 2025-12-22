#!/bin/bash
# Auto-deploy script for RentalCore
# Runs after successful git push and docker build

set -e

echo "🚀 Auto-deploying RentalCore to docker03..."

# Pull latest image on docker03
echo "📦 Pulling latest image..."
ssh noah@docker03 "cd /opt/docker/komodo/stacks/tscores && docker compose pull rentalcore"

# Restart container
echo "🔄 Restarting container..."
ssh noah@docker03 "docker restart rentalcore"

echo "✅ Deployment complete!"
echo "🌐 RentalCore should be live at https://rent.tsunami-events.de"
