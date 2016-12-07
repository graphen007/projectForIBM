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
	//"bytes"
	//"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"runtime"
)

var logger = shim.NewLogger("BTChaincode")

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
	TimeStamp   string `json:"timeStamp"`
	Name        string `json:"name"`
	CPR         string `json:"CPR"`
	Doctor      string `json:"doctor"`
	Hospital    string `json:"hospital"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	BloodTestID string `json:"bloodTestID"`
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
const ADMIN_INDEX = "adminIndex"
const DOCTOR_INDEX = "doctorIndex"
const CLIENT_INDEX = "clientIndex"
const HOSPITAL_INDEX = "hospitalIndex"
const BLOODBANK_INDEX = "bloodbankIndex"
const COLUMN_CERTS = "eCerts"
const COLUMN_VALUE = "value"

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

	// Create tables
	t.CreateTables(stub)

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
	} else if function == "get_enrollment_cert" {
		return t.get_enrollment_cert(stub, args)
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
	var err error

	fmt.Println("- start set status")
	fmt.Println(args[0] + " - " + args[1])
	marbleAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := bloodTest{}
	json.Unmarshal(marbleAsBytes, &res)
	fmt.Println(res)
	res.Status = args[1] //change the user
	fmt.Println(res)
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes) //rewrite the marble with id as key
	if err != nil {
		return nil, err
	}

	fmt.Println("- end set status")
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

	fmt.Println("changing doctor")
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	fmt.Println("creating list")
	var bloodInd []string

	fmt.Println("Unmarshaling doctor")
	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil {
		fmt.Println("you dun goofed")
	}
	fmt.Println("assigning to res")
	res := bloodTest{}
	var bloodAsBytes []byte
	fmt.Println("running through list")
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
	   "bloodTestID", "Hospital"
	   -------------------------------------------------------
	*/
	hospital := args[1]
	fmt.Println("it might actually work now")
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
			fmt.Println("found it")
			res.Hospital = hospital
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

		return nil, errors.New("This blood test already exists")
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
	      0  		1		       2	       3			4
	   	"ecert"	"typeOfUser"   "username"  "password"  "accesstoken"
	   -------------------------------------------------------
	*/

	fmt.Println("Creating the account")
	if len(args) != 4 {
		return nil, errors.New("Gimme more arguments, 4 to be exact, User and number pliz")
	}

	typeOfUser := args[1]
	username := args[2]
	password := args[3]
	ecert := args[0]

	accountAsBytes, err := stub.GetState(username)
	if err != nil {
		return nil, errors.New("")
	}
	res := account{}
	fmt.Println("checking if account exists")
	json.Unmarshal(accountAsBytes, &res)
	if res.Username == username {
		fmt.Println("This account already exists")
		return nil, errors.New("This account already exists")
	}

	// Check access token
	fmt.Println("getting AccesToken")
	accessCode, err := t.CheckToken(args[4])
	if err != nil {
		fmt.Println("Failed: during token approval")
		return nil, errors.New("Failed: during token approval")
	}

	if len(ecert) == 0 {
		fmt.Println("Caller has no eCert!")
		return nil, errors.New("Caller has no eCert!")
	}

	// Set account permissons
	// ADMIN | DOCTOR | CLIENT | HOSPITAL | BLOODBANK
	fmt.Println("Checking the permission")
	switch typeOfUser {
	case ADMIN:
		fmt.Println("It's an Admin ACC")
		// Check access code
		if accessCode != 0 {
			fmt.Println("Token does not give admin rights!")
			return nil, errors.New("Token does not give admin rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{ADMIN_INDEX, username, ecert})

		if err != nil {
			return nil, errors.New("SaveECertificate Failed:")
		}
		if ok != 1 {
			return nil, errors.New("SaveECertificate Failed")
		}

	case DOCTOR:
		fmt.Println("It's an doctor ACC")
		// Check access code
		if accessCode != 1 {
			return nil, errors.New("Token does not give doctor rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{DOCTOR_INDEX, username, ecert})

		if err != nil {
			return nil, errors.New("SaveECertificate Failed:")
		}
		if ok != 1 {
			return nil, errors.New("SaveECertificate Failed")
		}

	case CLIENT:
		fmt.Println("It's an CLIENT ACC")
		// Check access code
		if accessCode != 2 {
			return nil, errors.New("Token does not give client rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{DOCTOR_INDEX, username, ecert})

		if err != nil {
			return nil, errors.New("SaveECertificate Failed:")
		}
		if ok != 1 {
			return nil, errors.New("SaveECertificate Failed")
		}
	case HOSPITAL:
		fmt.Println("It's an Hospital ACC")
		// Check access code
		if accessCode != 3 {
			return nil, errors.New("Token does not give hospital rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{HOSPITAL_INDEX, username, ecert})

		if err != nil {
			return nil, errors.New("SaveECertificate Failed:")
		}
		if ok != 1 {
			return nil, errors.New("SaveECertificate Failed")
		}
	case BLOODBANK:
		fmt.Println("It's an bloodbank ACC")
		// Check access code
		if accessCode != 4 {
			return nil, errors.New("Token does not give blood bank rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{BLOODBANK_INDEX, username, ecert})

		if err != nil {
			return nil, errors.New("SaveECertificate Failed:")
		}
		if ok != 1 {
			return nil, errors.New("SaveECertificate Failed")
		}

	default:
		fmt.Println("User not supported. User has not been created!")
		return nil, errors.New("User not supported. User has not been created!")
	}

	json.Unmarshal(accountAsBytes, &res)

	stringss := `{"typeOfUser": "` + typeOfUser + `", "username": "` + username + `", "password": "` + password + `"}` //build the Json element
	err = stub.PutState(username, []byte(stringss))
	if err != nil {
		fmt.Println("Could not add account to list")
		return nil, err
	}

	//get the account index
	accountAsBytes, err = stub.GetState(accountIndex)
	if err != nil {
		fmt.Println("Could not get acc index")
		return nil, errors.New("you fucked up")
	}

	var accInd []string
	json.Unmarshal(accountAsBytes, &accInd)
	fmt.Println("Appending to List")
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

	/*
	   Our model looks like
	   -------------------------------------------------------
	       0         1	         2
	   	"ecert"  "username"  "password"
	   -------------------------------------------------------
	*/

	if len(args) != 3 {
		return nil, errors.New("Gimme more arguments, 3 to be exact")
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

	if t.CheckRole(stub, args[1], ADMIN_INDEX, args[0]) != true {
		fmt.Println("Access Denied!")
		return nil, errors.New("Access Denied!")
	}

	var accountAsBytes []byte
	var finalListForUser []byte = []byte(`"returnedObjects":[`)
	res := account{}
	for i := range userIndex {

		accountAsBytes, err = stub.GetState(userIndex[i])
		json.Unmarshal(accountAsBytes, &res)
		if res.Username == args[1] && res.Password == args[2] {

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
// Get ECERT HOLDER
// ============================================================================================================================
func (t *SimpleChaincode) get_enrollment_cert(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	       0         1
	   "tablename"  "key"
	   -------------------------------------------------------
	*/

	if len(args) != 2 {
		fmt.Println("At least 2 args must be provided\n")
		return nil, errors.New("Get Admin Ecerts failed. Must include at least 2 args value")
	}

	// Initial repsonse
	var finalList []byte = []byte(`"ecert":[`)

	// Getting the rows for admin
	var columns []shim.Column
	var tmpHolder []byte

	fmt.Println("Finding key/value pair key for key: ", args[1])
	colNext := shim.Column{Value: &shim.Column_String_{String_: args[1]}}
	columns = append(columns, colNext)

	row, err := stub.GetRow(args[0], columns)

	if err != nil {
		fmt.Println("Failed getting row for ", args[0])
		return nil, errors.New("Failed getting row")
	}

	if len(row.GetColumns()) != 0 {
		fmt.Println("Appending eCert")
		tmpHolder = append(tmpHolder, row.Columns[1].GetBytes()...)
	}

	finalList = append(finalList, tmpHolder...)
	finalList = append(finalList, []byte(`]`)...)

	fmt.Println("End of get_enrollment_cert!")

	return finalList, nil

}

// ============================================================================================================================
// CheckToken - The args[3] should contain the token of user type
// ============================================================================================================================
func (t *SimpleChaincode) CheckToken(token string) (int, error) {

	fmt.Println("checking token")
	if len(token) == 0 {
		fmt.Println("Invalid token. Empty.")
		return -1, errors.New("Invalid token. Empty.")
	}

	switch token {
	case ADMIN_TOKEN:
		fmt.Println("return 0")
		return 0, nil
	case DOCTOR_TOKEN:
		fmt.Println("return 1")
		return 1, nil
	case CLIENT_TOKEN:
		fmt.Println("return 2")
		return 2, nil
	case HOSPITAL_TOKEN:
		fmt.Println("return 3")
		return 3, nil
	case BLOODBANK:
		fmt.Println("return 4")
		return 4, nil
	default:
		fmt.Println("Invalid token. Not Correct.")
		return -1, errors.New("Invalid token. Not Correct.")

	}
}

// ============================================================================================================================
// SaveECertificate - Save Callers Enrollment Certificate
// ============================================================================================================================
func (t *SimpleChaincode) SaveECertificate(stub shim.ChaincodeStubInterface, args []string) (int, error) {

	// args: 0 = tablename, 1 = username, 2 = ecert

	fmt.Println("SaveECertificate")

	if len(args) != 3 {
		fmt.Println("Invaild number of arguments - 0 = tablename, 1 = username, 2 = ecert")
		return -1, errors.New("Invaild number of arguments ")
	}

	// Saving Ecert
	logger.Debug("Saving to table: ", args[0])

	// *Debugging*
	logger.Debug("Peer ecert: ", args[2])

	// Inserting rows
	fmt.Println("Inserting user: ", args[1])
	ok, err := stub.InsertRow(args[0], shim.Row{
		Columns: []*shim.Column{
			&shim.Column{Value: &shim.Column_String_{String_: args[1]}},
			&shim.Column{Value: &shim.Column_String_{String_: args[2]}}},
	})

	if err != nil {
		fmt.Println("Error: can't insert row! ", err)
		return -1, errors.New("Error: can't insert row!")
	} else if !ok {
		fmt.Println("Failed inserting row!")
		return -1, nil
	}

	fmt.Println("Insert successful!")

	fmt.Println("Checking for inserted data!")
	var columns []shim.Column
	colNext := shim.Column{Value: &shim.Column_String_{String_: args[1]}}
	columns = append(columns, colNext)

	row, err := stub.GetRow(args[0], columns)
	if err != nil {
		fmt.Println("Failed inserted row for ", args[0])
		return -1, errors.New("Failed getting rows")
	}

	if len(row.GetColumns()) != 0 {
		logger.Debug("Retrived ecert from table: [%x]", row.Columns[1].GetString_())
	}

	// Ran successfully!
	return 1, nil
}

// ============================================================================================================================
// CreateTables - Called from init
// ============================================================================================================================
func (t *SimpleChaincode) CreateTables(stub shim.ChaincodeStubInterface) {

	var tableName string
	for i := 0; i < 5; i++ {

		switch i {
		case 1:
			tableName = DOCTOR_INDEX
		case 2:
			tableName = HOSPITAL_INDEX
		case 3:
			tableName = CLIENT_INDEX
		case 4:
			tableName = BLOODBANK_INDEX
		default:
			tableName = ADMIN_INDEX
		}

		fmt.Println("Creating table: ", tableName)

		err := stub.CreateTable(tableName, []*shim.ColumnDefinition{
			&shim.ColumnDefinition{Name: COLUMN_CERTS, Type: shim.ColumnDefinition_STRING, Key: true},
			&shim.ColumnDefinition{Name: COLUMN_VALUE, Type: shim.ColumnDefinition_STRING, Key: false},
		})

		if err != nil {
			fmt.Println("Table is already created! Error: [%s]", err)
		}
	}
}

// ============================================================================================================================
// CheckRole - Called from all invoke/query func's.
// Access Control happens here
// Params: username, role
// Username should be pass from json input
// Roles: ADMIN_INDEX, DOCTOR_INDEX, CLIENT_INDEX, HOSPITAL_INDEX, BLOODBANK_INDEX
// ============================================================================================================================
func (t *SimpleChaincode) CheckRole(stub shim.ChaincodeStubInterface, username string, role string, ecert string) bool {

	fmt.Println("Checking Role")
	fmt.Println("Finding pair in table ", role)
	fmt.Println("Finding key/value pair for key: ", username)

	// Get the row for username
	var columns []shim.Column
	colNext := shim.Column{Value: &shim.Column_String_{String_: username}}
	columns = append(columns, colNext)

	row, err := stub.GetRow(role, columns)

	if err != nil {
		fmt.Println("Access denied! User not permitted to do this: ", role)
		return false
	}

	if len(row.GetColumns()) != 0 {

		fmt.Println("Getting saved callerCertificate")
		ecertSaved := row.Columns[1].GetString_()

		// Compare callers ecert & that which is stored
		fmt.Println("Checking signature")

		fmt.Print("\nIn table: ", ecertSaved)
		fmt.Print("\nIn Signature: ", ecert)

		if ecertSaved != ecert {
			fmt.Println("Access denied!")
			return false
		} else {
			fmt.Println("x509Cert signature matches!")
			return true
		}
	}

	fmt.Println("Access denied! Last in function")
	return false
}
