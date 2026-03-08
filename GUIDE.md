# Decoupled Canary Database Demo

This demo shows a decoupled **Canary Deployment** from Database using **Argo Rollouts** for stateful applications. The core focus is on managing database schema changes (adding `first_name` and `last_name` columns) while ensuring that multiple application versions can coexist without crashing.

## Overview

This project demonstrates how to safely migrate an application through multiple versions while evolving the database schema. The process uses canary deployments to validate changes before full promotion, allowing for rollback at critical points.

**Key Features:**
- **Backward Compatibility:** Legacy versions (v1) continue to work with new schema
- **Gradual Migration:** Each version is validated in canary before full promotion
- **Zero-Downtime:** Database evolves without stopping the application
- **Multi-Phase:** Shows v1 → v2 → v3 migration with schema changes

## Prerequisites

Before starting, ensure you have the following installed:

* **Kubernetes Cluster** (Minikube, Kind, or cloud provider)
* **kubectl** CLI
* **Argo Rollouts Controller**: [Installation Guide](https://argoproj.github.io/argo-rollouts/installation/)
* **Docker** (optional, only if building custom images)

## Project Structure

```
canarydemo/
├── apps/
│   ├── v1/app.go              # Legacy App (reads/writes full_name)
│   ├── v2/app.go              # Transition App (writes to 3 columns, reads from 2 new)
│   └── v3/app.go              # Final App (reads/writes first_name, last_name)
├── db_schema/
│   ├── 000_init_schema.sql    # Initial schema with full_name column
│   ├── 001_extend_schema.sql  # Add first_name, last_name columns
│   ├── 002_migrate_schema.sql # Populate new columns from full_name
│   └── 003_cleanup_schema.sql # Remove full_name column
├── Dockerfile                  # Multi-stage build for v1, v2, v3
├── deploy-postgres.yml         # PostgreSQL database deployment
├── rollout.yml                 # Argo Rollout & Service definitions
└── demo.sh                     # CLI tool for testing API endpoints
```

## Demo Phases

Follow the phases in sequence to see the complete migration flow:

### Phase 0: Environment Preparation and V1 Baseline App Ready

**Objective:** Set up the database and deploy v1 application.

1. Create namespace and deploy PostgreSQL database
2. Initialize database schema with `full_name` column
3. Deploy v1 application that reads/writes `full_name`
4. Verify v1 is operational

**Commands:**
```bash
kubectl create namespace stateful-app-demo
kubectl apply -f deploy-postgres.yml -n stateful-app-demo
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/000_init_schema.sql
kubectl apply -f rollout.yml
```

**Watch in separate terminals:**
```bash
# Terminal 1: Database
watch -n 2 "kubectl exec -n stateful-app-demo deploy/postgres -- psql -U postgres -d testdb -c 'SELECT * FROM users;'"

# Terminal 2: Rollout progress
kubectl argo rollouts get rollout canary-demo -n stateful-app-demo --watch
```

### Phase 1: V1 Running and Schema Expansion for V2

**Objective:** Demonstrate v1 working, then expand database schema for v2.

V1 continues to run while we add new columns. The application ignores the new columns.

**Commands:**
```bash
# Extend schema with first_name and last_name columns
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/001_extend_schema.sql

# Migrate data: populate first_name and last_name from full_name
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/002_migrate_schema.sql
```

**Test v1 still works:**
```bash
./demo.sh get stable "Isaac Newton"
./demo.sh add stable "Alan Turing"
```

### Phase 2: V2 Canary Deployment and Full Promotion

**Objective:** Deploy v2 as canary, validate it works with new schema, then promote to stable.

V2 reads from `first_name` and `last_name` columns and writes to all three columns.

**Commands:**
```bash
# Update rollout.yml to use v2 image, then deploy
kubectl apply -f rollout.yml

# Port forward to canary
kubectl port-forward svc/canary-demo-canary 8081:80 -n stateful-app-demo
```

**Test v2 canary:**
```bash
# Test v1 (stable) still works
./demo.sh get stable "Isaac Newton"

# Test v2 (canary) with new schema
./demo.sh get canary "Isaac" "Newton"
./demo.sh add canary "Anastasiia Gubska" "Anastasiia" "Gubska"
```

**Promote v2 to stable (3 times):**
```bash
kubectl argo rollouts promote canary-demo -n stateful-app-demo
```

Repeat command 3 times to fully promote v2 to 100% traffic.

### Phase 3: V3 Deployment and Cleanup Phase with Schema Finalization

**Objective:** Deploy v3 as canary, promote it, then drop the old schema column.

V3 reads and writes to `first_name` and `last_name` columns only (ignores `full_name`).

**Commands:**
```bash
# Update rollout.yml to use v3 image, then deploy
kubectl apply -f rollout.yml

# Port forward to canary
kubectl port-forward svc/canary-demo-canary 8081:80 -n stateful-app-demo
```

**Test v3 canary:**
```bash
# Test v3 (canary)
./demo.sh get canary "Anastasiia" "Gubska"
./demo.sh add canary "Mike Nelson"

# Test v2 (stable) still works
./demo.sh get stable "Isaac" "Newton"
```

**Promote v3 to stable (3 times):**
```bash
kubectl argo rollouts promote canary-demo -n stateful-app-demo
```

**⚠️ Point of No Return - Drop old schema:**
```bash
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/002_drop_old_schema.sql
```

**Verify v3 still works:**
```bash
./demo.sh get stable "Anastasiia" "Gubska"
./demo.sh add stable "Test1 Test2"
```

## Quick Start

### Using Pre-built Images

If you're using the pre-built images in the manifest:

```bash
# Phase 0: Setup
kubectl create namespace stateful-app-demo
kubectl apply -f deploy-postgres.yml -n stateful-app-demo
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/000_init_schema.sql

# Watch database and rollout in separate terminals
watch -n 2 "kubectl exec -n stateful-app-demo deploy/postgres -- psql -U postgres -d testdb -c 'SELECT * FROM users;'"
kubectl argo rollouts get rollout canary-demo -n stateful-app-demo --watch

# Deploy v1
kubectl apply -f rollout.yml
kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo
```

Then follow the commands in each Phase above.

### Building Custom Images

If you want to build your own images:

```bash
# Build v1
docker build --build-arg APP_VERSION=v1 -t your-repo/app:v1 .

# Build v2
docker build --build-arg APP_VERSION=v2 -t your-repo/app:v2 .

# Build v3
docker build --build-arg APP_VERSION=v3 -t your-repo/app:v3 .

# Update rollout.yml with your image names
# Then deploy
kubectl apply -f rollout.yml
```

## Demo CLI Tool

The `demo.sh` script provides convenient commands to interact with the application:

```bash
# Get user by name(s)
./demo.sh get stable "Isaac Newton"           # v1: full name
./demo.sh get canary "Isaac" "Newton"         # v2/v3: first name, last name

# Add user
./demo.sh add stable "Alan Turing"            # v1: full name
./demo.sh add canary "Charles" "Babbage"      # v2/v3: first name, last name
```

## Cleanup

When you're done with the demo:

```bash
# Delete the namespace and all resources
kubectl delete namespace stateful-app-demo

# Or manually delete components
kubectl delete -f rollout.yml -n stateful-app-demo
kubectl delete -f deploy-postgres.yml -n stateful-app-demo
kubectl delete namespace stateful-app-demo
```

## Key Concepts

### Decoupled Deployment
The database schema changes are decoupled from application deployment. Database changes happen first, followed by gradual application rollout.

### Backward Compatibility
Each version is compatible with the database schema it uses:
- **v1:** Reads/writes `full_name` only
- **v2:** Writes to all 3 columns, reads from new columns
- **v3:** Reads/writes new columns only

### Canary Deployment
With Argo Rollouts, new versions start with a small percentage of traffic (canary). Once validated, they're promoted to full traffic.

### Rollback Points
Rollback is possible until the cleanup phase. Once old schema is dropped, the migration is complete and irreversible.

## Troubleshooting

### Port forwarding not working
Ensure you're running the port-forward command in a separate terminal and it stays running.

### Application errors on new schema
Check that the database migration script (002_migrate_schema.sql) successfully populated the new columns.

### Canary not progressing
Watch the Argo Rollouts status to see if there are validation issues preventing promotion.

## Additional Resources

- [Argo Rollouts Documentation](https://argoproj.github.io/argo-rollouts/)
- [Kubernetes Stateful Applications](https://kubernetes.io/docs/tutorials/stateful-application/)
- [Database Migration Best Practices](https://en.wikipedia.org/wiki/Schema_evolution)

## Pro Tips

1. **Use separate terminals** for database watch, rollout watch, and test commands
2. **Check the database** at each phase to see schema changes in real-time
3. **Test both versions** during canary phase to ensure backward compatibility
4. **Review logs** if anything fails: `kubectl logs -n stateful-app-demo <pod-name>`