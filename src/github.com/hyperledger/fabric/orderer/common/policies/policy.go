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

package policies

import (
	"fmt"

	"github.com/hyperledger/fabric/orderer/common/cauthdsl"
	cb "github.com/hyperledger/fabric/protos/common"

	"github.com/golang/protobuf/proto"
)

// Policy is used to determine if a signature is valid
type Policy interface {
	// Evaluate returns nil if a digest is properly signed by sigs, or an error indicating why it failed
	Evaluate(header [][]byte, payload []byte, identities [][]byte, signatures [][]byte) error
}

// Manager is intended to be the primary accessor of ManagerImpl
// It is intended to discourage use of the other exported ManagerImpl methods
// which are used for updating policy by the ConfigManager
type Manager interface {
	// GetPolicy returns a policy and true if it was the policy requested, or false if it is the default policy
	GetPolicy(id string) (Policy, bool)
}

type policy struct {
	source    *cb.Policy
	evaluator *cauthdsl.SignaturePolicyEvaluator
}

func newPolicy(policySource *cb.Policy, ch cauthdsl.CryptoHelper) (*policy, error) {
	envelopeWrapper, ok := policySource.Type.(*cb.Policy_SignaturePolicy)

	if !ok {
		return nil, fmt.Errorf("Unknown policy type: %T", policySource.Type)
	}

	if envelopeWrapper.SignaturePolicy == nil {
		return nil, fmt.Errorf("Nil signature policy received")
	}

	sigPolicy := envelopeWrapper.SignaturePolicy

	evaluator, err := cauthdsl.NewSignaturePolicyEvaluator(sigPolicy, ch)
	if err != nil {
		return nil, err
	}

	return &policy{
		evaluator: evaluator,
		source:    policySource,
	}, nil
}

// Evaluate returns nil if a msg is properly signed by sigs, or an error indicating why it failed
// For each identity, it concatenates the corresponding header and the payload together, then
// verifies the corresponding signature.
func (p *policy) Evaluate(header [][]byte, payload []byte, identities [][]byte, signatures [][]byte) error {
	if p == nil {
		return fmt.Errorf("Evaluated default policy, results in reject")
	}

	// XXX This is wrong, as the signatures are over the payload envelope, not the message, fix either here, or in cauthdsl once transaction is finalized
	if !p.evaluator.Authenticate(payload, identities, signatures) {
		return fmt.Errorf("Failed to authenticate policy")
	}
	return nil
}

// ManagerImpl is an implementation of Manager and configtx.ConfigHandler
// In general, it should only be referenced as an Impl for the configtx.ConfigManager
type ManagerImpl struct {
	policies        map[string]*policy
	pendingPolicies map[string]*policy
	ch              cauthdsl.CryptoHelper
}

// NewManagerImpl creates a new ManagerImpl with the given CryptoHelper
func NewManagerImpl(ch cauthdsl.CryptoHelper) *ManagerImpl {
	return &ManagerImpl{
		ch:       ch,
		policies: make(map[string]*policy),
	}
}

// GetPolicy returns a policy and true if it was the policy requested, or false if it is the default policy
func (pm *ManagerImpl) GetPolicy(id string) (Policy, bool) {
	policy, ok := pm.policies[id]
	// Note the nil policy evaluates fine
	return policy, ok
}

// BeginConfig is used to start a new configuration proposal
func (pm *ManagerImpl) BeginConfig() {
	if pm.pendingPolicies != nil {
		panic("Programming error, cannot call begin in the middle of a proposal")
	}
	pm.pendingPolicies = make(map[string]*policy)
}

// RollbackConfig is used to abandon a new configuration proposal
func (pm *ManagerImpl) RollbackConfig() {
	pm.pendingPolicies = nil
}

// CommitConfig is used to commit a new configuration proposal
func (pm *ManagerImpl) CommitConfig() {
	if pm.pendingPolicies == nil {
		panic("Programming error, cannot call commit without an existing proposal")
	}
	pm.policies = pm.pendingPolicies
	pm.pendingPolicies = nil
}

// ProposeConfig is used to add new configuration to the configuration proposal
func (pm *ManagerImpl) ProposeConfig(configItem *cb.ConfigurationItem) error {
	if configItem.Type != cb.ConfigurationItem_Policy {
		return fmt.Errorf("Expected type of ConfigurationItem_Policy, got %v", configItem.Type)
	}

	policy := &cb.Policy{}
	err := proto.Unmarshal(configItem.Value, policy)
	if err != nil {
		return err
	}

	cPolicy, err := newPolicy(policy, pm.ch)
	if err != nil {
		return err
	}

	pm.pendingPolicies[configItem.Key] = cPolicy
	return nil
}
