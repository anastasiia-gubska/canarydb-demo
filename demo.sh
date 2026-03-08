#!/bin/bash

# Configuration for Port Forwarding
STABLE_URL="http://localhost:8080"
CANARY_URL="http://localhost:8081"

# Helper to pick the URL based on the keyword
get_url() {
    if [[ "$1" == "stable" ]]; then echo "$STABLE_URL"; else echo "$CANARY_URL"; fi
}

case "$1" in
  # Usage: ./demo.sh add stable "Full Name" "First" "Last"
  # Usage: ./demo.sh add canary "Full Name" "First" "Last"
  add)
    URL=$(get_url "$2")
    FULL="$3"
    FIRST="$4"
    LAST="$5"

    echo "Adding to $2 ($URL)..."
    curl -X POST -H "Content-Type: application/json" \
    -d "{
        \"full_name\": \"$FULL\",
        \"first_name\": \"$FIRST\",
        \"last_name\": \"$LAST\"
    }" "$URL/user"
    echo -e "\n"
    ;;

  add_v3)
    URL=$(get_url "$2")
    FIRST="$3"
    LAST="$4"

    echo "Adding to $2 ($URL)..."
    curl -X POST -H "Content-Type: application/json" \
    -d "{
        \"first_name\": \"$FIRST\",
        \"last_name\": \"$LAST\"
    }" "$URL/user"
    echo -e "\n"
    ;;

  # Usage: ./demo.sh get stable "Full Name"
  # Usage: ./demo.sh get stable "First" "Last"
  # Usage: ./demo.sh get canary "First" "Last"
  get)
    URL=$(get_url "$2")
    
    # Check if we are searching by Full Name (1 arg) or First/Last (2 args)
    if [ "$#" -eq 3 ]; then
        # Search by Full Name (v1 style)
        NAME="$3"
        echo "Querying $2 ($URL) by Full Name: $NAME"
        curl -s -G --data-urlencode "full_name=$NAME" "$URL/user"
    elif [ "$#" -eq 4 ]; then
        # Search by First/Last (v2/v3 style)
        FNAME="$3"
        LNAME="$4"
        echo "Querying $2 ($URL) by First/Last: $FNAME $LNAME"
        curl -s -G \
          --data-urlencode "first_name=$FNAME" \
          --data-urlencode "last_name=$LNAME" \
          "$URL/user"
    fi
    echo -e "\n"
    ;;

  *)
    echo "Usage Examples:"
    echo "  ./demo.sh add stable \"John Doe\" \"John\" \"Doe\""
    echo "  ./demo.sh get stable \"John Doe\""
    echo "  ./demo.sh get canary \"John\" \"Doe\""
    ;;
esac