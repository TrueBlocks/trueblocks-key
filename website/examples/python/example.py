import requests
import json

# TODO: add quicknode url when we release on the Marketplace

# Counts appearances of an address by block number

url = "https://quick-node-url"
payload = json.dumps({
  "id": 1,
  "jsonrpc": "2.0",
  "method": "tb_getAppearances",
  "params": [{
    "address": "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
    "perPage": 10,
    "page": 1
  }]
})

headers = {
  "Content-Type": "application/json"
}

response = requests.request("POST", url, headers=headers, data=payload)
rpcResponse = json.loads(response.text)

appearances_by_block = {}

for appearance in rpcResponse["result"]:
    key = appearance["blockNumber"]
    current = appearances_by_block.get(key) or 0
    appearances_by_block[key] = current + 1

print(appearances_by_block)