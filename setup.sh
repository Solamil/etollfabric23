export $(./setOrgEnv.sh "${1:-Org1}" | xargs )
./network.sh up createChannel -ca -c "$CHANNEL_NAME" && ./scripts/deployCC.sh "$CHANNEL_NAME" "$CONTRACT_NAME" ../asset-toll/chaincode-go/ go 1 1 'InitLedger'

# ./network.sh deployCC -ccn $CONTRACT_NAME -c $CHANNEL_NAME -ccp ../asset-transfer-$CONTRACT_NAME/chaincode-go -ccl go

# peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C $CHANNEL_NAME -n $CONTRACT_NAME --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'

# peer chaincode query -C $CHANNEL_NAME -n $CONTRACT_NAME -c '{"Args":["GetAllAssets"]}'
# peer chaincode query -C $CHANNEL_NAME -n $CONTRACT_NAME -c '{"Args":["ReadAsset","asset6"]}'

# Export contract into package for loading on the network
# peer lifecycle chaincode package ${CONTRACT_NAME}.tar.gz --path ../asset-toll/chaincode-go/ --lang golang --label ${CONTRACT_NAME}_1.0 && peer lifecycle chaincode install ${CONTRACT_NAME}.tar.gz
# peer lifecycle chaincode queryinstalled

# peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile "${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem" -C $CHANNEL_NAME -n $CONTRACT_NAME --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'

