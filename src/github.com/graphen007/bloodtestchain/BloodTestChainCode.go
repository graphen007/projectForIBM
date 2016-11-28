/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
//commit this bullshit
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"runtime"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we will use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
const ADMIN = "admin"         // 0
const DOCTOR = "doctor"       // 1
const CLIENT = "client"       // 2
const HOSPITAL = "hospital"   // 3
const BLOODBANK = "bloodbank" // 4

//==============================================================================================================================
//	 Hardcoded access tokens
//==============================================================================================================================
const ADMIN_TOKEN = "pNAQvsgTSz"
const DOCTOR_TOKEN = "9Hk5e3rdR9"
const CLIENT_TOKEN = "ERE4zwMnao"
const HOSPITAL_TOKEN = "XpK9cGH22x"
const BLOODBANK_TOKEN = "TdFeAzGlrI"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//name for the key/value that will store a list of all known tests/accounts
var bloodTestIndex = "_bloodTestIndex"
var accountIndex = "_accountIndex"

type bloodTest struct {
	TimeStamp   string `json:"timestamp"`
	Name        string `json:"name"`
	CPR         string `json:"CPR"`
	Doctor      string `json:"doctor"`
	Hospital    string `json:"hospital"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	BloodTestID string `json:"BloodTestID"`
}

//==============================================================================================================================
//	account - Struct for storing the JSON of a account
//==============================================================================================================================
type account struct {
	TypeOfUser string `json:"typeOfUser"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================
type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert    string `json:"ecert"`
}

//name for the key/value that will store a list of all known permissionholders
var adminIndex = "_adminIndex"
var doctorIndex = "_doctorIndex"
var clientIndex = "_clientIndex"
var hospitalIndex = "_hospitalIndex"
var bloodbankIndex = "_bloodbankIndex"

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {

	// maximize CPU usage for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	// *shim.ChaincodeStub 0.5
	// shim.ChaincodeStubInterface 0.6
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point to invoke a chaincode function
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "init_bloodtest" {
		return t.init_bloodtest(stub, args)
	} else if function == "change_status" {
		return t.change_status(stub, args)
	} else if function == "change_doctor" {
		return t.change_doctor(stub, args)
	} else if function == "change_hospital" {
		return t.change_hospital(stub, args)
	} else if function == "change_result" {
		return t.change_result(stub, args)
	} else if function == "create_user" {
		return t.create_user(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	//for 0.6 stub shim.ChaincodeStubInterface

	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	} else if function == "read_list" {
		return t.read_list(stub, args)
	} else if function == "patient_read" {
		return t.patient_read(stub, args)
	} else if function == "doctor_read" {
		return t.doctor_read(stub, args)
	} else if function == "hospital_read" {
		return t.hospital_read(stub, args)
	} else if function == "get_user" {
		return t.get_user(stub, args)
	} else if function == "get_admin_certs" {
		return t.get_admin_certs(stub, args)
	}

	fmt.Println("query did not find func: " + function)
	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]
	value = args[1]
	err = stub.PutState(name, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// ============================================================================================================================
// Patient Read
// ============================================================================================================================
func (t *SimpleChaincode) patient_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte = []byte(`"returnedObjects":[`)
	res := bloodTest{}
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.CPR == args[0] {

			finalList = append(finalList, bloodAsBytes...)
			if i < (len(bloodInd) - 1) {
				finalList = append(finalList, []byte(`,`)...)
			}
		}
	}
	finalList = append(finalList, []byte(`]`)...)

	return finalList, nil
}

// ============================================================================================================================
// Doctor Read
// ============================================================================================================================
func (t *SimpleChaincode) doctor_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte = []byte(`"returnedObjects":[`)
	res := bloodTest{}
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.Doctor == args[0] {

			finalList = append(finalList, bloodAsBytes...)
			if i < (len(bloodInd) - 1) {
				finalList = append(finalList, []byte(`,`)...)
			}
		}
	}
	finalList = append(finalList, []byte(`]`)...)

	return finalList, nil
}

// ============================================================================================================================
// Hospital Read
// ============================================================================================================================
func (t *SimpleChaincode) hospital_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte = []byte(`"returnedObjects":[`)
	res := bloodTest{}
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.Hospital == args[0] {

			finalList = append(finalList, bloodAsBytes...)
			if i < (len(bloodInd) - 1) {
				finalList = append(finalList, []byte(`,`)...)
			}
		}
	}
	finalList = append(finalList, []byte(`]`)...)

	return finalList, nil
}

// ============================================================================================================================
// bloodbank Read !! HAS NOT BEEN ADDED YET AND IS NOT FULLY FUNCTIONAL!!!
// ============================================================================================================================
func (t *SimpleChaincode) bloodbank_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte = []byte(`"returnedObjects":[`)
	res := bloodTest{}
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.Result == args[0] {

			finalList = append(finalList, bloodAsBytes...)
			if i < (len(bloodInd) - 1) {
				finalList = append(finalList, []byte(`,`)...)
			}
		}
	}
	finalList = append(finalList, []byte(`]`)...)

	return finalList, nil
}

// ============================================================================================================================
// Read list
// ============================================================================================================================
func (t *SimpleChaincode) read_list(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	var finalList []byte
	var bloodAsBytes []byte
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		finalList = append(finalList, bloodAsBytes...)

	}

	return finalList, nil
}

// ============================================================================================================================
// Change Status
// ============================================================================================================================
func (t *SimpleChaincode) change_status(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	      0              1
	   "bloodTestID", "Status"
	   -------------------------------------------------------
	*/

	if len(args) != 2 {
		return nil, errors.New("Gimme more arguments, 2 to be exact, ID and status")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	res := bloodTest{}
	var bloodAsBytes []byte
	for i := range bloodInd {
		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.BloodTestID == args[0] {
			res.Status = args[1]
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes) //rewrite the bloodtest with id as key
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

// ============================================================================================================================
// Change Doctor
// ============================================================================================================================
func (t *SimpleChaincode) change_doctor(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	      0              1
	   "bloodTestID", "Status"
	   -------------------------------------------------------
	*/

	if len(args) != 2 {
		return nil, errors.New("Gimme more arguments, 2 to be exact, ID and status")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	res := bloodTest{}
	var bloodAsBytes []byte
	for i := range bloodInd {
		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.BloodTestID == args[0] {
			res.Doctor = args[1]
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes) //rewrite the blodtest with id as key
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// ============================================================================================================================
// Change Hospital
// ============================================================================================================================
func (t *SimpleChaincode) change_hospital(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	      0              1
	   "bloodTestID", "Status"
	   -------------------------------------------------------
	*/

	if len(args) != 2 {
		return nil, errors.New("Gimme more arguments, 2 to be exact, ID and status")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	res := bloodTest{}
	var bloodAsBytes []byte
	for i := range bloodInd {
		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.BloodTestID == args[0] {
			res.Hospital = args[1]
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes) //rewrite the marble with id as key
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// ============================================================================================================================
// Change Result
// ============================================================================================================================
func (t *SimpleChaincode) change_result(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	/*
	   Our model looks like
	   -------------------------------------------------------
	      0              1
	   "bloodTestID", "Status"
	   -------------------------------------------------------
	*/
	if len(args) != 2 {
		return nil, errors.New("Gimme more arguments, 2 to be exact, ID and status")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	res := bloodTest{}
	var bloodAsBytes []byte
	for i := range bloodInd {
		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.BloodTestID == args[0] {
			res.Result = args[1]
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes) //rewrite the marble with id as key
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

// ============================================================================================================================
// Init Bloodtest
// ============================================================================================================================
func (t *SimpleChaincode) init_bloodtest(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	/*
	   Our model looks like
	   -------------------------------------------------------
	   -------------------------------------------------------
	         0         1      2        3        4	       5	    6	       7
	   "timestamp", "name", "CPR", "doctor", "hospital" "status" "result" "bloodTestID
	   -------------------------------------------------------
	*/

	fmt.Println("Creating the bloodTest")
	if len(args) != 8 {
		return nil, errors.New("Gimme more arguments, 8 to be exact, User and number pliz")
	}

	timeStamp := args[0]
	name := args[1]
	CPR := args[2]
	doctor := args[3]
	hospital := args[4]
	status := args[5]
	result := args[6]
	bloodTestID := args[7]

	bloodAsBytes, err := stub.GetState(bloodTestID)
	if err != nil {
		return nil, errors.New("blood")
	}
	res := bloodTest{}
	json.Unmarshal(bloodAsBytes, &res)
	if res.BloodTestID == bloodTestID {

		return nil, errors.New("This blood test arleady exists")
	}

	json.Unmarshal(bloodAsBytes, &res)

	// "STILL TESTING! timeStamp": " + timeStamp + ", "name": " + name + ", "CPR": " + CPR + ", "doctor": " + doctor + ", "hospital": " + hospital + ", "status": " + status + ", "result": " + result + ", "bloodTestID": " + bloodTestID + "

	str := []byte(`{"timeStamp": "` + timeStamp + `","name": "` + name + `","CPR": "` + CPR + `","doctor": "` + doctor + `","hospital": "` + hospital + `","status": "` + status + `","result": "` + result + `","bloodTestID": "` + bloodTestID + `"}`) //build the Json element

	err = stub.PutState(bloodTestID, str)
	if err != nil {
		return nil, err
	}

	//get the blood index
	bloodAsBytes, err = stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("you fucked up")
	}

	var bloodInd []string
	json.Unmarshal(bloodAsBytes, &bloodInd)

	//append it to the list
	bloodInd = append(bloodInd, bloodTestID)
	jsonAsBytes, _ := json.Marshal(bloodInd)
	err = stub.PutState(bloodTestIndex, jsonAsBytes)

	fmt.Println("Ended of creation")

	return nil, nil
}

// ============================================================================================================================
// Create User - user account stuff, login, create user, etc.
// ============================================================================================================================
func (t *SimpleChaincode) create_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	/*
	   Our model looks like
	   -------------------------------------------------------
	        0		       1	       2
	   "typeOfUser"   "username"  "password"
	   -------------------------------------------------------
	*/

	fmt.Println("Creating the account")
	if len(args) != 3 {
		return nil, errors.New("Gimme more arguments, 3 to be exact, User and number pliz")
	}

	typeOfUser := args[0]
	username := args[1]
	password := args[2]

	accountAsBytes, err := stub.GetState(username)
	if err != nil {
		return nil, errors.New("")
	}
	res := account{}
	json.Unmarshal(accountAsBytes, &res)
	if res.Username == username {
		return nil, errors.New("This account arleady exists")
	}

	// Check access token
	accessCode, err := CheckToken(stub)
	if err != nil {
		return nil, errors.New("Failed during token approval")
	}

	ecert, err := stub.GetCallerCertificate()
	if err != nil {
		return nil, errors.New("Failed during ecert retrival")
	}

	// Convert byte[] to str
	ecertStr := string(ecert[:])

	// Set account permissons
	// ADMIN | DOCTOR | CLIENT | HOSPITAL | BLOODBANK
	switch typeOfUser {
	case ADMIN:

		// Check access code
		if accessCode != 0 {
			return nil, errors.New("Token does not give admin rights!")
		}

		// Get holder
		adminAsBytes, err := stub.GetState(adminIndex)
		if err != nil {
			return nil, errors.New("Failed getting adminIndex")
		}

		// Create tmp
		var tmpHolder []string
		json.Unmarshal(adminAsBytes, &tmpHolder)

		// Append this users eCert to the list
		tmpHolder = append(tmpHolder, ecertStr)
		jsonAsBytes, _ := json.Marshal(tmpHolder)
		err = stub.PutState(adminIndex, jsonAsBytes)
	case DOCTOR:

		// Check access code
		if accessCode != 1 {
			return nil, errors.New("Token does not give doctor rights!")
		}

		// Get holder
		doctorAsBytes, err := stub.GetState(doctorIndex)
		if err != nil {
			return nil, errors.New("Failed getting doctorIndex")
		}

		// Create tmp
		var tmpHolder []string
		json.Unmarshal(doctorAsBytes, &tmpHolder)

		// Append this users eCert to the list
		tmpHolder = append(tmpHolder, ecertStr)
		jsonAsBytes, _ := json.Marshal(tmpHolder)
		err = stub.PutState(doctorIndex, jsonAsBytes)
	case CLIENT:

		// Check access code
		if accessCode != 2 {
			return nil, errors.New("Token does not give client rights!")
		}

		// Get holder
		clientAsBytes, err := stub.GetState(clientIndex)
		if err != nil {
			return nil, errors.New("Failed getting doctorIndex")
		}

		// Create tmp
		var tmpHolder []string
		json.Unmarshal(clientAsBytes, &tmpHolder)

		// Append this users eCert to the list
		tmpHolder = append(tmpHolder, ecertStr)
		jsonAsBytes, _ := json.Marshal(tmpHolder)
		err = stub.PutState(clientIndex, jsonAsBytes)
	case HOSPITAL:

		// Check access code
		if accessCode != 3 {
			return nil, errors.New("Token does not give hospital rights!")
		}

		// Get holder
		hospitalAsBytes, err := stub.GetState(hospitalIndex)
		if err != nil {
			return nil, errors.New("Failed getting hospitalIndex")
		}

		// Create tmp
		var tmpHolder []string
		json.Unmarshal(hospitalAsBytes, &tmpHolder)

		// Append this users eCert to the list
		tmpHolder = append(tmpHolder, ecertStr)
		jsonAsBytes, _ := json.Marshal(tmpHolder)
		err = stub.PutState(hospitalIndex, jsonAsBytes)
	case BLOODBANK:

		// Check access code
		if accessCode != 4 {
			return nil, errors.New("Token does not give blood bank rights!")
		}

		// Get holder
		bloodBankAsBytes, err := stub.GetState(bloodbankIndex)
		if err != nil {
			return nil, errors.New("Failed getting doctorIndex")
		}

		// Create tmp
		var tmpHolder []string
		json.Unmarshal(bloodBankAsBytes, &tmpHolder)

		// Append this users eCert to the list
		tmpHolder = append(tmpHolder, ecertStr)
		jsonAsBytes, _ := json.Marshal(tmpHolder)
		err = stub.PutState(bloodbankIndex, jsonAsBytes)

	default:
		return nil, errors.New("User not supported. User has not been created!")
	}

	json.Unmarshal(accountAsBytes, &res)

	stringss := `{"typeOfUser": "` + typeOfUser + `", "username": "` + username + `", "password": "` + password + `"}` //build the Json element
	err = stub.PutState(username, []byte(stringss))
	if err != nil {
		return nil, err
	}

	//get the account index
	accountAsBytes, err = stub.GetState(accountIndex)
	if err != nil {
		return nil, errors.New("you fucked up")
	}

	var accInd []string
	json.Unmarshal(accountAsBytes, &accInd)

	//append it to the list
	accInd = append(accInd, username)
	jsonAsBytes, _ := json.Marshal(accInd)
	err = stub.PutState(accountIndex, jsonAsBytes)

	fmt.Println("Ended of creation")

	return nil, nil
}

// ============================================================================================================================
// Get User
// ============================================================================================================================
func (t *SimpleChaincode) get_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Gimme more arguments, 2 to be exact")
	}
	userList, err := stub.GetState(accountIndex)
	if err != nil {
		return nil, errors.New("Failed to get accountList")
	}
	var userIndex []string

	err = json.Unmarshal(userList, &userIndex)
	if err != nil {
		fmt.Println("you dun goofed")
	}

	var accountAsBytes []byte
	var finalListForUser []byte = []byte(`"returnedObjects":[`)
	res := account{}
	for i := range userIndex {

		accountAsBytes, err = stub.GetState(userIndex[i])
		json.Unmarshal(accountAsBytes, &res)
		if res.Username == args[0] && res.Password == args[1] {

			finalListForUser = append(finalListForUser, accountAsBytes...)
			if i < (len(userIndex) - 1) {
				finalListForUser = append(finalListForUser, []byte(`,`)...)
			}
		}
	}
	finalListForUser = append(finalListForUser, []byte(`]`)...)

	return finalListForUser, nil

}

// ============================================================================================================================
// Get ADMIN CERT HOLDER
// ============================================================================================================================
func (t *SimpleChaincode) get_admin_certs(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	adminCerts, err := stub.GetState(adminIndex)
	if err != nil {
		return nil, errors.New("Failed to get adminEcertList")
	}

	var finalList []byte = []byte(`"admin_ecerts":[`)

	finalList = append(finalList, adminCerts...)

	finalList = append(finalList, []byte(`]`)...)

	return finalListForUser, nil

}

// ============================================================================================================================
// CheckToken - The metadata should contain the token of user type
// ============================================================================================================================
func CheckToken(stub shim.ChaincodeStubInterface) (int, error) {

	token, err := stub.GetCallerMetadata()
	if err != nil {
		return -1, errors.New("Failed getting metadata.")
	}
	if len(token) == 0 {
		return -1, errors.New("Invalid token. Empty.")
	}

	tokenStr := string(token[:])

	switch tokenStr {
	case ADMIN_TOKEN:
		return 0, nil
	case DOCTOR_TOKEN:
		return 1, nil
	case CLIENT_TOKEN:
		return 2, nil
	case HOSPITAL_TOKEN:
		return 3, nil
	case BLOODBANK:
		return 4, nil
	default:
		return -1, errors.New("Invalid token. Not Correct.")

	}
}
