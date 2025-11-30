#!/bin/sh

# ambros-docker-utils plugin

# Function for the cleanup command
cleanup() {
    echo "This will remove all stopped containers, dangling images, and unused volumes and networks."
    printf "Are you sure you want to continue? [y/N] "
    read -r response
    if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
        echo "Cleanup aborted."
        exit 0
    fi

    echo "Removing stopped containers..."
    docker ps -a -q -f status=exited | xargs -r docker rm

    echo "Removing dangling images..."
    docker images -f "dangling=true" -q | xargs -r docker rmi

    echo "Removing unused volumes..."
    docker volume ls -qf dangling=true | xargs -r docker volume rm

    echo "Removing unused networks..."
    docker network ls -q --filter "driver=bridge" --format "{{.ID}}" | xargs -r docker network rm

    echo "Docker cleanup complete."
}

# Function for the stats command
stats() {
    echo "Showing Docker container stats (press Ctrl+C to exit)..."
    docker stats
}

# Main command dispatcher
case "$1" in
    "cleanup")
        cleanup
        ;;
    "stats")
        stats
        ;;
    *)
        echo "Usage: ambros ambros-docker-utils {cleanup|stats}"
        exit 1
        ;;
esac
