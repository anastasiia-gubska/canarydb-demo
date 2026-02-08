#!/bin/bash

# Usage: ./demo.sh add_v1 "Full Name" "email@example.com"
# Usage: ./demo.sh add_v2 "First" "Last" "email@example.com"
# Usage: ./demo.sh add_v2 "Full Name" "First Name" "email@examle"
# Usage: ./demo.sh clean

# Configuration
V1_URL="http://localhost:8080"
V2_URL="http://localhost:8081"

case "$1" in
  # Usage: ./demo.sh add_v1 "Full Name" "email"
  add_v1)
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"full_name\": \"$2\", \"email_addr\": \"$3\"}" \
    $V1_URL/user
    ;;

  # Usage: ./demo.sh add_v2 "First" "Last" "email"
  add_v2)
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"first_name\": \"$2\", \"last_name\": \"$3\", \"email_addr\": \"$4\"}" \
    $V2_URL/user
    ;;

  # Usage: ./demo.sh add_v2_all "Full" "First" "Last" "email"
  add_v2_all)
    curl -X POST -H "Content-Type: application/json" \
    -d "{\"full_name\": \"$2\", \"first_name\": \"$3\", \"last_name\": \"$4\", \"email_addr\": \"$5\"}" \
    $V2_URL/user
    ;;

  clean_v1)
    curl -X POST $V1_URL/clean
    ;;

  clean_v2)
    curl -X POST $V2_URL/clean
    ;;

  *)
    echo "Commands available:"
    echo "  ./demo.sh add_v1 \"Name\" \"Email\""
    echo "  ./demo.sh add_v2 \"First\" \"Last\" \"Email\""
    echo "  ./demo.sh add_v2_all \"Full\" \"First\" \"Last\" \"Email\""
    echo "  ./demo.sh clean_v1"
    echo "  ./demo.sh clean_v2"
    ;;
esac