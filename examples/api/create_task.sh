curl -X POST \
    http://localhost:9808/api/v1/apps/68b01029966f394ebdebfde8/tasks \
    -H "Content-Type: application/json" \
    -d '{
        "description": "Create an app for me to manage my timesheet. I should be able to select a week and enter my hours for a specific customer"
    }'
