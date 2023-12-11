package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type OnBoardUnit struct {
	Axles    int     `json:"Axles"`
	Country  string  `json:"Country"`
	Credit   float64 `json:"Credit"`
	Currency string  `json:"Currency"`
	ID       string  `json:"ID"`
	SPZ      string  `json:"SPZ"`
	Weight   int     `json:"Weight"`
	Emission string  `json:"Emission"`
	Category string  `json:"Category"`
}

var gw *gateway.Gateway
var contract *gateway.Contract

var obuJsonList []OnBoardUnit

func GetObu(id, spz, country, dbType string) *OnBoardUnit {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return nil

	}
	result, err := contract.EvaluateTransaction("ReadObu", id, spz, country)
	if err != nil {
		fmt.Errorf("%v", err)
		return nil
	}
	var o OnBoardUnit
	err = json.Unmarshal(result, &o)
	return &o
}

func UpdateObu(id, spz, country, newEmission, newWeight, newAxles string, dbType string) {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return

	}
	_, err := contract.SubmitTransaction("UpdateObu", id, spz, country, newEmission, newWeight, newAxles)
	if err != nil {
		fmt.Errorf("%v", err)
	}
}

func SetTollAmount(o *OnBoardUnit, amount float64, dbType string) {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return
	}

	obuByte, err := contract.SubmitTransaction("TollRoadObu", o.ID, o.SPZ, o.Country, fmt.Sprintf("%.2f", amount))
	if err != nil {
		fmt.Errorf("%v", err)
	}
	err = json.Unmarshal(obuByte, o)
	if err != nil {
		fmt.Errorf("%v", err)
	}
}

func SetNullCredit(id, spz, country, dbType string) {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return
	}

	_, err := contract.SubmitTransaction("SetNullCredit", id, spz, country)
	if err != nil {
		fmt.Errorf("%v", err)
	}
}

func CreateObu(o *OnBoardUnit, dbType string) {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return
	}

	_, err := contract.SubmitTransaction("CreateObu", o.ID, o.SPZ, o.Country, o.Currency, o.Emission, o.Category, fmt.Sprintf("%d", o.Weight), fmt.Sprintf("%d", o.Axles))
	if err != nil {
		fmt.Errorf("%v", err)
	}

}

func DeleteObu(id, spz, country string, dbType string) {
	if contract == nil {
		fmt.Errorf("Database %s is not initialized", dbType)
		return
	}
	_, err := contract.SubmitTransaction("DeleteObu", id, spz, country)
	if err != nil {
		fmt.Errorf("%v", err)
	}
}

func InitDb(dbType string) {
	switch dbType {
	case "Blockchain":
		initDbBlockchain("channel1", "toll")
	case "JSON":
		initDbJson(filepath.Join("obu", "obuList.json"))
	}
}

func initDbBlockchain(chname string, ccname string) {
	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environment variable: %v", err)
	}

	walletPath := "wallet"
	// remove any existing wallet from prior runs
	os.RemoveAll(walletPath)
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}
	var ccpPath string = filepath.Join(
		"..",
		"..",
		"fabric",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err = gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}

	channelName := chname

	log.Println("--> Connecting to channel", channelName)
	network, err := gw.GetNetwork(channelName)
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	chaincodeName := ccname

	log.Println("--> Using chaincode", chaincodeName)
	contract = network.GetContract(chaincodeName)
	// result, err := contract.EvaluateTransaction("GetAllObus")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v", err)
	// }
	// fmt.Println(string(result))
}

func initDbJson(filepath string) {

}

func CloseDbBlockchain() {
	if gw != nil {
		gw.Close()
	}
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"fabric",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
