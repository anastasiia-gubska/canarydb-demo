# Multi-Phase Canary Rollout Guide

A step-by-step guide for demonstrating database schema evolution with canary deployments using Kubernetes and Argo Rollouts.

## Overview

This guide demonstrates how to safely migrate an application through multiple versions while evolving the database schema. The process uses canary deployments to validate changes before full promotion, allowing for rollback at critical points.

---

## Phase 0: Environment Preparation

### Deploy PostgreSQL Database

```bash
kubectl apply -f deploy-postgres.sql
```

### Initialize Database Schema

```bash
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/000_init_schema.sql
```

### Watch Database Activity ( in separate terminal)

Open a terminal to monitor the `users` table:

```bash
watch -n 2 "kubectl exec -n stateful-app-demo deploy/postgres -- psql -U postgres -d testdb -c 'SELECT * FROM users;'"
```
**Note:** The PostgreSQL database is now running with the initial schema ready for the application deployments.

## Deploy Version 1 Application (V1)

### Overview
V1 app reads and writes to the `full_name` column only. This establishes the baseline application.

### Deploy V1 Application

```bash
kubectl apply -f rollout.yml
```
### Watch Rollout Status ( in separate terminal)

Monitor the canary rollout progress:

```bash
kubectl argo rollouts get rollout canary-demo -n stateful-app-demo --watch
```

### Enable Port Forwarding

Access the V1 application:

```bash
kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo
```

### Test V1 Functionality

Verify that V1 works as expected by retrieving and adding user records:

```bash
./demo.sh get stable "Isaac Newton"
./demo.sh add stable "Alan Turing"
```

**Expected Result:** Successfully retrieves full names and adds new full names to the database.

The V1 app is now deployed and ready for demonstration.

----
## Phase 1: V1 Running and Expand Schema

### Overview
V1 is already running and reads/writes to the `full_name` column only. Now, expand the database schema in preparation for v2 while v1 continues work as expected.

### Prepare Database for V2

Extend the schema to add `first_name` and `last_name` columns:

```bash
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/001_extend_schema.sql
```

Populate the new columns by migrating existing data:

```bash
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/002_migrate_schema.sql
```

**Note:** V1 continues to work as it only reads/writes to the `full_name` column, ignoring the new columns.

---

## Phase 2: Canary Deploy Version 2 (V2)

### Overview
V2 app reads from `first_name` and `last_name` columns and writes to all three columns (`full_name`, `first_name`, `last_name`).

### Deploy V2 as Canary

Change v1 to v2 and deploy:

```bash
kubectl apply -f rollout.yml
```

This deployment runs V2 alongside V1, allowing validation before full promotion.

### Enable Canary Port Forwarding

```bash
kubectl port-forward svc/canary-demo-canary 8081:80 -n stateful-app-demo
```

### Test V1 (Stable) Still Works

Verify the stable V1 remains functional:

```bash
./demo.sh get stable "Isaac Newton"
```

### Test V2 (Canary) Functionality

Test retrieval and addition with the new schema:

```bash
./demo.sh get canary "Isaac" "Newton"
./demo.sh add canary "Anastasiia Gubska" "Anastasiia" "Gubska"
```

**Expected Result:** V2 successfully reads from and writes to both the old and new columns.

### Promote V2 to Stable

Once validated, promote V2 through three promotion steps:

```bash
kubectl argo rollouts promote canary-demo -n stateful-app-demo
```

**NOTE:** Repeat the promotion command three times to fully transition from canary to stable.

### Update Port Forwarding

Access the V2 application as stable:

```bash
kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo
```

---

## Phase 3: V3 Deployment and Cleanup Phase with Schema Finalization (V3)

### Overview
V3 app reads and writes to `first_name` and `last_name` columns only, ignores `full_name`.
Deploy V3 as a canary, validate it works correctly, promote it to stable, then perform the cleanup phase by dropping the old schema. This is the point of no return—once the old schema is dropped, rollback is impossible.
### Deploy V3 as Canary

Change v2 to v3 and deploy

```bash
kubectl apply -f rollout.yml
```

### Update Canary Port Forwarding

```bash
kubectl port-forward svc/canary-demo-canary 8081:80 -n stateful-app-demo
```

### Test V3 (Canary) Functionality

```bash
./demo.sh get canary "Anastasiia" "Gubska"
./demo.sh get stable "Anastasiia" "Gubska"
./demo.sh add_v3 canary "Mike" "Nelson"
```

**Expected Result:** V3 operates correctly with the new schema structure.

### Verify V2 (Stable) Still Works

Switch to the stable service to confirm V2 is unaffected:

```bash
kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo
```

Test V2 functionality:

```bash
./demo.sh get stable "Isaac" "Newton"
```

### Promote V3 to Stable

Once validated, promote V3 through three promotion steps:

```bash
kubectl argo rollouts promote canary-demo -n stateful-app-demo
```

Repeat three times to fully transition.

### Final Canary Verification

Verify V3 is now fully deployed:

```bash
kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo
./demo.sh get stable "Anastasiia" "Gubska"
```

---

## Cleanup and Schema Finalization

**⚠️ Point of No Return:** Once executed, rollback is impossible.
### Drop Old Schema

Remove the `full_name` column:

```bash
kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/003_cleanup_schema.sql#
```

### Verify V3 Functionality

Confirm the application works correctly without the old column:

```bash
./demo.sh get stable "Anastasiia" "Gubska"
./demo.sh get stable "Alan" "Turing"
```

**Expected Result:** V3 continues to function as expected, confirming the migration is complete.

---

## Summary

The multi-phase rollout successfully demonstrates:

1. **Phase 0:** Environment setup with initial schema and v1 baseline app ready
2. **Phase 1:** v1 running and schema expansion for v2
4. **Phase 2:** v2 canary deployment and full promotion
5. **Phase 3:** v3 deployment and cleanup phase with schema finalization

At each phase, canary deployments allow validation before committing to the change. Once Phase 3 is complete, the old schema is gone and the application is fully migrated to the new version.

## Key Takeaways

- **Canary deployments** enable safe validation of new versions before full rollout
- **Schema evolution** can be managed gradually with backward-compatible intermediate versions
- **Rollback points** exist until the cleanup phase
- **Multi-phase promotion** ensures stability before advancing to the next version