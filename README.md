# spdkctrl

spdkctrl is a golang package to operate SPDK application through JsonRPC.

# Usage

Users may find usage examples from the test cases

## app

app_test.go shows how to use spdkctrl to start and terminate SPDK application

## client

client_test.go shows client how to use spdkctrl to connect to a running SPDK application

## rpc

rpc_test.go shows how to send RPC commands through connected client to SPDK application.

* Note: more RPC methods are required to add.
