#!/bin/bash

TARGET="${1:-http://localhost:8080}"
EMAIL="prober_user@example.com"
PASSWORD="prober_password123"

GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
RESET="\033[0m"

format_status() {
  local code="$1"
  if [[ "$code" =~ ^2 ]]; then
    echo -e "${GREEN}${code}${RESET}"
  elif [[ "$code" =~ ^4 ]]; then
    echo -e "${YELLOW}${code}${RESET}"
  else
    echo -e "${RED}${code}${RESET}"
  fi
}

echo -e "[*] Launching the prober on ${TARGET}..."

curl -s -o /dev/null -X POST "${TARGET}/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}"

get_token() {
  local resp
  resp=$(curl -s -X POST "${TARGET}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")
  
  echo "$resp" | jq -r '.access_token // empty' 2>/dev/null
}

TOKEN=$(get_token)
if [[ -z "$TOKEN" ]]; then
  echo -e "${RED}[!] Failed to receive an authorization token${RESET}"
else
  echo -e "${GREEN}[+] Authorization token received${RESET}"
fi

iteration=0

while true; do
  ((iteration++))

  if [[ -z "$TOKEN" ]]; then
    TOKEN=$(get_token)
  fi

  code_static=$(curl -s -o /dev/null -w "%{http_code}" "${TARGET}/catalog.html")

  resp_create=$(curl -s -w "\n%{http_code}" -X POST "${TARGET}/api/v1/items" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Prober Item #${iteration}\"}")

  code_create=$(echo "$resp_create" | tail -n1)
  body_create=$(echo "$resp_create" | sed '$d')
  item_id=$(echo "$body_create" | jq -r '.id // empty' 2>/dev/null)

  if [[ "$code_create" == "401" ]]; then
    TOKEN=""
  fi

  code_update="---"
  if [[ -n "$item_id" ]]; then
    code_update=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "${TARGET}/api/v1/items/${item_id}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      -d "{\"title\":\"Updated Item #${iteration}\"}")
  fi

  code_list=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${TARGET}/api/v1/items" \
    -H "Authorization: Bearer ${TOKEN}")

  code_delete="---"
  if [[ -n "$item_id" ]]; then
    code_delete=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "${TARGET}/api/v1/items/${item_id}" \
      -H "Authorization: Bearer ${TOKEN}")
  fi

  echo -e "[$(date +%T)] ${iteration} | Static: $(format_status "$code_static") | Create: $(format_status "$code_create") | Update: $(format_status "$code_update") | List: $(format_status "$code_list") | Delete: $(format_status "$code_delete")"

  sleep 1
done