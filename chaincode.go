package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"os"
	"strconv"
)

var logger = shim.NewLogger("fabric-boilerplate")

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	SimpleChaincode - A blank struct for use with Shim (An IBM Blockchain included go file used for get/put state
//					  and other IBM Blockchain functions)
//==============================================================================================================================
type SimpleChaincode struct {
}

//==============================================================================================================================
//	Index string & Constants
//==============================================================================================================================

var companiesIndexStr = "_companies"
var mappCodesIndexStr = "_mappCodes"

var indexes = []string{companiesIndexStr, mappCodesIndexStr}

//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Passes the
//  		 initial arguments passed are passed on to the called function.
//==============================================================================================================================

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	logger.Infof("Invoke is running " + function)

	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "reset_indexes" {
		return t.reset_indexes(stub, args)
	} else if function == "add_company" {
		return t.add_company(stub, args)
	}

	return nil, errors.New("Received unknown invoke function name")
}

//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	logger.Infof("Query is running " + function)

	if function == "get_company" {
		return t.get_company(stub, args)
	} else if function == "get_all_companies" {
		return t.get_all_companies(stub, args)
	}

	return nil, errors.New("Received unknown query function name")
}

//=================================================================================================================================
//  Main - main - Starts up the chaincode
//=================================================================================================================================

func main() {

	// LogDebug, LogInfo, LogNotice, LogWarning, LogError, LogCritical (Default: LogDebug)
	logger.SetLevel(shim.LogInfo)

	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s", err)
	}
}

//==============================================================================================================================
//  Init Function - Called when the user deploys the chaincode
//==============================================================================================================================

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	return nil, nil
}

//==============================================================================================================================
//  Utility Functions
//==============================================================================================================================

//==============================================================================================================================
//  Invoke Functions
//==============================================================================================================================
func (t *SimpleChaincode) reset_indexes(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	for _, i := range indexes {
		// Marshal the index
		var emptyIndex []string

		empty, err := json.Marshal(emptyIndex)
		if err != nil {
			return nil, errors.New("Error marshalling")
		}
		err = stub.PutState(i, empty)

		if err != nil {
			return nil, errors.New("Error deleting index")
		}
		logger.Infof("Delete with success from ledger: " + i)
	}
	return nil, nil
}

func append_id(stub *shim.ChaincodeStub, indexStr string, id string, create bool) ([]byte, error) {

	indexAsBytes, err := stub.GetState(indexStr)
	if err != nil {
		return nil, errors.New("Failed to get " + indexStr)
	}

	// Unmarshal the index
	var tmpIndex []string
	json.Unmarshal(indexAsBytes, &tmpIndex)

	// Create new id
	var newId = id
	if create {
		newId += strconv.Itoa(len(tmpIndex) + 1)
	}

	// append the new id to the index
	tmpIndex = append(tmpIndex, newId)

	jsonAsBytes, _ := json.Marshal(tmpIndex)
	err = stub.PutState(indexStr, jsonAsBytes)
	if err != nil {
		return nil, errors.New("Error storing new " + indexStr + " into ledger")
	}

	return []byte(newId), nil

}

func (t *SimpleChaincode) add_company(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//Args
	//			0				1
	//		  index		company JSON object (as string)

	id, err := append_id(stub, companiesIndexStr, args[0], false)
	if err != nil {
		return nil, errors.New("Error creating new id for user " + args[0])
	}

	err = stub.PutState(string(id), []byte(args[1]))
	if err != nil {
		return nil, errors.New("Error putting user data on ledger")
	}

	return nil, nil
}

//==============================================================================================================================
//		Query Functions
//==============================================================================================================================

func (t *SimpleChaincode) get_company(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	bytes, err := stub.GetState(args[0])

	if err != nil {
		return nil, errors.New("Could not retrieve information for the given company ID")
	}

	return bytes, nil

}

func (t *SimpleChaincode) get_all_companies(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	indexAsBytes, err := stub.GetState(companiesIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get " + companiesIndexStr)
	}

	// Unmarshal the index
	var companyIndex []string
	json.Unmarshal(indexAsBytes, &companyIndex)

	var companies []Company
	for _, company := range companyIndex {

		bytes, err := stub.GetState(company)
		if err != nil {
			return nil, errors.New("Unable to get thing with ID: " + company)
		}

		var c Company
		json.Unmarshal(bytes, &c)
		companies = append(companies, c)
	}

	companiesAsJsonBytes, _ := json.Marshal(companies)
	if err != nil {
		return nil, errors.New("Could not convert things to JSON ")
	}

	return companiesAsJsonBytes, nil
}
