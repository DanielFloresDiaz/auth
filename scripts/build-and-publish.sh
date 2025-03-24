#!/bin/bash

set -e  # Exit on any error
set -x  # Enable verbose output

# Default values
TAG="latest"
GITHUB_USER="danielfloresdiaz"
BUILD_ARGS=""
NO_CACHE=""

function print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -t, --tag TAG        Image tag (default: latest)"
    echo "  -b, --build-arg ARG  Build argument (can be used multiple times)"
    echo "  --no-cache           Disable Docker cache during build"
    echo "  -h, --help           Show this help message"
}

# Parse all command line arguments in a single pass
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift
            ;;
        -b|--build-arg)
            BUILD_ARGS="$BUILD_ARGS --build-arg \"$2\""
            shift
            ;;
        --no-cache)
            NO_CACHE="--no-cache"
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo "Unknown parameter: $1"
            print_usage
            exit 1
            ;;
    esac
    shift
done

read -p "You are about to build and publish the main Auth image. Are you sure? (y/N): " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "Build cancelled."
        exit 0
    fi

# Set fixed image details for auth
IMAGE_NAME="ghcr.io/$GITHUB_USER/auth"
DOCKERFILE="../dockerfiles/Dockerfile"

# Uncomment the following lines to enable GitHub Container Registry login
# echo "🔑 Logging in to GitHub Container Registry..."
# echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USER --password-stdin

echo "🏗️ Building auth Docker image using BuildKit..."
DOCKER_BUILDKIT=1 docker build \
  --ssh default \
    $BUILD_ARGS \
    $NO_CACHE \
  -t $IMAGE_NAME:$TAG \
  -f $DOCKERFILE \
  ..

echo "🚀 Pushing image to GitHub Container Registry..."
docker push $IMAGE_NAME:$TAG

echo "✅ Image successfully built and published to $IMAGE_NAME:$TAG"