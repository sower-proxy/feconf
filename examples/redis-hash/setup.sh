#!/bin/bash

# Redis Hash Field Configuration Setup Script

echo "🔧 Setting up Redis hash field configuration example..."

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis is not running. Please start Redis server first:"
    echo "   sudo systemctl start redis"
    echo "   # or"
    echo "   redis-server"
    exit 1
fi

echo "✅ Redis is running"

# Create sample configuration in Redis hash
echo "📝 Creating sample configuration in Redis hash..."

redis-cli HSET app-config database '{
  "host": "localhost",
  "port": 5432,
  "database": "myapp_dev",
  "username": "dev_user"
}' > /dev/null

redis-cli HSET app-config cache '{
  "host": "localhost",
  "port": 6379,
  "database": 0
}' > /dev/null

redis-cli HSET app-config logging '{
  "level": "debug",
  "file": "/tmp/app.log"
}' > /dev/null

echo "✅ Redis hash 'app-config' created with multiple fields:"
echo "   - database: Database configuration"
echo "   - cache: Cache configuration"  
echo "   - logging: Logging configuration"

echo ""
echo "🔍 Current hash contents:"
redis-cli HGETALL app-config

echo ""
echo "🚀 Ready to run the example!"
echo "   go run main.go"
echo ""
echo "💡 To test configuration updates:"
echo '   redis-cli HSET app-config database '"'"'{"host":"production.db","port":5432,"database":"prod_db","username":"prod_user"}'"'"''