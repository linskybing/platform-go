#! /bin/bash

sudo apt install mockgen

for file in ./repositories/*_repository.go; do
    base=$(basename $file .go)
    mockgen -source=$file -destination=./repositories/mock_repositories/${base}_mock.go -package=mock_repositories
done

for file in ./services/*_service.go; do
    base=$(basename $file .go)
    mockgen -source=$file -destination=./services/mock_services/${base}_mock.go -package=mock_services
done