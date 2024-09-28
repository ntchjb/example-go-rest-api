#!/bin/bash

curl -X GET -v "http://localhost:8000/inventories?limit=2&cursor=0" \
  -H 'Content-Type: application/json; charset=utf-8'
