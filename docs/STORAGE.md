# Storage Backends

The OpenID Connect Server supports multiple storage backends for flexibility and scalability.

## Available Storage Options

### 1. JSON File Storage (Default)

Simple file-based storage using JSON format. Ideal for development, testing, and small deployments.

**Pros:**
- No external dependencies
- Easy to backup (single file)
- Human-readable format
- Zero setup required

**Cons:**
- Not suitable for high-concurrency environments
- Limited scalability
- No built-in clustering support

**Configuration:**
```toml
[storage]
type = "json"
json_file_path = "data.json"
```

**Command Line Override:**
```bash
./openid-server --json-store
```

### 2. MongoDB Storage

Production-grade NoSQL database storage. Recommended for production deployments.

**Pros:**
- Excellent scalability
- High performance
- Built-in replication and sharding
- Supports clustering
- ACID transactions

**Cons:**
- Requires MongoDB server
- Additional infrastructure to manage

**Configuration:**
```toml
[storage]
type = "mongodb"
mongo_uri = "mongodb://localhost:27017/openid"
```

**MongoDB Connection URI Examples:**

- **Local MongoDB:** `mongodb://localhost:27017/openid`
- **With Authentication:** `mongodb://username:password@localhost:27017/openid`
- **Replica Set:** `mongodb://host1:27017,host2:27017,host3:27017/openid?replicaSet=myReplicaSet`
- **MongoDB Atlas:** `mongodb+srv://username:password@cluster.mongodb.net/openid`

## Choosing a Storage Backend

### Use JSON Storage When:
- Running in development or testing
- Small number of users (< 100)
- Low traffic applications
- Single-server deployment
- Simple backup requirements

### Use MongoDB When:
- Running in production
- Need high availability
- Expecting high traffic
- Multiple server instances
- Compliance or audit requirements
- Advanced backup and recovery needs

## Data Migration

Currently, there is no built-in migration tool between storage backends. If you need to migrate:

1. **From JSON to MongoDB:**
   - Stop the server
   - Read the JSON file
   - Write a script to import data into MongoDB
   - Update configuration
   - Restart server

2. **From MongoDB to JSON:**
   - Export data from MongoDB
   - Convert to JSON format
   - Update configuration
   - Restart server

## Storage Schema

All storage backends implement the same interface and store the following entities:

### Users
- ID (UUID)
- Username (unique)
- Email (unique)
- Password Hash
- Created At / Updated At

### Clients
- ID (UUID)
- Name
- Secret
- Redirect URIs
- Created At

### Authorization Codes
- Code (unique)
- Client ID
- User ID
- Redirect URI
- Scope
- Expires At
- Created At

### Tokens
- ID (UUID)
- Access Token (unique)
- Refresh Token (unique)
- Client ID
- User ID
- Scope
- Expires At
- Created At

### Sessions
- ID (UUID)
- User ID
- Expires At
- Created At

## Performance Considerations

### JSON Storage
- All data is loaded into memory on startup
- Write operations persist to disk immediately
- Read operations are from in-memory cache
- Thread-safe with RWMutex

### MongoDB Storage
- Uses connection pooling
- Indexed on username, email, access_token, refresh_token
- Context-based timeouts (5 seconds per operation)
- Automatic cleanup of expired tokens via TTL indexes

## Backup and Recovery

### JSON Storage
Simply backup the `data.json` file regularly:
```bash
cp data.json data.json.backup
```

### MongoDB Storage
Use MongoDB's built-in backup tools:
```bash
mongodump --uri="mongodb://localhost:27017/openid" --out=/backup/
```

Restore:
```bash
mongorestore --uri="mongodb://localhost:27017/openid" /backup/openid/
```

## Security Considerations

- **JSON Storage:** Ensure proper file permissions (e.g., `chmod 600 data.json`)
- **MongoDB:** Use authentication, enable TLS, restrict network access
- **Both:** Passwords are always hashed with bcrypt
- **Both:** Never commit configuration files with credentials to version control
