import subprocess
import time

def cmd(command):
    print(f"cmd: {command}")
    return subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)
# Dictionary to keep track of active background processes
# Format: { local_port: process_object }
active_pfs = {}

def start_pf(svc, local_port):
    global active_pfs
    # 1. Kill any existing port-forward on this local port
    subprocess.run(f"lsof -ti:{local_port} | xargs kill -9", shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    #if local_port in active_pfs:
    #    print(f">> Closing old port-forward on {local_port}...")
    #    active_pfs[local_port].terminate()
    #    active_pfs[local_port].wait()

    # 2. Start the new port-forward in the background
    print(f"> magic pfw: {local_port} -> {svc}")
    full_cmd = f"kubectl port-forward svc/{svc} {local_port}:80 -n stateful-app-demo"
    proc = subprocess.Popen(full_cmd, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    active_pfs[local_port] = proc
    time.sleep(5) # Brief pause to allow the tunnel to open

try:
    while True:
        step = input("\nEnter Step [1-5] (q to quit): ")

        if step == "1":
            print(">> SETUP: Deploy V1 & Database")
            cmd("kubectl apply -f deploy-postgres.yml")
            cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/000_init_schema.sql")
            #cmd("kubectl exec -n stateful-app-demo deploy/postgres -- psql -U postgres -d testdb < db_schema/000_init_schema.sql")
            cmd("kubectl apply -f rollout_v1.yml")
            start_pf("canary-demo-stable", 8080)
            #start_pf("canary-demo-stable", 8080)
            #cmd("kubectl port-forward svc/canary-demo-stable 8080:80 -n stateful-app-demo")
            cmd("./demo.sh get stable 'Isaac Newton'")
            #cmd("./demo.sh add stable 'Alan Turing'")
            print("\n>> ENTER 2: EXPAND & MIGRATE")

        elif step == "2":
            #cmd("./demo.sh add stable 'Alan Turing'")
            print(">>\n EXPAND: Adding FIRST_NAME and LAST_NAME Columns")
            #cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/001_extend_schema.sql")
            print("\n")
            cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/001_extend_schema.sql")
            start_pf("canary-demo-stable", 8080)
            cmd("./demo.sh add stable 'George Boole'")
            time.sleep(10)
        #elif step == "3":
            print("\n>> MIGRATE: Syncing Data")
            print("\n")
            cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/002_migrate_schema.sql")
            #print(">> V1 READS FULL_NAME")
            #cmd("./demo.sh get stable 'Charles Darwin'")
            #cmd("./demo.sh add stable 'George Boole'")
            print("\n>> ENTER 3: Deploy V2")

        elif step == "3":
            print(">> ROLLOUT V2: Dual Writing")
            print("\n")
            cmd("kubectl apply -f rollout_v2.yml")
            time.sleep(5) # Wait for rollout start
            start_pf("canary-demo-canary", 8081)
            #cmd("kubectl port-forward svc/canary-demo-canary 8081:80 -n stateful-app-demo")
            print("\n>> V2 READS NEW COLUMNS")
            cmd("./demo.sh get canary 'Charles' 'Darwin'")
            cmd("kubectl argo rollouts promote canary-demo -n stateful-app-demo")
            print("\n>> WATCH YELLOW V2: CANARY 50%") 
            time.sleep(10) # Wait for DEMO
            cmd("kubectl argo rollouts promote canary-demo -n stateful-app-demo --full")
            print("\n>> V2 GREEN NOW: CANARY FULLY PROMOTED") 
            start_pf("canary-demo-stable", 8080)
            #time.sleep(5) # Wait for DEMO
            print("\n>> V2 WRITES TO ALL THREE COLUMNS (John Venn)")
            print("\n>> ENTER 3 AGAIN IF NO NAME (John Venn)")
            start_pf("canary-demo-stable", 8080)             
            cmd("./demo.sh add stable 'John Venn' 'John' 'Venn'")
            print("\n>> ENTER 4: Deploy V3")


        elif step == "4":
            start_pf("canary-demo-stable", 8080)
            print("\n>> V2 READS NEW COLUMNS")
            cmd("./demo.sh get stable 'Isaac' 'Newton'")
            print("\n>> ROLLOUT V3: New Schema Only")
            cmd("kubectl apply -f rollout_v3.yml")
            time.sleep(5) # Wait for rollout start
            start_pf("canary-demo-canary", 8081)
            print("\n>> V3 READS NEW COLUMNS")
            cmd("./demo.sh get canary 'Charles' 'Darwin'")
            cmd("kubectl argo rollouts promote canary-demo -n stateful-app-demo --full")
            #time.sleep(10) # Wait for promotion to complete
            print("\n>> ENTER 5: DROP full_name")

        elif step == "5":
            print("\n>> CONTRACT: Dropping Full Name")
            #cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/003_cleanup_schema.sql")
            cmd("kubectl exec -n stateful-app-demo deploy/postgres -it -- psql -U postgres -d testdb < db_schema/003_cleanup_schema.sql")
            start_pf("canary-demo-stable", 8080)
            print("\n>> V3 WRITES NEW COLUMNS - Alan Turing")
            cmd("./demo.sh add_v3 stable 'Alan' 'Turing'")
            print("Final Clean State Reached.")
            print("\n>> ENTER q: COMPLETED")

        elif step == "q": break
finally:
    for p in active_pfs: p.kill()