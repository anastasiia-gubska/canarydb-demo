#!/bin/bash

NAMESPACE="stateful-app-demo"

echo "Starting Port-Forwarding for Demo..."

# Start the forwards in the background
sudo kubectl port-forward svc/canary-demo-stable 8888:80 -n $NAMESPACE > /dev/null 2>&1 &
sudo kubectl port-forward svc/canary-demo-canary 9999:80 -n $NAMESPACE > /dev/null 2>&1 &
sudo kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE > /dev/null 2>&1 &

echo "-------------------------------------------------------"
echo " Ports are now active:"
echo "   - Stable (v1):   http://localhost:8888"
echo "   - Canary (v2):   http://localhost:9999"
echo "   - Postgres:      localhost:5432"
echo "-------------------------------------------------------"
echo "Now run kubectl get po -A to check"

# Ensure that even if you just close the terminal, it tries to kill the children
trap "echo 'Stopping...'; pgrep -f 'kubectl port-forward' | xargs kill -9 2>/dev/null; exit" SIGINT SIGTERM