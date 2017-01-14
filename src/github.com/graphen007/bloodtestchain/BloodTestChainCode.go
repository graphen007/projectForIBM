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
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"runtime"
	"strconv"
	"time"
)

var logger = shim.NewLogger("BTChaincode")

//==============================================================================================================================
// Participant types - Each participant type is mapped to an integer which we will use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
const ADMIN = "admin"         // 0
const DOCTOR = "doctor"       // 1
const CLIENT = "client"       // 2
const HOSPITAL = "hospital"   // 3
const LAB = "lab" 	      // 4

//==============================================================================================================================
// Hardcoded access tokens
//==============================================================================================================================
const ADMIN_TOKEN = "pNAQvsgTSz"
const DOCTOR_TOKEN = "9Hk5e3rdR9"
const CLIENT_TOKEN = "ERE4zwMnao"
const HOSPITAL_TOKEN = "XpK9cGH22x"
const LAB_TOKEN = "TdFeAzGlrI"

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//name for the key/value that will store a list of all known tests/accounts
var bloodTestIndex = "_bloodTestIndex"
var accountIndex = "_accountIndex"

type bloodTest struct {
	TimeStampDoctor   string `json:"timeStampDoctor"`
	TimeStampHospital   string `json:"timeStampHospital"`
	TimeStampLab   string `json:"timeStampLab"`
	TimeStampAnalyse   string `json:"timeStampAnalyse"`
	TimeStampResult   string `json:"timeStampResult"`
	Name        string `json:"name"`
	CPR         string `json:"CPR"`
	Doctor      string `json:"doctor"`
	Hospital    string `json:"hospital"`
	Lab	    string `json:"lab"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	BloodTestID string `json:"bloodTestID"`
}

//==============================================================================================================================
// account - Struct for storing the JSON of a account
//==============================================================================================================================
type account struct {
	TypeOfUser string `json:"typeOfUser"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

//==============================================================================================================================
// Name for the key/value that will store a list of all known permissionholders
//==============================================================================================================================
const ADMIN_INDEX = "adminIndex"
const DOCTOR_INDEX = "doctorIndex"
const CLIENT_INDEX = "clientIndex"
const HOSPITAL_INDEX = "hospitalIndex"
const LAB_INDEX = "labIndex"
const COLUMN_CERTS = "eCerts"
const COLUMN_VALUE = "value"

// ============================================================================================================================
// Main
// ============================================================================================================================
func main(){

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
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){
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
	}else if function == "change_lab" {
		return t.change_lab(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	} else if function == "read_list" {
		return t.read_list(stub, args)
	} else if function == "client_read" {
		return t.client_read(stub, args)
	} else if function == "doctor_read" {
		return t.doctor_read(stub, args)
	} else if function == "hospital_read" {
		return t.hospital_read(stub, args)
	} else if function == "lab_read" {
		return t.lab_read(stub, args)
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
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error){
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
// Client Read
// ============================================================================================================================
func (t *SimpleChaincode) client_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

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
		fmt.Println("you dun goofed", err)
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

	fmt.Println("doctor_read started")

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
		fmt.Println("bloodTestList not array of strings")
	}


	type doctorReadStruct struct {
		BloodTestList []bloodTest `json:"returnedObjects"`
	}


	var bloodAsBytes []byte
	bloodTestListStruct := doctorReadStruct{}
	bloodTest := bloodTest{}
	for i := range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		if err != nil {
			fmt.Println("could not get state")
		}
		err = json.Unmarshal(bloodAsBytes, &bloodTest)
		if err != nil {
			fmt.Println("bloodtest does not match doctorReadStruct")
		}
		if bloodTest.Doctor == args[0] {

			bloodTestListStruct.BloodTestList = append(bloodTestListStruct.BloodTestList, bloodTest)

		}
	}

	var finalList []byte;
	finalList, err = json.Marshal(bloodTestListStruct)
	fmt.Println(finalList)
	return finalList, err
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
// lab Read !! HAS NOT BEEN ADDED YET AND IS NOT FULLY FUNCTIONAL!!!
// ============================================================================================================================
func (t *SimpleChaincode) lab_read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

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
		if res.Lab == args[0] {

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

	// Change the user
	res.Status = args[1]
	if args[1] == "Analysing" {
	t:= time.Now()
	t.Format("20060102150405")
	res.TimeStampAnalyse = t.String()

	}
	fmt.Println(res)
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)
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
			t := time.Now()
			t.Format("20060102150405")
			res.TimeStampDoctor = t.String()
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}



// ============================================================================================================================
// Change Lab
// ============================================================================================================================
func (t *SimpleChaincode) change_lab(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	      0              1
	   "bloodTestID", "Lab"
	   -------------------------------------------------------
	*/

	fmt.Println("changing lab")
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	fmt.Println("creating list")
	var bloodInd []string

	fmt.Println("Unmarshaling lab")
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
			res.Lab = args[1]
			t := time.Now()
			t.Format("20060102150405")
			res.TimeStampLab = t.String()
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes)
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
			t := time.Now()
			t.Format("20060102150405")
			res.TimeStampHospital = t.String()
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes)
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
			t := time.Now()
			t.Format("20060102150405")
			res.TimeStampResult = t.String()
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes)
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

	str := []byte(`{"timeStampDoctor": "` + timeStamp  + `","timeStampHospital": "` +`null` + `","timeStampLab": "` + `null`  + `","timeStampAnalyse": "` + `null`  + `","timeStampResult": "` + `null`+ `","name": "` + name + `","CPR": "` + CPR + `","doctor": "` + doctor + `","hospital": "` + hospital + `","lab": "` + `unassigned` + `","status": "` + status + `","result": "` + result + `","bloodTestID": "` + bloodTestID + `"}`) //build the Json element

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
	      0  	     1		   2	       3	     4
	   "ecert"	"typeOfUser"   "username"  "password"  "accesstoken"
	   -------------------------------------------------------
	*/

	fmt.Println("Creating the account")
	if len(args) != 5 {
		return nil, errors.New("Gimme more arguments, 5 to be exact, User and number pliz")
	}

	// Create vars
	ecert := args[0]
	typeOfUser := args[1]
	username := args[2]
	password := args[3]

	accountAsBytes, err := stub.GetState(username)
	if err != nil {
		return nil, errors.New("Error getting state for username")
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

	// Set account permission
	// ADMIN | DOCTOR | CLIENT | HOSPITAL | LAB
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
		ok, err := t.SaveECertificate(stub, []string{CLIENT_INDEX, username, ecert})

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
	case LAB:
		fmt.Println("It's an laboratory ACC")
		// Check access code
		if accessCode != 4 {
			return nil, errors.New("Token does not give LAB rights!")
		}

		// Store eCert in table
		ok, err := t.SaveECertificate(stub, []string{LAB_INDEX, username, ecert})

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
// Get User - Retreives a users data
// ============================================================================================================================
func (t *SimpleChaincode) get_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	/*
	   Our model looks like
	   -------------------------------------------------------
	       0         1	       2	     3
	     "ecert"  "username"  "password"   "typeOfUser"
	   -------------------------------------------------------
	*/

	if len(args) != 4 {
		return nil, errors.New("Gimme more arguments, 4 to be exact")
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

	// Example of checking role
	// Note how model looks like and keep it the same!
	// Meaning "ecert" is always args[0]
	if t.CheckRole(stub, args[1], t.GetTable(args[3]), args[0]) != true {
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
// Get ECERT HOLDER - Function to retreive a enrollment certificate stored in tablename and key
// @Params: tablename, key
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
		fmt.Println("At least 2 args must be provided")
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
// CheckToken - The args[4] should contain the token of user type
// ============================================================================================================================
func (t *SimpleChaincode) CheckToken(token string) (int, error) {

	fmt.Println("Checking token")
	if len(token) == 0 {
		fmt.Println("Invalid token. Empty.")
		return -1, errors.New("Invalid token. Empty.")
	}

	switch token {
	case ADMIN_TOKEN:
		fmt.Println("Returned 0")
		return 0, nil
	case DOCTOR_TOKEN:
		fmt.Println("Returned 1")
		return 1, nil
	case CLIENT_TOKEN:
		fmt.Println("Returned 2")
		return 2, nil
	case HOSPITAL_TOKEN:
		fmt.Println("Returned 3")
		return 3, nil
	case LAB_TOKEN:
		fmt.Println("Returned 4")
		return 4, nil
	default:
		fmt.Println("Invalid token. Not Correct.")
		return -1, errors.New("Invalid token. Not Correct.")

	}
}

// ============================================================================================================================
// GetTable - The args[3] should contain the table of user type
// ============================================================================================================================
func (t *SimpleChaincode) GetTable(name string) string {

	fmt.Println("Getting table")

	switch name {
	case ADMIN:
		return ADMIN_INDEX
	case LAB:
		return LAB_INDEX
	case CLIENT:
		return CLIENT_INDEX
	case HOSPITAL:
		return HOSPITAL_INDEX
	default:
		return DOCTOR_INDEX
	}
}

// ============================================================================================================================
// SaveECertificate - Save Callers Enrollment Certificate
// @Params: args[] -> 0 = tablename, 1 = username, 2 = ecert
// ============================================================================================================================
func (t *SimpleChaincode) SaveECertificate(stub shim.ChaincodeStubInterface, args []string) (int, error) {

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

	//------------
	// DEBUGGING
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
	//-------------

	// Ran successfully!
	return 1, nil
}

// ============================================================================================================================
// CreateTables - Called from init
// ============================================================================================================================
func (t *SimpleChaincode) CreateTables(stub shim.ChaincodeStubInterface) {

	fmt.Print("Creating tables...")

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
			tableName = LAB_INDEX
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

	// Creating clients
	fmt.Println("Creating clients")
	for i := 0; i < 100; i++  {

		lastFourSSN := strconv.Itoa(i+1)
		clientEcert := "MIIBoTCCAUegAwIBAgIBATAKBggqhkjOPQQDAzApMQswCQYDVQQGEwJVUzEMMAoGA1UEChMDSUJNMQwwCgYDVQQDEwNlY2EwHhcNMTcwMTEzMTIyMjMwWhcNMTcwNDEzMTIyMjMwWjA5MQswCQYDVQQGEwJVUzEMMAoGA1UEChMDSUJNMRwwGgYDVQQDDBN1c2VyX3R5cGUxXzRcZ3JvdXAxMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEDlr4qaGjUkt+dJK6vUGNXhhZVkc1KpX5hakJ/UVXV/wI7W8h6nLduKgCUe6k+Vw4eE5GrKmDiumOO8Tp1yviD6NQME4wDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwDQYDVR0OBAYEBAECAwQwDwYDVR0jBAgwBoAEAQIDBDAOBgZRAwQFBgcBAf8EATEwCgYIKoZIzj0EAwMDSAAwRQIhAPRwlo3AyVyGMr+/VWgxPwiOznaExiHY1u211mQAC0a7AiBPScyn4GtDVE+HHiBYSCw5rY5DTAgXpSw0G+sfQ9YVHw=="

		if i+1 < 10 {
			fmt.Println("Creating client: 010101-000" + lastFourSSN)
			_, err := t.create_user(stub, []string{clientEcert, CLIENT, "010101-000" + lastFourSSN , "000" + lastFourSSN, CLIENT_TOKEN})

			if err != nil{
				fmt.Println("Error creating client: 010101-000" + lastFourSSN, err)
			}
		} else if i+1 < 99 {
			fmt.Println("Creating client: 010101-00" + lastFourSSN)

			_, err := t.create_user(stub, []string{clientEcert, CLIENT, "010101-00" + lastFourSSN , "00" + lastFourSSN, CLIENT_TOKEN})

			if err != nil{
				fmt.Println("Error creating client: 010101-00" + lastFourSSN, err)
			}

		} else {
			fmt.Println("Creating client: 010101-0" + lastFourSSN)

			_, err := t.create_user(stub, []string{clientEcert, CLIENT, "010101-0" + lastFourSSN , "0" + lastFourSSN, CLIENT_TOKEN})

			if err != nil{
				fmt.Println("Error creating client: 010101-0" + lastFourSSN, err)
			}
		}

	}

}

// ============================================================================================================================
// CheckRole - Called from all invoke/query func's.
// Access Control happens here
// Params: username, role
// Username should be pass from json input
// Roles: ADMIN_INDEX, DOCTOR_INDEX, CLIENT_INDEX, HOSPITAL_INDEX, LAB_INDEX
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

		if ecertSaved != ecert {
			fmt.Println("\nAccess denied! \n x509Cert not a match!")
			return false
		} else {
			fmt.Println("\nx509Cert signature matches!")
			return true
		}
	}

	fmt.Println("Access denied! Last in function")
	return false
}