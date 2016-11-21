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
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"




)


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var bloodTestIndex = "_bloodTestIndex"

type bloodTest struct{


	TimeStamp string `json:"timestamp"`
	Name string `json:"name"`
	CPR string `json:"CPR"`
	Doctor string `json:"doctor"`
	Hospital string `json:"hospital"`
	Status string `json:"status"`
	Result string `json:"result"`
	BloodTestID string `json:"BloodTestID"`


}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "init_bloodtest"{
		return t.init_bloodtest(stub,args)
	} else if function == "change_status"{
		return t.change_status(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}


// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//for 0.6 stub shim.ChaincodeStubInterface

	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {                            //read a variable
		return t.read(stub, args)
	}else if function == "read_list"{
		return t.read_list(stub,args)
	}else if function == "patient_read"{
		return t.patient_read(stub,args)
	}else if function == "doctor_read"{
		return t.doctor_read(stub,args)
	}else if function == "hospital_read"{
		return t.hospital_read(stub,args)
	}


	fmt.Println("query did not find func: " + function)
	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var name, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]                            //rename for fun
	value = args[1]
	err = stub.PutState(name, []byte(value))  //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (t *SimpleChaincode) read(stub  *shim.ChaincodeStub, args []string) ([]byte, error) {
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
func (t *SimpleChaincode) patient_read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*
		    Our model looks like
		    -------------------------------------------------------

		    -------------------------------------------------------
		       0
		    "CPR"
		    -------------------------------------------------------
		    */
	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil{
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte
	res := bloodTest{}
	for i:= range bloodInd {


		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.CPR == args[0]{
			finalList = append(finalList, bloodAsBytes...)

		}
	}


	return finalList, nil
}


func (t *SimpleChaincode) doctor_read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*
		    Our model looks like
		    -------------------------------------------------------

		    -------------------------------------------------------
		       0
		    "Doctor"
		    -------------------------------------------------------
		    */
	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil{
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte
	res := bloodTest{}
	for i:= range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.Doctor == args[0]{
			finalList = append(finalList, bloodAsBytes...)

		}
	}


	return finalList, nil
}

func (t *SimpleChaincode) hospital_read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*
		    Our model looks like
		    -------------------------------------------------------
		     String
		    -------------------------------------------------------
		       0
		    "hospital"
		    -------------------------------------------------------
		    */

	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(bloodTestIndex)
	if err != nil {
		return nil, errors.New("Failed to get bloodList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil{
		fmt.Println("you dun goofed")
	}

	var bloodAsBytes []byte
	var finalList []byte
	res := bloodTest{}
	for i:= range bloodInd {


		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.Hospital == args[0]{
			finalList = append(finalList, bloodAsBytes...)

		}
	}


	return finalList, nil
}




func (t *SimpleChaincode) read_list(stub *shim.ChaincodeStub, args []string) ([]byte, error){


	if len(args) != 1 {
		return nil, errors.New("Gimme more arguments, 1 to be exact")
	}
	bloodTestList, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var bloodInd []string

	err = json.Unmarshal(bloodTestList, &bloodInd)
	if err != nil{
		fmt.Println("you dun goofed")
	}
	var finalList []byte
	var bloodAsBytes []byte
	for i:= range bloodInd {

		bloodAsBytes, err = stub.GetState(bloodInd[i])
		finalList = append(finalList, bloodAsBytes...)

	}


	return finalList, nil
}

func (t *SimpleChaincode) change_status(stub *shim.ChaincodeStub, args []string) ([]byte, error){
	/*
		    Our model looks like
		    -------------------------------------------------------

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
	if err != nil{
		fmt.Println("you dun goofed")
	}

	res := bloodTest{}
	var bloodAsBytes []byte
	for i:= range bloodInd {


		bloodAsBytes, err = stub.GetState(bloodInd[i])
		json.Unmarshal(bloodAsBytes, &res)
		if res.BloodTestID == args[0]{
			res.Status = args[1]
			jsonAsBytes, _ := json.Marshal(res)
			err = stub.PutState(args[0], jsonAsBytes)								//rewrite the marble with id as key
			if err != nil {
				return nil, err
			}

		}




	}


	return nil, nil
}

func (t *SimpleChaincode) init_bloodtest(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	/*
	    Our model looks like
	    -------------------------------------------------------
	    -------------------------------------------------------
	       0           1        2       3          4	5	  6	  7
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
	if err != nil{
		return nil, errors.New("blood")
	}
	res := bloodTest{}
	json.Unmarshal(bloodAsBytes, &res)
	if res.BloodTestID == bloodTestID{

		return nil, errors.New("This blood test arleady exists")
	}

	json.Unmarshal(bloodAsBytes, &res)

	str := `{"timeStamp": "` + timeStamp + `", "name": "` + name + `", "CPR": "` + CPR + `", "doctor": "` + doctor +`", "hospital": "` + hospital +`", "status": "` + status +`", "result": "` + result +`", "bloodTestID": "` + bloodTestID +`"}`  		//build the Json element
	err = stub.PutState(bloodTestID, []byte(str))
	if err != nil{
		return nil, err
	}


	//get the blood index
	bloodAsBytes, err = stub.GetState(bloodTestIndex)
	if err != nil{
		return nil, errors.New("you fucked up")
	}

	var bloodInd[]string
	json.Unmarshal(bloodAsBytes, &bloodInd)

	//append it to the list
	bloodInd = append(bloodInd, bloodTestID)
	jsonAsBytes, _ := json.Marshal(bloodInd)
	err = stub.PutState(bloodTestIndex, jsonAsBytes)


	fmt.Println("Ended of creation")

	return nil, nil
}



