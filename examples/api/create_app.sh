curl -X POST \
    http://localhost:9808/api/v1/apps \
    -H "Content-Type: application/json" \
    -d '{
        "name": "E-commerce Dashboard",
        "description": "Modern dashboard for online store management"
    }'