#!/bin/bash

curl -X DELETE -v localhost:8000/inventories/3 \
  -H 'Content-Type: application/json; charset=utf-8'
