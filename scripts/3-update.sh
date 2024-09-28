#!/bin/bash

curl -X PATCH -v localhost:8000/inventories/1 \
  -H 'Content-Type: application/json; charset=utf-8' \
  --data-binary @- << EOF
{
    "name": "Car racing TG4599 Turbo MAX Toys",
    "description": "Car racing turbo for children +12", 
    "fullPriceTHB": 20000,
    "count": 9,
    "manufacturerId": 3333
}
EOF
