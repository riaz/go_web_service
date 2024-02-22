# go_web_service
This is a simple TDD go web service with CI/CD using github actions and adopted from - https://github.com/TomFern/go-mux-api/tree/master
which uses semaphore for CI which seems to not be free anymore.

[![Go](https://github.com/riaz/go_web_service/actions/workflows/go.yml/badge.svg)](https://github.com/riaz/go_web_service/actions/workflows/go.yml)

### Running the docker postgres image (without password)

    docker run -e POSTGRES_HOST_AUTH_METHOD=trust -it -p 5432:5432 -d postgres 

### Running the tests locally 
    
    go test -v

### Trouble shooting if the docker is running

    docker ps -a  # this gives the list of containers that ran but failed , copy the container id

    docker logs <container_id> # this allows you to see the logs

    # common steps to stop the container and remove if something is incorrect

    docker stop <container_id>

    docker rm <container_id>

