# Redis Hash Field Configuration Example

This example demonstrates how to use the configuration library with Redis hash fields as the data source.

## Features

- ✅ **Hash Field Support** - Read configuration from specific Redis hash fields
- ✅ **Real-time Updates** - Subscribe to changes using Redis keyspace notifications
- ✅ **JSON Configuration** - Store structured configuration data in hash fields
- ✅ **Multiple Fields** - Organize different configuration sections in separate hash fields
- ✅ **Type Safety** - Automatic unmarshaling into Go structs

## Prerequisites

1. Redis server running on localhost:6379
2. Go environment with required dependencies

## Quick Start

1. **Setup the configuration:**
   ```bash
   ./setup.sh
   ```

2. **Run the example:**
   ```bash
   go run main.go
   ```

3. **Test configuration updates:**
   ```bash
   # In another terminal while main.go is running
   redis-cli HSET app-config database '{"host":"production.db","port":5432,"database":"prod_db","username":"prod_user"}'
   ```

## How It Works

The example uses Redis hash fields to organize configuration:

1. **Hash Key**: `app-config` - Contains multiple configuration sections
2. **Hash Fields**:
   - `database` - Database connection settings
   - `cache` - Redis cache configuration  
   - `logging` - Application logging settings

3. **URI Format**: `redis://localhost:6379/app-config?content-type=application/json#database`
   - Base: `redis://localhost:6379/`
   - Hash Key: `app-config`
   - Content Type: `application/json` (required for format detection)
   - Hash Field: `database` (specified with `#`)

## Configuration Structure

### Database Field (`app-config#database`)
```json
{
  "host": "localhost",
  "port": 5432,
  "database": "myapp_dev",
  "username": "dev_user"
}
```

### Cache Field (`app-config#cache`)  
```json
{
  "host": "localhost",
  "port": 6379,
  "database": 0
}
```

### Logging Field (`app-config#logging`)
```json
{
  "level": "debug",
  "file": "/tmp/app.log"
}
```

## Hash Field Benefits

- **Organization**: Group related configuration together
- **Atomicity**: Update individual sections without affecting others
- **Efficiency**: Only subscribe to changes for specific fields
- **Flexibility**: Different applications can use different fields from the same hash

## Test Configuration Updates

### Method 1: Update database configuration
```bash
redis-cli HSET app-config database '{"host":"production.db","port":5432,"database":"prod_db","username":"prod_user"}'
```

### Method 2: View all hash fields
```bash
redis-cli HGETALL app-config
```

### Method 3: Get specific field
```bash
redis-cli HGET app-config database
```

## Expected Output

```
🔧 Redis Hash Field Configuration Example
=========================================
📍 Reading from: redis://localhost:6379/app-config#database
   - Hash key: 'app-config'
   - Hash field: 'database'

📥 Loading initial configuration...
✅ Configuration loaded successfully!
🗄️  Database Configuration:
   📍 Host: localhost
   🔌 Port: 5432
   💾 Database: myapp_dev
   👤 Username: dev_user

🔄 Subscribing to configuration changes...
💡 To test changes, run:
   redis-cli HSET app-config database '{"host":"production.db","port":5432,"database":"prod_db","username":"prod_user"}'

🔥 Configuration Update #1 Detected!
🗄️  Database Configuration:
   📍 Host: production.db
   🔌 Port: 5432
   💾 Database: prod_db
   👤 Username: prod_user
```

## Use Cases

- **Multi-tenant applications**: Each tenant has their own hash field
- **Environment-specific configs**: dev, staging, prod fields in the same hash
- **Service configuration**: Each microservice uses its own hash field
- **Feature toggles**: Organize feature flags by category in hash fields

## Troubleshooting

1. **"hash field not found" error:**
   - Ensure Redis is running: `redis-cli ping`
   - Check if hash field exists: `redis-cli HGET app-config database`
   - Run setup script: `./setup.sh`

2. **No configuration updates detected:**
   - Ensure keyspace notifications are enabled in Redis
   - Check Redis config: `redis-cli CONFIG GET notify-keyspace-events`
   - The library automatically enables required notifications

3. **Connection refused error:**
   - Start Redis server: `sudo systemctl start redis`
   - Or run Redis in foreground: `redis-server`