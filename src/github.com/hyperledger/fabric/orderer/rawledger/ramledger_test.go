/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package rawledger_test

import (
	. "github.com/hyperledger/fabric/orderer/rawledger"
	"github.com/hyperledger/fabric/orderer/rawledger/ramledger"
)

func init() {
	testables = append(testables, &ramLedgerTestEnv{})
}

type ramLedgerFactory struct{}

type ramLedgerTestEnv struct{}

func (env *ramLedgerTestEnv) Initialize() (ledgerFactory, error) {
	return &ramLedgerFactory{}, nil
}

func (env *ramLedgerTestEnv) Name() string {
	return "ramledger"
}

func (env *ramLedgerFactory) Destroy() error {
	return nil
}

func (env *ramLedgerFactory) Persistent() bool {
	return false
}

func (env *ramLedgerFactory) New() ReadWriter {
	historySize := 10
	return ramledger.New(historySize, genesisBlock)
}
