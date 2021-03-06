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

package endorser

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode"
	"github.com/hyperledger/fabric/core/container"
	"github.com/hyperledger/fabric/core/crypto"
	"github.com/hyperledger/fabric/core/crypto/primitives"
	"github.com/hyperledger/fabric/core/db"
	"github.com/hyperledger/fabric/core/ledger/kvledger"
	"github.com/hyperledger/fabric/core/peer"
	"github.com/hyperledger/fabric/core/util"
	"github.com/hyperledger/fabric/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
	pbutils "github.com/hyperledger/fabric/protos/utils"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var testDBWrapper = db.NewTestDBWrapper()
var endorserServer pb.EndorserServer
var mspInstance msp.PeerMSP
var signer msp.SigningIdentity

//initialize peer and start up. If security==enabled, login as vp
func initPeer() (net.Listener, error) {
	//start clean
	finitPeer(nil)
	var opts []grpc.ServerOption
	if viper.GetBool("peer.tls.enabled") {
		creds, err := credentials.NewServerTLSFromFile(viper.GetString("peer.tls.cert.file"), viper.GetString("peer.tls.key.file"))
		if err != nil {
			return nil, fmt.Errorf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)

	viper.Set("peer.fileSystemPath", filepath.Join(os.TempDir(), "hyperledger", "production"))

	peerAddress, err := peer.GetLocalAddress()
	if err != nil {
		return nil, fmt.Errorf("Error obtaining peer address: %s", err)
	}
	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		return nil, fmt.Errorf("Error starting peer listener %s", err)
	}

	//initialize ledger
	lpath := viper.GetString("peer.fileSystemPath")
	kvledger.Initialize(lpath)

	getPeerEndpoint := func() (*pb.PeerEndpoint, error) {
		return &pb.PeerEndpoint{ID: &pb.PeerID{Name: "testpeer"}, Address: peerAddress}, nil
	}

	// Install security object for peer
	var secHelper crypto.Peer
	if viper.GetBool("security.enabled") {
		//TODO:  integrate new crypto / idp
		securityLevel := viper.GetInt("security.level")
		hashAlgorithm := viper.GetString("security.hashAlgorithm")
		primitives.SetSecurityLevel(hashAlgorithm, securityLevel)
	} else {
		// the primitives need to be instantiated no matter what. Otherwise
		// the escc code won't have a hash algorithm available to hash the proposal
		primitives.SetSecurityLevel("SHA2", 256)
	}

	ccStartupTimeout := time.Duration(30000) * time.Millisecond
	pb.RegisterChaincodeSupportServer(grpcServer, chaincode.NewChaincodeSupport(chaincode.DefaultChain, getPeerEndpoint, false, ccStartupTimeout, secHelper))

	chaincode.RegisterSysCCs()

	chaincodeID := &pb.ChaincodeID{Path: "github.com/hyperledger/fabric/core/chaincode/lccc", Name: "lccc"}
	spec := pb.ChaincodeSpec{Type: pb.ChaincodeSpec_Type(pb.ChaincodeSpec_Type_value["GOLANG"]), ChaincodeID: chaincodeID, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("")}}}

	chaincode.DeploySysCC(context.Background(), &spec)

	go grpcServer.Serve(lis)

	return lis, nil
}

func finitPeer(lis net.Listener) {
	closeListenerAndSleep(lis)
	os.RemoveAll(filepath.Join(os.TempDir(), "hyperledger"))
}

func closeListenerAndSleep(l net.Listener) {
	if l != nil {
		l.Close()
		time.Sleep(2 * time.Second)
	}
}

//getProposal gets the proposal for the chaincode invocation
//Currently supported only for Invokes (Queries still go through devops client)
func getProposal(cis *pb.ChaincodeInvocationSpec, creator []byte) (*pb.Proposal, error) {
	return pbutils.CreateChaincodeProposal(cis, creator)
}

//getDeployProposal gets the proposal for the chaincode deployment
//the payload is a ChaincodeDeploymentSpec
func getDeployProposal(cds *pb.ChaincodeDeploymentSpec, creator []byte) (*pb.Proposal, error) {
	b, err := proto.Marshal(cds)
	if err != nil {
		return nil, err
	}

	//wrap the deployment in an invocation spec to lccc...
	lcccSpec := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{Type: pb.ChaincodeSpec_GOLANG, ChaincodeID: &pb.ChaincodeID{Name: "lccc"}, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("deploy"), []byte("default"), b}}}}

	//...and get the proposal for it
	return getProposal(lcccSpec, creator)
}

func getSignedProposal(prop *pb.Proposal, signer msp.SigningIdentity) (*pb.SignedProposal, error) {
	propBytes, err := pbutils.GetBytesProposal(prop)
	if err != nil {
		return nil, err
	}

	signature, err := signer.Sign(propBytes)
	if err != nil {
		return nil, err
	}

	return &pb.SignedProposal{ProposalBytes: propBytes, Signature: signature}, nil
}

func getDeploymentSpec(context context.Context, spec *pb.ChaincodeSpec) (*pb.ChaincodeDeploymentSpec, error) {
	codePackageBytes, err := container.GetChaincodePackageBytes(spec)
	if err != nil {
		return nil, err
	}
	chaincodeDeploymentSpec := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec, CodePackage: codePackageBytes}
	return chaincodeDeploymentSpec, nil
}

func deploy(endorserServer pb.EndorserServer, spec *pb.ChaincodeSpec, f func(*pb.ChaincodeDeploymentSpec)) (*pb.ProposalResponse, error) {
	var err error
	var depSpec *pb.ChaincodeDeploymentSpec

	ctxt := context.Background()
	depSpec, err = getDeploymentSpec(ctxt, spec)
	if err != nil {
		return nil, err
	}

	if f != nil {
		f(depSpec)
	}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, err
	}

	var prop *pb.Proposal
	prop, err = getDeployProposal(depSpec, creator)
	if err != nil {
		return nil, err
	}

	var signedProp *pb.SignedProposal
	signedProp, err = getSignedProposal(prop, signer)
	if err != nil {
		return nil, err
	}

	var resp *pb.ProposalResponse
	resp, err = endorserServer.ProcessProposal(context.Background(), signedProp)

	return resp, err
}

func invoke(spec *pb.ChaincodeSpec) (*pb.ProposalResponse, error) {
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, err
	}

	var prop *pb.Proposal
	prop, err = getProposal(invocation, creator)
	if err != nil {
		return nil, fmt.Errorf("Error creating proposal  %s: %s\n", spec.ChaincodeID, err)
	}

	var signedProp *pb.SignedProposal
	signedProp, err = getSignedProposal(prop, signer)
	if err != nil {
		return nil, err
	}

	resp, err := endorserServer.ProcessProposal(context.Background(), signedProp)
	if err != nil {
		return nil, fmt.Errorf("Error endorsing %s: %s\n", spec.ChaincodeID, err)
	}

	return resp, err
}

//begin tests. Note that we rely upon the system chaincode and peer to be created
//once and be used for all the tests. In order to avoid dependencies / collisions
//due to deployed chaincodes, trying to use different chaincodes for different
//tests

//TestDeploy deploy chaincode example01
func TestDeploy(t *testing.T) {
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeID: &pb.ChaincodeID{Name: "ex01", Path: "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example01"}, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte("200")}}}

	_, err := deploy(endorserServer, spec, nil)
	if err != nil {
		t.Fail()
		t.Logf("Deploy-error in deploy %s", err)
		chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
		return
	}
	chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
}

//TestDeployBadArgs sets bad args on deploy. It should fail, and example02 should not be deployed
func TestDeployBadArgs(t *testing.T) {
	//invalid arguments
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeID: &pb.ChaincodeID{Name: "ex02", Path: "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example02"}, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b")}}}

	_, err := deploy(endorserServer, spec, nil)
	if err == nil {
		t.Fail()
		t.Log("DeployBadArgs-expected error in deploy but succeeded")
		chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
		return
	}
	chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
}

//TestDeployBadPayload set payload to nil and do a deploy. It should fail and example02 should not be deployed
func TestDeployBadPayload(t *testing.T) {
	//invalid arguments
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeID: &pb.ChaincodeID{Name: "ex02", Path: "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example02"}, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte("200")}}}

	f := func(cds *pb.ChaincodeDeploymentSpec) {
		cds.CodePackage = nil
	}
	_, err := deploy(endorserServer, spec, f)
	if err == nil {
		t.Fail()
		t.Log("DeployBadPayload-expected error in deploy but succeeded")
		chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
		return
	}
	chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
}

//TestRedeploy - deploy two times, second time should fail but example02 should remain deployed
func TestRedeploy(t *testing.T) {

	//invalid arguments
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeID: &pb.ChaincodeID{Name: "ex02", Path: "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example02"}, CtorMsg: &pb.ChaincodeInput{Args: [][]byte{[]byte("init"), []byte("a"), []byte("100"), []byte("b"), []byte("200")}}}

	_, err := deploy(endorserServer, spec, nil)
	if err != nil {
		t.Fail()
		t.Logf("error in endorserServer.ProcessProposal %s", err)
		chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
		return
	}

	//second time should not fail as we are just simulating
	_, err = deploy(endorserServer, spec, nil)
	if err != nil {
		t.Fail()
		t.Logf("error in endorserServer.ProcessProposal %s", err)
		chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
		return
	}
	chaincode.GetChain(chaincode.DefaultChain).Stop(context.Background(), &pb.ChaincodeDeploymentSpec{ChaincodeSpec: spec})
}

// TestDeployAndInvoke deploys and invokes chaincode_example01
func TestDeployAndInvoke(t *testing.T) {
	var ctxt = context.Background()

	url := "github.com/hyperledger/fabric/examples/chaincode/go/chaincode_example01"
	chaincodeID := &pb.ChaincodeID{Path: url, Name: "ex01"}

	args := []string{"10"}

	f := "init"
	argsDeploy := util.ToChaincodeArgs(f, "a", "100", "b", "200")
	spec := &pb.ChaincodeSpec{Type: 1, ChaincodeID: chaincodeID, CtorMsg: &pb.ChaincodeInput{Args: argsDeploy}}
	resp, err := deploy(endorserServer, spec, nil)
	chaincodeID1 := spec.ChaincodeID.Name
	if err != nil {
		t.Fail()
		t.Logf("Error deploying <%s>: %s", chaincodeID1, err)
		return
	}

	err = endorserServer.(*Endorser).commitTxSimulation(resp)
	if err != nil {
		t.Fail()
		t.Logf("Error committing <%s>: %s", chaincodeID1, err)
		return
	}

	f = "invoke"
	invokeArgs := append([]string{f}, args...)
	spec = &pb.ChaincodeSpec{Type: 1, ChaincodeID: chaincodeID, CtorMsg: &pb.ChaincodeInput{Args: util.ToChaincodeArgs(invokeArgs...)}}
	resp, err = invoke(spec)
	if err != nil {
		t.Fail()
		t.Logf("Error invoking transaction: %s", err)
		return
	} else {
		fmt.Printf("Invoke test passed\n")
		t.Logf("Invoke test passed")
	}

	chaincode.GetChain(chaincode.DefaultChain).Stop(ctxt, &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{ChaincodeID: chaincodeID}})
}

func TestMain(m *testing.M) {
	SetupTestConfig()
	testDBWrapper.CleanDB(nil)
	viper.Set("peer.fileSystemPath", filepath.Join(os.TempDir(), "hyperledger", "production"))
	lis, err := initPeer()
	if err != nil {
		os.Exit(-1)
		fmt.Printf("Could not initialize tests")
		finitPeer(lis)
		return
	}

	endorserServer = NewEndorserServer(nil)

	// setup the MSP manager so that we can sign/verify
	mspMgrConfigFile := "../../msp/peer-config.json"
	msp.GetManager().Setup(mspMgrConfigFile)
	mspId := "DEFAULT"
	id := "PEER"
	signingIdentity := &msp.IdentityIdentifier{Mspid: msp.ProviderIdentifier{Value: mspId}, Value: id}
	signer, err = msp.GetManager().GetSigningIdentity(signingIdentity)
	if err != nil {
		os.Exit(-1)
		fmt.Printf("Could not initialize msp/signer")
		finitPeer(lis)
		return
	}

	retVal := m.Run()

	finitPeer(lis)

	os.Exit(retVal)
}
