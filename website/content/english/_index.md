---
# Banner
banner:
  title: "Everything that ever happened anywhere on the chain"
  content: "Get a complete historical list of appearances (block number, transaction id) for any Ethereum address"
  button:
    enable: true
    label: "Try on QuickNode Marketplace"
    link: ""

# Features
features:
  - title: "Appearance Index API"
    content: "A single API endpoint returns significantly more appearances (block number, transaction ID) than other indexers. Includes:"
    bulletpoints:
      - "regular transactions"
      - "internal transactions"
      - "token transfers"
      - "withdrawals"
    footer: "The API can also return total number of appearances for any Ethereum address."
    button:
      enable: true
      label: "Read docs"
      link: "#"

  - title: "Coming soon"
    content: ""
    bulletpoints:
      - "Appearance Index API: date and time queries"
      - "Neighbors Index API: Given any Ethereum address, this endpoint returns a list of all other addresses that appeared in the same transactions as the given address. If one were to ever solve the Sybil problem, the solution would start with this list of neighbors."
      - "Unchained Index API: The <a href=\"https://trueblocks.io/papers/2023/specification-for-the-unchained-index-v2.0.0-release.pdf\">Unchained Index</a> is a novel way to distribute immutable databases such as blockchain data and our Appearance Index API data using IPFS. This endpoint returns a list of IPFS hashes for portions of the indexed data relevant to your address. You may download these portions locally making it impossible for anyone (including us) to censor your access later."
      - "Custom APIs: Do you have a need in mind that doesn't appear here? We have you covered. We can index anything including data extraction customized for your smart contract (or any) needs."
      - "FourByte Index API: Given N function names and M function signatures, this API generates NxM fourbyte signatures. We then attach a frequency count to the number times each of these NxM fourbytes are found in the wild (on-chain). This allows us to decode call data and event topics without hard-to-find ABIs or a reliance on a contract's ABI or address. We've generated many 100s of millions of fourbytes. "
      - "Applications: A collection of local-first, blazingly-fast, perfectly-private desktop applications demonstrating the power of TrueBlocks. (Including 18-decimal-place, perfect accounting off-chain.)"
      - "Dalle-dress API"
    button:
      enable: false
      label: "Get Started Now"
      link: "https://github.com/zeon-studio/hugoplate"

  - title: "Based on decentralized public good"
    content: "TrueBlocks Key uses the index of appearances created by TrueBlocks Core, which publishes the index as a public good through the Unchained Index smart contract. Unlike Key, Core is fully local and open source. Both are designed to work with any EVM-based blockchain. TrueBlocks Key provides a simple, easy-to-use (for pay) Web 2.0 interface to this index. "
    button:
      enable: true
      label: "Learn more about TrueBlocks"
      link: "https://trueblocks.io"
---
