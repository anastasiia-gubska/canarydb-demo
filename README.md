# Decoupled Canary Database Demo

This demo shows a decoupled **Canary Deployment** from Database using **Argo Rollouts** for stateful applications. The core focus is on managing database schema changes (adding `first_name` and `last_name` columns) while ensuring that the legacy version (v1) and the new version (v2) can coexist without crashing.

## Prerequisites
Before starting, ensure you have the following installed:
* **Kubernetes Cluster** (Minikube, or Kind)
* **kubectl** CLI
* **Argo Rollouts Controller**: [Installation Guide](https://argoproj.github.io/argo-rollouts/installation/)

## Project Structure
```text
canarydemo/
├── apps/
│   ├── v1/app.go         # Legacy App (handles full_name)
│   └── v2/app.go         # New App (handles split names)
├── Dockerfile            # Multi-stage build for v1 and v2
├── deploy-postgres.yml   # Database
├── rollout.yml           # Argo Rollout & Service definitions
├── start_demo.sh         # Port-forwarding automation
└── demo.sh               # CLI tool for testing API endpoints
```
## Step 1. Deployment
### 1. Setup Namespace and Database
First, create the dedicated namespace and deploy the PostgreSQL database.

```
kubectl create namespace stateful-app-demo
kubectl apply -f deploy-postgres.yml -n stateful-app-demo
```
### 2. Deploy app v1 (stable)
Apply the Rollout manifest. Ensure the image in rollout.yml is your v1 image.

```
kubectl apply -f rollout.yml 
```

### 3. Start Local Access
Run this in a new terminal to link your local machine to the cluster (port-fowrarding):

```
chmod +x start_demo.sh
./start_demo.sh
```
- Stable (v1): http://localhost:8888
- Canary (v2): http://localhost:9999 (won't be enabled)
- Database: localhost:5432

## Step 2. Interactive Demo
Follow these steps across your terminal windows to see the magic.
#### Terminal Window 1: Database
Run this to see the table structure and data change in real-time:
```
watch -n 1 'kubectl exec -n stateful-app-demo deploy/postgres -- psql -U postgres -d testdb -c "SELECT * FROM users;"'
```
#### Terminal Window 2: Argo Rollout 
Monitor the rollout progress:
```
kubectl argo rollouts get rollout canary-demo -n stateful-app-demo --watch
```
#### Terminal Window 3: Test
Phase 1: Stable (v1)
Verify v1 can read data.
```
./demo.sh get_user "Anastasiia Gubska"
```
Phase 2: Update `rollout.yml` to use the `v2` image, then apply
```
kubectl apply -f rollout.yml -n stateful-app-demo
```
<i>Argo will pause at 25% traffic.</i> Watch <b>Terminal Window 2</b> to see this.

#### Terminal Window 4: Enable app v2 port-forwarding
```
kubectl port-forward svc/canary-demo-canary 9999:80 -n stateful-app-demo
```
or re-run `./start_demo.sh`

Phase 3: Magic Moment - Add a user via Canary (v2) endpoint:
```
./demo.sh add_v2 "Jane" "Doe" "jane@example.com" (WRONG) v1 won't see full_name
./demo.sh add_v2_all "Isaac Newton" "Isaac" "Newton" "isaac.newton@example.com"
```
Watch <b>Terminal Window 1</b>. You'll see `first_name` and `last_name` columns appear in database instantly.

Phase 4: Verify Backward Compatibility
Check if the Stable (v1) endpoint still works despite the new column
```
./demo.sh get_user "Jane Doe" (User not found)
./demo.sh get_user "Isaac Newton" (via v1) 
```
<i>if it returns data without 500 error, your decoupling succeeded!</i>

### Step 3: Completion
Once confirming v1 is safe, promote v2 to 100% traffic:
```
kubectl argo rollouts promote canary-demo -n stateful-app-demo
```

### Cleanup

### Clean database to delete v2 inputs
```
./demo.sh clean_v1 (start fresh)
./demo.sh clean_v2 (delete v2 inputs)
```

### Wrap up demo
```
kubectl delete namespace stateful-app-demo
kubectl delete -f rollout.yml
kubectl delete -f deploy-postgres.yml
```

### Pro-Tip for Users
If you are building your own images instead of using the ones in the manifest, use the build arguments provided in the `Dockerfile`:

```
# Build v1
docker build --build-arg APP_VERSION=v1 -t your-repo/app:v1 .

# Build v2
docker build --build-arg APP_VERSION=v2 -t your-repo/app:v2 .
```