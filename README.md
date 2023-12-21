## Electronic Toll Road via Blockchain
The bachelor thesis on electronic toll road.
This project applied Blockchain technology especially DLT ([Distributed Ledger Technology](https://en.wikipedia.org/wiki/Distributed_ledger)) - [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/en/latest/index.html) to record payments for toll road from OBU (On Board Unit).
The toll road system has many similarities with toll road system used especially in [Czechia](https://mytocz.eu/en/) and the EU.

The project is structured as followed:

- `obu/` On Board Unit(OBU) simulate the driving on toll road and then send the information about driven toll road back to the server. 
- `server/`  The server opens connection to Fabric database and comunicates with OBUs.
- `asset-toll/` The smart contract for Fabric, written in Go. It stores informations about OBUs and their payments.

## Prerequisites

- `Go` version >= 1.19, `Git`

The application is written in [Golang](https://go.dev/).

## Installation

- Follow [these](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html) instructions to install prerequisites for Hyperledger Fabric.
- Then go onto [these](https://hyperledger-fabric.readthedocs.io/en/latest/install.html) instructions to install Fabric and Fabric samples.

- Clone this project `git clone https://github.com/Solamil/etollfabric23`

- The directory `asset-toll/` copy to your directory `fabric-samples/` directory.

-  Additionally copy modified files `setOrgEnv.sh` and `setup.sh` into directory `fabric-samples/test-network/` to be more comfortable.


## Use
- Start Fabric test network database. At directory `test-network/`, run `export $(./setOrgEnv.sh)` then `./setup.sh`.
- Start the server `cd server/ && go run ./cmd/server/main.go`. It starts http server listens on default port 8905 and connects itself to Fabric.
- Start the OBU. `cd obu/ && go run main.go` Results are then written into Fabric database.

## Author
michal.kukla@tul.cz
2023
