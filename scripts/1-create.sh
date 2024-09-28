#!/bin/bash

curl -X POST -v localhost:8000/inventories \
  -H 'Content-Type: application/json; charset=utf-8' \
  --data-binary @- << EOF
{
    "name": "Meiji Power Milk 500ML",
    "description": "A very good choice for those who is lactose tolerance",
    "fullPriceTHB": 4500, 
    "count": 43,
    "manufacturerId": 463453
}
EOF
