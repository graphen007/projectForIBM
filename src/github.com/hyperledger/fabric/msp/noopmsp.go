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

package msp

type noopmsp struct {
}

func newNoopMsp() PeerMSP {
	mspLogger.Infof("Creating no-op MSP instance")
	return &noopmsp{}
}

func (msp *noopmsp) Setup(configFile string) error {
	return nil
}

func (msp *noopmsp) Reconfig(reconfigMessage string) error {
	return nil
}

func (msp *noopmsp) Type() ProviderType {
	return 0
}

func (msp *noopmsp) Identifier() (*ProviderIdentifier, error) {
	return &ProviderIdentifier{}, nil
}

func (msp *noopmsp) Policy() string {
	return ""
}

func (msp *noopmsp) ImportSigningIdentity(req *ImportRequest) (SigningIdentity, error) {
	return nil, nil
}

func (msp *noopmsp) GetSigningIdentity(identifier *IdentityIdentifier) (SigningIdentity, error) {
	mspLogger.Infof("Obtaining signing identity for %s", identifier)
	id, _ := newNoopSigningIdentity()
	return id, nil
}

func (msp *noopmsp) DeserializeIdentity(serializedID []byte) (Identity, error) {
	mspLogger.Infof("Obtaining identity for %s", string(serializedID))
	id, _ := newNoopIdentity()
	return id, nil
}

func (msp *noopmsp) DeleteSigningIdentity(identifier string) (bool, error) {
	return true, nil
}

func (msp *noopmsp) IsValid(id Identity) (bool, error) {
	return true, nil
}

type noopidentity struct {
}

func newNoopIdentity() (Identity, error) {
	mspLogger.Infof("Creating no-op identity instance")
	return &noopidentity{}, nil
}

func (id *noopidentity) Identifier() *IdentityIdentifier {
	return &IdentityIdentifier{Mspid: ProviderIdentifier{Value: "NOOP"}, Value: "Bob"}
}

func (id *noopidentity) GetMSPIdentifier() string {
	return "MSPID"
}

func (id *noopidentity) Validate() (bool, error) {
	mspLogger.Infof("Identity is valid")
	return true, nil
}

func (id *noopidentity) ParticipantID() string {
	return "dunno"
}

func (id *noopidentity) Verify(msg []byte, sig []byte) (bool, error) {
	mspLogger.Infof("Signature is valid")
	return true, nil
}

func (id *noopidentity) VerifyOpts(msg []byte, sig []byte, opts SignatureOpts) (bool, error) {
	return true, nil
}

func (id *noopidentity) VerifyAttributes(proof [][]byte, spec *AttributeProofSpec) (bool, error) {
	return true, nil
}

func (id *noopidentity) Serialize() ([]byte, error) {
	mspLogger.Infof("Serialinzing identity")
	return []byte("cert"), nil
}

type noopsigningidentity struct {
	noopidentity
}

func newNoopSigningIdentity() (SigningIdentity, error) {
	mspLogger.Infof("Creating no-op signing identity instance")
	return &noopsigningidentity{}, nil
}

func (id *noopsigningidentity) Identity() {

}

func (id *noopsigningidentity) Sign(msg []byte) ([]byte, error) {
	mspLogger.Infof("signing message")
	return []byte("signature"), nil
}

func (id *noopsigningidentity) SignOpts(msg []byte, opts SignatureOpts) ([]byte, error) {
	return nil, nil
}

func (id *noopsigningidentity) GetAttributeProof(spec *AttributeProofSpec) (proof []byte, err error) {
	return nil, nil
}

func (id *noopsigningidentity) GetPublicVersion() Identity {
	return id
}

func (id *noopsigningidentity) Renew() error {
	return nil
}
