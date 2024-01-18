Getting started
===============

TrueBlocks Key is currently accessible only through QuickNode Marketplace. To sign up, please visit [TODO: add marketplace link](https://www.quicknode.com).

Once you pick a plan, a new RPC method will be made available: `tb_getAppearances`. It can be used just as any other RPC method: via QuickNode SDK, language-specific
libraries (e.g. Ethers.js) and command line tools (e.g. cURL).

Let's get first 1,000 appearances of address `0xf503017d7baf7fbc0fff7492b751025c6a78179b`:
```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"tb_getAppearances","params":[{ "address": "0xf503017d7baf7fbc0fff7492b751025c6a78179b", "perPage": 1000, "page": 1 }],"id":1}' https://your-quicknode-endpoint
```

Note the `params` field. It is an array, just like in the standard RPC methods, but it can only hold a single item. This item is where we define all the details:
- the address
- how many items per page we want
- which page we need.

Here is an example response:
```json
{
  "jsonrpc": "2.0",
  "id": 0,
  "result": [
    {
      "blockNumber": 18487775,
      "transactionId": 125
    },
    {
      "blockNumber": 18487771,
      "transactionId": 98
    },
// fast forward to the 1000th item
    {
      "blockNumber": 17291975,
      "transactionId": 54
    }
  ]
}
```

The returned data can be used for statistics or to fetch the details of transactions of interest.