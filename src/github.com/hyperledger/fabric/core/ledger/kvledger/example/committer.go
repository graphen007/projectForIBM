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

package example

import (
	"github.com/hyperledger/fabric/core/ledger"

	pb "github.com/hyperledger/fabric/protos/peer"
)

// Committer a toy committer
type Committer struct {
	ledger ledger.ValidatedLedger
}

// ConstructCommitter constructs a committer for the example
func ConstructCommitter(ledger ledger.ValidatedLedger) *Committer {
	return &Committer{ledger}
}

// CommitBlock commits the block
func (c *Committer) CommitBlock(rawBlock *pb.Block2) (*pb.Block2, []*pb.InvalidTransaction, error) {
	var validBlock *pb.Block2
	var invalidTxs []*pb.InvalidTransaction
	var err error
	logger.Debugf("Committer validating the block...")
	if validBlock, invalidTxs, err = c.ledger.RemoveInvalidTransactionsAndPrepare(rawBlock); err != nil {
		return nil, nil, err
	}
	logger.Debugf("Committer committing the block...")
	if err = c.ledger.Commit(); err != nil {
		return nil, nil, err
	}
	return validBlock, invalidTxs, err
}
