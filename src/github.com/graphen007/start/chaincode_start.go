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

var integerIndexname = "_integerindex"

type allIntegers struct {
	AllInts []integerDefine `json:"all_integers"`
}

type integerDefine struct{
	User string `json:"user"`
	TheNumber string `json:"number"`
	Name string `json:"name"`

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

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
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
	} else if function == "init_integer"{
		return t.init_integer(stub,args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}


// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {                            //read a variable
		return t.read(stub, args)
	}else if function =="read_list"{
		return t.read_list(stub,args)
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
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
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

func (t *SimpleChaincode) read_list(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	intList, err := stub.GetState(integerIndexname)
	if err != nil {
		return nil, errors.New("Failed to get intList")
	}
	var intToReturn integerDefine
	err = json.Unmarshal(intList, &intToReturn)
	if err != nil{
		fmt.Println("you dun goofed")
	}


	return intToReturn, nil
}
func (t *SimpleChaincode) init_integer(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var user string
	var number string
	var err error
	//var err error
	fmt.Println("Creating the Int")
	if len(args) != 3 {
		return nil, errors.New("Gimme more arguments, 2 to be exact, User and number pliz")
	}

	user = args[0]
	number = args[1]
	name:= args[2]

	intAsBytes, err := stub.GetState(number)
	if err != nil{
		return nil, errors.New("Failed to get integer")
	}

	res := integerDefine{} 						// Get the above defined marble struct
	json.Unmarshal(intAsBytes, &res)
	if res.TheNumber == number {
		fmt.Println("The number already exists: " + number)
		fmt.Println(res)
		return nil, errors.New("This number already exists")
	}



	str := `{"user": "` + user + `", "number": "` + number + `"name": "` + name + `"}`  		//build the Json element
	err = stub.PutState(name, []byte(str))								// store int with key
	if err != nil{
		return nil, err
	}


	//get the int index
	intAsBytes, err = stub.GetState(integerIndexname)
	if err != nil{
		return nil, errors.New("you fucked up")
	}

	var integerIndex[]string
	json.Unmarshal(intAsBytes, &integerIndex)

	//append it to the list
	integerIndex = append(integerIndex, name)
	jsonAsBytes, _ := json.Marshal(integerIndex)
	err = stub.PutState(integerIndexname, jsonAsBytes)


	fmt.Println("Ended of creation")

	return nil, nil
}