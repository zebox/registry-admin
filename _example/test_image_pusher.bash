#!/bin/bash
for i in {2..200}
do
   echo "pushed $i times"
   docker image tag hello-world:latest 127.0.0.1:50554/test_1/hello-world:0.$i.0
   docker image push 127.0.0.1:50554/test_1/hello-world:0.$i.0
done
