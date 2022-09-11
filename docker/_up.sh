#!/bin/bash
docker-compose -f docker-compose.yml -f docker/docker-compose.yml up --build

# Initialize a local docker instance