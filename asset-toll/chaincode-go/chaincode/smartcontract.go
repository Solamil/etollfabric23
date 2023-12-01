package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

const obuIndex = "id~spz~country"

type OnBoardUnit struct {
	Axles 	        int     `json:"Axles"`
	Country	        string  `json:"Country"`
	Credit		float64 `json:"Credit"`
	Currency        string  `json:"Currency"`
	ID              string  `json:"ID"`
	SPZ             string  `json:"SPZ"`
	Weight 	        int     `json:"Weight"`
	Emission      	string  `json:"Emission"` 
	Category	string  `json:"Category"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	obuList := []OnBoardUnit{
		{ID: "2c9fa1aa-4403-4cc9-96f4-09a05638bcad", Country: "CZ", SPZ: "1SA1234", Credit: 0.0, 
		Currency: "CZK", Weight: 8500, Emission: "6", Category: "N", Axles: 4 },
		{ID: "7873527e-4d58-4e94-a71c-8ad908f59e00", Country: "CZ", SPZ: "1S15244", Credit: 40.0, 
		Currency: "CZK", Weight: 12500, Emission: "2", Category: "N", Axles: 5 },
	}

	for _, obu := range obuList {
		obuJSON, err := json.Marshal(obu)
		if err != nil {
			return err
		}
		id, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{obu.ID, obu.SPZ, obu.Country})	
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)

		} 
		err = ctx.GetStub().PutState(id, obuJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}
func (s *SmartContract) CreateObu(ctx contractapi.TransactionContextInterface, id, spz, country, currency, emission, category string, weight, axles int) error {
	exists, err := s.ObuExists(ctx, id, spz, country)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the onBoardUnit %s already exists", id)
	}

	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)

	} 
	obu := OnBoardUnit{
		ID:             id,
		SPZ:          	spz,
		Country:        country,
		Credit:         0.0,
		Currency: 	currency,
		Emission:	emission,
		Weight:		weight,
		Axles:		axles,
	}
	obuJSON, err := json.Marshal(obu)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idObu, obuJSON)
}
func (s *SmartContract) TollRoadObu(ctx contractapi.TransactionContextInterface, id, spz, country string, sum float64) (*OnBoardUnit, error) {
	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	} 
	obuJSON, err := ctx.GetStub().GetState(idObu)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if obuJSON == nil {
		return nil, fmt.Errorf("the obu %s does not exist", idObu)
	}
	var obu OnBoardUnit
	err = json.Unmarshal(obuJSON, &obu)
	if err != nil {
		return nil, err
	}
	obu.Credit += sum

	obuJSON, err = json.Marshal(obu)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(idObu, obuJSON)
	if err != nil {
		return nil, err
	}
	return &obu, nil
}

func (s *SmartContract) UpdateObu(ctx contractapi.TransactionContextInterface, id, spz, country, newEmission string, newWeight, newAxles int) error {
	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	} 
	obuJSON, err := ctx.GetStub().GetState(idObu)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if obuJSON == nil {
		return fmt.Errorf("the obu %s does not exist", idObu)
	}
	var obu OnBoardUnit
	err = json.Unmarshal(obuJSON, &obu)
	if err != nil {
		return err
	}
	obu.Emission = newEmission
	obu.Weight = newWeight
	obu.Axles = newAxles
	obuJSON, err = json.Marshal(obu)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(idObu, obuJSON)

}
func (s *SmartContract) ReadObu(ctx contractapi.TransactionContextInterface, id, spz, country string) (*OnBoardUnit, error) {
	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)

	} 
	obuJSON, err := ctx.GetStub().GetState(idObu)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if obuJSON == nil {
		return nil, fmt.Errorf("the obu %s does not exist", idObu)
	}

	var obu OnBoardUnit
	err = json.Unmarshal(obuJSON, &obu)
	if err != nil {
		return nil, err
	}

	return &obu, nil
}

func (s *SmartContract) DeleteObu(ctx contractapi.TransactionContextInterface, id, spz, country string) error {
	exists, err := s.ObuExists(ctx, id, spz, country)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the obu %s with parameters %s, %s, does not exist", id, spz, country)
	}

	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)

	} 
	return ctx.GetStub().DelState(idObu)
}

func (s *SmartContract) SetNullCredit(ctx contractapi.TransactionContextInterface, id, spz, country string) error {
	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)

	} 
	obuJSON, err := ctx.GetStub().GetState(idObu)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if obuJSON == nil {
		return fmt.Errorf("the obu %s does not exist", idObu)
	}

	var obu OnBoardUnit
	err = json.Unmarshal(obuJSON, &obu)
	obu.Credit = 0

	if err != nil {
		return err
	}
	obuJSON, err = json.Marshal(obu)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(idObu, obuJSON)
}

func (s *SmartContract) ObuExists(ctx contractapi.TransactionContextInterface, id, spz, country string) (bool, error) {
	idObu, err := ctx.GetStub().CreateCompositeKey(obuIndex, []string{id, spz, country})	
	if err != nil {
		return false, fmt.Errorf("failed to create composite key: %v", err)

	} 

	obuJSON, err := ctx.GetStub().GetState(idObu)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return obuJSON != nil, nil
}


func (s *SmartContract) GetAllObus(ctx contractapi.TransactionContextInterface) ([]*OnBoardUnit, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all obuList in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(obuIndex, []string{})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var obuList []*OnBoardUnit
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var obu OnBoardUnit
		err = json.Unmarshal(queryResponse.Value, &obu)
		if err != nil {
			return nil, err
		}
		obuList = append(obuList, &obu)
	}

	return obuList, nil
}

