# Deployment Guide

This guide covers different deployment strategies for the payloadBuddy, specifically tailored for ServiceNow consultants and developers who need to test REST integrations in various environments.

## Overview

The payloadBuddy can be deployed in multiple ways depending on your testing scenario and security requirements:

| Solution | Use Case | Complexity | Security | Best For |
|----------|----------|------------|----------|----------|
| **ngrok** | Quick testing, demos, external access | Low | Medium | Development, POCs, training |
| **Docker + MID-Server** | Production-like testing, secure connections | High | High | Enterprise testing, staging environments |

---

## Solution 1: ngrok Deployment

### When to Use
- **Rapid Prototyping**: Quick setup for testing ServiceNow REST integrations
- **Remote Demos**: Allow ServiceNow instances to reach your local server
- **Training Sessions**: Easy setup for workshop environments
- **External Testing**: When ServiceNow needs to reach endpoints outside the corporate network

### Prerequisites
- Go 1.24.5 or newer
- [ngrok account](https://ngrok.com/) (free tier available)

### Step-by-Step Setup

#### 1. Install and Configure ngrok
```bash
# Download and install ngrok
# Visit https://ngrok.com/download for your platform

# Authenticate (after creating account)
ngrok authtoken YOUR_AUTH_TOKEN
```

#### 2. Build and Start the Server
```bash
# Clone and build the application
git clone https://github.com/dtrabandt/payloadBuddy.git
cd payloadBuddy
go build -o payloadBuddy

# Start with authentication (recommended)
./payloadBuddy -auth
```

#### 3. Expose via ngrok
```bash
# In a new terminal, expose port 8080
ngrok http 8080

# For custom domain (paid plans)
ngrok http 8080 --domain=your-custom-domain.ngrok.io
```

#### 4. Configure ServiceNow
```javascript
// In ServiceNow REST Message or Flow Action
// Use the ngrok URL from terminal output

// Example REST Message Configuration
var r = new sn_ws.RESTMessageV2();
r.setEndpoint('https://abc123.ngrok.io/huge_payload');
r.setHttpMethod('GET');

// Add authentication if enabled
r.setBasicAuth('username', 'password');

var response = r.execute();
gs.info('Response: ' + response.getBody());
```

### ngrok Configuration Options

#### Basic Configuration
```yaml
# ~/.ngrok2/ngrok.yml
version: "2"
authtoken: YOUR_AUTH_TOKEN

tunnels:
  payloadBuddy:
    addr: 8080
    proto: http
    bind_tls: true
    auth: "user:password"  # Basic auth at ngrok level
```

#### Advanced Configuration
```yaml
# ~/.ngrok2/ngrok.yml
version: "2"
authtoken: YOUR_AUTH_TOKEN

tunnels:
  payloadBuddy-secure:
    addr: 8080
    proto: http
    bind_tls: true
    oauth:
      provider: google
      allow_emails:
        - consultant@company.com
    request_headers:
      add:
        - "X-ServiceNow-Testing: true"
```

#### Start with Configuration
```bash
# Start specific tunnel
ngrok start payloadBuddy

# Start all tunnels
ngrok start --all
```

### ServiceNow Integration Examples

#### Flow Action Configuration
```javascript
// Flow Action Script
(function execute(inputs, outputs) {
    var restMessage = new sn_ws.RESTMessageV2();
    restMessage.setEndpoint(inputs.ngrok_url + '/stream_payload');
    restMessage.setHttpMethod('GET');
    restMessage.setQueryParameter('count', inputs.record_count || '1000');
    restMessage.setQueryParameter('scenario', 'peak_hours');
    restMessage.setQueryParameter('servicenow', 'true');
    
    // Add authentication
    restMessage.setBasicAuth(inputs.username, inputs.password);
    
    var response = restMessage.execute();
    
    outputs.response_code = response.getStatusCode();
    outputs.response_body = response.getBody();
    outputs.success = response.getStatusCode() == 200;
})(inputs, outputs);
```

#### Scheduled Job Testing
```javascript
// Scheduled Script Execution
var gr = new GlideRecord('sys_script_execution');
gr.addQuery('state', 'ready');
gr.query();

while (gr.next()) {
    var restClient = new sn_ws.RESTMessageV2();
    restClient.setEndpoint('https://your-ngrok-url.ngrok.io/huge_payload');
    restClient.setHttpMethod('GET');
    restClient.setQueryParameter('count', '5000');
    
    var response = restClient.execute();
    
    if (response.getStatusCode() == 200) {
        var data = JSON.parse(response.getBody());
        gs.info('Processed ' + data.length + ' records');
    }
}
```

---

## Solution 2: Docker + MID-Server Deployment

### When to Use
- **Enterprise Testing**: Simulate production-like environments
- **Security Requirements**: Keep traffic within corporate network boundaries
- **MID-Server Integration**: Test scenarios involving ServiceNow MID-Server
- **Scalability Testing**: Test with multiple containers and load balancing
- **CI/CD Pipelines**: Automated testing environments

### Architecture Overview
```
                    ServiceNow Cloud
          ┌─────────────────────────────────────┐
          │         ┌─────────────────┐         │
          │         │   ServiceNow    │         │
          │         │   Instance/PDI  │         │
          │         └─────────────────┘         │
          └─────────────────┬───────────────────┘
                            ▲ HTTPS (Outbound only)
                            │ MID-Server connects TO instance
                            │
    ┌───────────────────────────────────────────────┐
    │                 Docker Network                │
    │                                               │
    │  ┌──────────────────┐    ┌──────────────────┐ │
    │  │   MID-Server     │◄──►│   payloadBuddy   | │
    │  │   Container      │    │    Container     │ │
    │  └──────────────────┘    └──────────────────┘ │
    │                                               │
    └───────────────────────────────────────────────┘
              Local Network/Docker Host
```

### Prerequisites
- Docker and Docker Compose
- ServiceNow MID-Server installation files
- Basic understanding of Docker networking
- ServiceNow instance (cloud-based PDI or production instance)

### Connection Flow
1. **ServiceNow Instance**: Lives in ServiceNow cloud (not in Docker network)
2. **MID-Server Container**: Initiates HTTPS connection TO ServiceNow instance
3. **Payload Server Container**: Accessible by MID-Server within Docker network
4. **Data Flow**: ServiceNow → MID-Server → Payload Server (for testing scenarios)

### Step-by-Step Setup

#### 1. Create Project Structure
```bash
mkdir servicenow-testing-environment
cd servicenow-testing-environment

mkdir -p {midserver,payloadserver,config,logs}
```

#### 2. Create Dockerfile for Payload Server
```dockerfile
# payloadserver/Dockerfile
FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o payloadBuddy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/payloadBuddy .

EXPOSE 8080

# Default to authenticated mode in container
CMD ["./payloadBuddy", "-auth"]
```

#### 3. Create MID-Server Dockerfile
```dockerfile
# midserver/Dockerfile
FROM openjdk:11-jre-slim

# Install required packages
RUN apt-get update && apt-get install -y \
    wget \
    unzip \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create midserver user
RUN useradd -m -s /bin/bash midserver

# Set working directory
WORKDIR /opt/servicenow/midserver

# Copy MID-Server files (you need to download these from ServiceNow)
COPY midserver-files/ .

# Set ownership
RUN chown -R midserver:midserver /opt/servicenow/midserver

USER midserver

# Expose MID-Server port (if needed)
EXPOSE 8080

CMD ["./start.sh"]
```

#### 4. Create Docker Compose Configuration
```yaml
# docker-compose.yml
version: '3.8'

services:
  payloadserver:
    build: 
      context: .
      dockerfile: payloadserver/Dockerfile
    container_name: payloadBuddy
    ports:
      - "8080:8080"
    environment:
      - AUTH_ENABLED=true
      - DEFAULT_USER=testuser
      - DEFAULT_PASS=testpass123
    networks:
      - servicenow-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/huge_payload?count=1"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    volumes:
      - ./logs:/var/log/payloadserver

  midserver:
    build:
      context: .
      dockerfile: midserver/Dockerfile
    container_name: servicenow-midserver
    environment:
      - MID_INSTANCE_URL=https://your-instance.servicenowservices.com
      - MID_USERNAME=mid.server.user
      - MID_PASSWORD=your_password
      - MID_NAME=docker-midserver
    networks:
      - servicenow-network
    depends_on:
      payloadserver:
        condition: service_healthy
    volumes:
      - ./config/midserver:/opt/servicenow/midserver/config
      - ./logs:/opt/servicenow/midserver/logs
    restart: unless-stopped

  # Optional: nginx reverse proxy for load balancing
  nginx:
    image: nginx:alpine
    container_name: nginx-proxy
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./config/ssl:/etc/nginx/ssl:ro
    networks:
      - servicenow-network
    depends_on:
      - payloadserver
    restart: unless-stopped

networks:
  servicenow-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

#### 5. Configure MID-Server
```properties
# config/midserver/config.xml
<?xml version="1.0" encoding="UTF-8"?>
<config>
    <parameter name="url" value="https://your-instance.servicenowservices.com" />
    <parameter name="mid.instance.username" value="mid.server.user" />
    <parameter name="mid.instance.password" value="your_encrypted_password" />
    <parameter name="name" value="docker-midserver" />
    
    <!-- Custom parameters for payload server testing -->
    <parameter name="ext.payload.server.url" value="http://payloadserver:8080" />
    <parameter name="ext.payload.server.auth" value="true" />
</config>
```

#### 6. Create nginx Configuration (Optional)
```nginx
# config/nginx.conf
events {
    worker_connections 1024;
}

http {
    upstream payloadserver {
        server payloadserver:8080;
        # Add more servers for load balancing
        # server payloadserver2:8080;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://payloadserver;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # For streaming endpoints
            proxy_buffering off;
            proxy_cache off;
        }

        # Health check endpoint
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
```

#### 7. Deploy the Environment
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f payloadserver
docker-compose logs -f midserver

# Check service status
docker-compose ps

# Scale payload servers (if needed)
docker-compose up -d --scale payloadserver=3
```

### ServiceNow MID-Server Integration

#### Custom MID-Server Script
```javascript
// In ServiceNow - Script Include for MID-Server communication
var PayloadServerUtil = Class.create();
PayloadServerUtil.prototype = {
    initialize: function() {
        this.midServer = 'docker-midserver'; // MID-Server name
        this.baseUrl = 'http://payloadserver:8080';
    },

    testHugePayload: function(count) {
        var probe = new MIDRequest();
        probe.setName('JavascriptRequest');
        probe.setMIDServer(this.midServer);
        
        var script = 'var url = "' + this.baseUrl + '/huge_payload?count=' + count + '";';
        script += 'var request = new XMLHttpRequest();';
        script += 'request.open("GET", url, false);';
        script += 'request.setRequestHeader("Authorization", "Basic " + btoa("testuser:testpass123"));';
        script += 'request.send();';
        script += 'response = { status: request.status, body: request.responseText };';
        
        probe.setScript(script);
        probe.post();
        
        return probe.getResponseRecord();
    },

    testStreamingPayload: function(scenario) {
        var probe = new MIDRequest();
        probe.setName('JavascriptRequest');
        probe.setMIDServer(this.midServer);
        
        var script = 'var url = "' + this.baseUrl + '/stream_payload?scenario=' + scenario + '&servicenow=true";';
        script += 'var request = new XMLHttpRequest();';
        script += 'request.open("GET", url, false);';
        script += 'request.setRequestHeader("Authorization", "Basic " + btoa("testuser:testpass123"));';
        script += 'request.send();';
        script += 'response = { status: request.status, body: request.responseText };';
        
        probe.setScript(script);
        probe.post();
        
        return probe.getResponseRecord();
    },

    type: 'PayloadServerUtil'
};
```

#### Flow Action for MID-Server Testing
```javascript
// Flow Action: Test Payload Server via MID-Server
(function execute(inputs, outputs) {
    var util = new PayloadServerUtil();
    
    try {
        var response = util.testHugePayload(inputs.record_count || 1000);
        
        outputs.success = response.getValue('status') == '200';
        outputs.response_body = response.getValue('response');
        outputs.mid_server = util.midServer;
        
        if (outputs.success) {
            var data = JSON.parse(outputs.response_body);
            outputs.record_count = data.length;
            gs.info('Successfully processed ' + data.length + ' records via MID-Server');
        }
    } catch (e) {
        outputs.success = false;
        outputs.error_message = e.message;
        gs.error('MID-Server payload test failed: ' + e.message);
    }
})(inputs, outputs);
```

### Container Management

#### Monitoring and Debugging
```bash
# Monitor resource usage
docker stats

# Check container health
docker-compose exec payloadserver wget -qO- http://localhost:8080/huge_payload?count=1

# View real-time logs
docker-compose logs -f --tail=100 payloadserver

# Execute commands in container
docker-compose exec payloadserver /bin/sh

# Restart specific service
docker-compose restart payloadserver
```

#### Scaling and Load Testing
```bash
# Scale payload servers
docker-compose up -d --scale payloadserver=5

# Test load balancing
for i in {1..10}; do
    curl -u testuser:testpass123 http://localhost/huge_payload?count=100
done

# Monitor nginx logs
docker-compose logs -f nginx
```

---

## Security Considerations

### ngrok Security
- **Tunnel URLs**: Use HTTPS tunnels (default with ngrok)
- **Authentication**: Enable application-level basic auth
- **Access Control**: Use ngrok's built-in authentication (paid plans)
- **Tunnel Monitoring**: Monitor ngrok dashboard for unauthorized access
- **Rate Limiting**: Implement rate limiting in your application

```bash
# Secure ngrok tunnel
ngrok http 8080 --basic-auth="user:strongpassword" --region=us
```

### Docker Security
- **Network Isolation**: Use custom Docker networks
- **Secrets Management**: Use Docker secrets for sensitive data
- **Container Security**: Run containers as non-root users
- **Image Security**: Scan images for vulnerabilities
- **Access Logs**: Monitor container access logs

```yaml
# docker-compose.yml security enhancements
services:
  payloadserver:
    # ... other config
    security_opt:
      - no-new-privileges:true
    user: "1000:1000"
    read_only: true
    tmpfs:
      - /tmp:rw,noexec,nosuid,size=100m
```

### ServiceNow Security
- **MID-Server Certificates**: Use proper SSL certificates
- **User Permissions**: Create dedicated service accounts
- **Network Segmentation**: Isolate test environments
- **Audit Logging**: Enable detailed audit logs

---

## Troubleshooting

### Common ngrok Issues

#### Tunnel Connection Failed
```bash
# Check ngrok status
curl -s http://localhost:4040/api/tunnels | jq

# Restart ngrok with debug
ngrok http 8080 --log=debug
```

#### ServiceNow Can't Reach Endpoint
```javascript
// Test connectivity from ServiceNow
var r = new sn_ws.RESTMessageV2();
r.setEndpoint('https://your-ngrok-url.ngrok.io/huge_payload?count=1');
r.setHttpTimeout(30000);
var response = r.execute();
gs.info('Status: ' + response.getStatusCode());
gs.info('Headers: ' + response.getAllHeaders());
```

### Common Docker Issues

#### Container Won't Start
```bash
# Check container logs
docker-compose logs payloadserver

# Check resource usage
docker system df
docker system prune

# Rebuild container
docker-compose build --no-cache payloadserver
```

#### MID-Server Connection Issues
```bash
# Check MID-Server logs
docker-compose logs midserver | grep -i error

# Test network connectivity
docker-compose exec midserver ping payloadserver
docker-compose exec midserver curl http://payloadserver:8080/huge_payload?count=1
```

#### Performance Issues
```bash
# Monitor resource usage
docker stats --no-stream

# Check container limits
docker inspect payloadBuddy | grep -i memory

# Scale horizontally
docker-compose up -d --scale payloadserver=3
```

---

## Best Practices

### Development Workflow
1. **Local Testing**: Start with local development
2. **ngrok Integration**: Use ngrok for ServiceNow integration testing
3. **Docker Staging**: Deploy to Docker for production-like testing
4. **Production Deployment**: Use proper container orchestration

### Performance Optimization
- **Container Resources**: Set appropriate CPU/memory limits
- **Caching**: Use Redis for caching large payloads
- **Load Balancing**: Distribute load across multiple containers
- **Monitoring**: Implement proper monitoring and alerting

### Maintenance
- **Regular Updates**: Keep base images updated
- **Log Rotation**: Implement log rotation for long-running containers
- **Backup**: Backup configuration and data volumes
- **Health Checks**: Implement comprehensive health checks

---

## Advanced Configurations

### Multi-Environment Setup
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  payloadserver:
    environment:
      - ENV=development
      - DEBUG=true
      - LOG_LEVEL=debug

# docker-compose.prod.yml
version: '3.8'
services:
  payloadserver:
    environment:
      - ENV=production
      - DEBUG=false
      - LOG_LEVEL=info
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
```

### Kubernetes Deployment
```yaml
# k8s/deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: payloadBuddy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: payloadserver
  template:
    metadata:
      labels:
        app: payloadserver
    spec:
      containers:
      - name: payloadserver
        image: payloadBuddy:latest
        ports:
        - containerPort: 8080
        env:
        - name: AUTH_ENABLED
          value: "true"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

This deployment guide provides comprehensive options for different testing scenarios and environments, ensuring ServiceNow consultants can choose the most appropriate deployment strategy for their specific needs.