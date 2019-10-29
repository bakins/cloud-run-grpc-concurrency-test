#!/bin/bash
set -eux

NAME=grpc-concurrency-test
PROJECT=rsg-login-sandbox
VERSION=$(git rev-parse --short HEAD)

TAG=us.gcr.io/$PROJECT/$NAME:$VERSION
docker build -t $TAG .
docker push $TAG

gcloud beta run deploy "$NAME" \
    --project="$PROJECT" \
    --image="$TAG" \
    --region=us-east1 \
    --allow-unauthenticated \
    --platform=managed \
    --concurrency=1 \
    --timeout=30s

