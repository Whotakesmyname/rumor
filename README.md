# rumor
A decentralized social networking service capable of NAT traversal.

## Introduction
This project is still under developing with no release.

This project is designed to function like Twitter, however, in a decentralized way. To enlarge the user group and make the most of the distributed system,
mobile terminals should be considered. Thus, NAT traversal is a scheduled functionality.

## Design
### DHT
The decentralization is built on the Kademlia Distributed Hash Table algorithm. Most of it followed the original paper of Kadmelia, despite of some customizing
communication protocols.

### NAT Traversal
Because the requirement of decentralization, the signaling server becomes the obstacle, and the traversal between symmetrical NATs becomes impossible.
The solution is to employ XMPP as a way to transfer signals. By this way, the decentralization feature could stay alive and most data could be protected.

### IPC
The communication between cli and daemon employs named pipe on Windows and unix sockets on Unix-like systems.

## Status Quo
This is still far from alpha. The structure needed for Kademlia DHT has been finished. The overall framework of 'CLI + Daemon' has also been established.

## Temporary Drawbacks
- Lacking objects reuse
- System-level configuration needs to be done before compilation

## Thanks to
This list may not be complete
- docopt
- XMPP and public servers
- go-stun and public stun servers
- go-winio from Microsoft
- ...

For licenses please refer to those libraries' pages temporarily.
