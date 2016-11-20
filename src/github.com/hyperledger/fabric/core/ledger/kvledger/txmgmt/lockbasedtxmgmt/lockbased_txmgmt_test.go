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

package lockbasedtxmgmt

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/ledger/testutil"
)

func TestTxSimulatorWithNoExistingData(t *testing.T) {
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()
	s, _ := txMgr.NewTxSimulator()
	value, err := s.GetState("ns1", "key1")
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in GetState(): %s", err))
	testutil.AssertNil(t, value)

	s.SetState("ns1", "key1", []byte("value1"))
	s.SetState("ns1", "key2", []byte("value2"))
	s.SetState("ns2", "key3", []byte("value3"))
	s.SetState("ns2", "key4", []byte("value4"))

	value, _ = s.GetState("ns2", "key3")
	testutil.AssertEquals(t, value, []byte("value3"))

	s.DeleteState("ns2", "key3")
	value, _ = s.GetState("ns2", "key3")
	testutil.AssertNil(t, value)
}

func TestTxSimulatorWithExistingData(t *testing.T) {
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)

	// simulate tx1
	s1, _ := txMgr.NewTxSimulator()
	defer txMgr.Shutdown()
	s1.SetState("ns1", "key1", []byte("value1"))
	s1.SetState("ns1", "key2", []byte("value2"))
	s1.SetState("ns2", "key3", []byte("value3"))
	s1.SetState("ns2", "key4", []byte("value4"))
	s1.Done()
	// validate and commit RWset
	txRWSet := s1.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, fmt.Sprintf("Error while calling commit(): %s", err))

	// simulate tx2 that make changes to existing data
	s2, _ := txMgr.NewTxSimulator()
	value, _ := s2.GetState("ns1", "key1")
	testutil.AssertEquals(t, value, []byte("value1"))
	s2.SetState("ns1", "key1", []byte("value1_1"))
	s2.DeleteState("ns2", "key3")
	value, _ = s2.GetState("ns1", "key1")
	testutil.AssertEquals(t, value, []byte("value1_1"))
	s2.Done()
	// validate and commit RWset for tx2
	txRWSet = s2.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err = txMgr.validateTx(txRWSet)
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	txMgr.Commit()

	// simulate tx3
	s3, _ := txMgr.NewTxSimulator()
	value, _ = s3.GetState("ns1", "key1")
	testutil.AssertEquals(t, value, []byte("value1_1"))
	value, _ = s3.GetState("ns2", "key3")
	testutil.AssertEquals(t, value, nil)
	s3.Done()

	// verify the versions of keys in persistence
	ver, _ := txMgr.getCommitedVersion("ns1", "key1")
	testutil.AssertEquals(t, ver, uint64(2))
	ver, _ = txMgr.getCommitedVersion("ns1", "key2")
	testutil.AssertEquals(t, ver, uint64(1))
	ver, _ = txMgr.getCommitedVersion("ns2", "key3")
	testutil.AssertEquals(t, ver, uint64(2))
}

func TestTxValidation(t *testing.T) {
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()

	// simulate tx1
	s1, _ := txMgr.NewTxSimulator()
	s1.SetState("ns1", "key1", []byte("value1"))
	s1.SetState("ns1", "key2", []byte("value2"))
	s1.SetState("ns2", "key3", []byte("value3"))
	s1.SetState("ns2", "key4", []byte("value4"))
	s1.Done()
	// validate and commit RWset
	txRWSet := s1.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, fmt.Sprintf("Error while calling commit(): %s", err))

	// simulate tx2 that make changes to existing data
	s2, _ := txMgr.NewTxSimulator()
	value, _ := s2.GetState("ns1", "key1")
	testutil.AssertEquals(t, value, []byte("value1"))

	s2.SetState("ns1", "key1", []byte("value1_2"))
	s2.DeleteState("ns2", "key3")
	s2.Done()

	// simulate tx3 before committing tx2 changes. Reads and modifies the key changed by tx2
	s3, _ := txMgr.NewTxSimulator()
	s3.GetState("ns1", "key1")
	s3.SetState("ns1", "key1", []byte("value1_3"))
	s3.Done()

	// simulate tx4 before committing tx2 changes. Reads and Deletes the key changed by tx2
	s4, _ := txMgr.NewTxSimulator()
	s4.GetState("ns2", "key3")
	s4.DeleteState("ns2", "key3")
	s4.Done()

	// simulate tx5 before committing tx2 changes. Modifies and then Reads the key changed by tx2 and writes a new key
	s5, _ := txMgr.NewTxSimulator()
	s5.SetState("ns1", "key1", []byte("new_value"))
	s5.GetState("ns1", "key1")
	s5.Done()

	// simulate tx6 before committing tx2 changes. Only writes a new key, does not reads/writes a key changed by tx2
	s6, _ := txMgr.NewTxSimulator()
	s6.SetState("ns1", "new_key", []byte("new_value"))
	s6.Done()

	// validate and commit RWset for tx2
	txRWSet = s2.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err = txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	txMgr.Commit()

	//RWSet for tx3 and tx4 should not be invalid now
	isValid, err = txMgr.validateTx(s3.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, false)

	isValid, err = txMgr.validateTx(s4.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, false)

	//tx5 shold still be valid as it over-writes the key first and then reads
	isValid, _ = txMgr.validateTx(s5.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertSame(t, isValid, true)

	// tx6 should still be valid as it only writes a new key
	isValid, _ = txMgr.validateTx(s6.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertSame(t, isValid, true)
}

func TestEncodeDecodeValueAndVersion(t *testing.T) {
	testValueAndVersionEncodeing(t, []byte("value1"), uint64(1))
	testValueAndVersionEncodeing(t, nil, uint64(2))
}

func testValueAndVersionEncodeing(t *testing.T, value []byte, version uint64) {
	encodedValue := encodeValue(value, version)
	val, ver := decodeValue(encodedValue)
	testutil.AssertEquals(t, val, value)
	testutil.AssertEquals(t, ver, version)
}

func TestIterator(t *testing.T) {
	testIterator(t, 10, 2, 7)
	testIterator(t, 10, 1, 11)
	testIterator(t, 10, 0, 0)
	testIterator(t, 10, 5, 0)
	testIterator(t, 10, 0, 5)
}

func testIterator(t *testing.T, numKeys int, startKeyNum int, endKeyNum int) {
	cID := "cID"
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()
	s, _ := txMgr.NewTxSimulator()
	for i := 1; i <= numKeys; i++ {
		k := createTestKey(i)
		v := createTestValue(i)
		t.Logf("Adding k=[%s], v=[%s]", k, v)
		s.SetState(cID, k, v)
	}
	s.Done()
	// validate and commit RWset
	txRWSet := s.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, "")
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, "")

	var startKey string
	var endKey string
	var begin int
	var end int

	if startKeyNum != 0 {
		begin = startKeyNum
		startKey = createTestKey(startKeyNum)
	} else {
		begin = 1 //first key in the db
		startKey = ""
	}

	if endKeyNum != 0 {
		endKey = createTestKey(endKeyNum)
		end = endKeyNum
	} else {
		endKey = ""
		end = numKeys + 1 //last key in the db
	}

	expectedCount := end - begin

	queryExecuter, _ := txMgr.NewQueryExecutor()
	itr, _ := queryExecuter.GetStateRangeScanIterator(cID, startKey, endKey)
	count := 0
	for {
		kv, _ := itr.Next()
		if kv == nil {
			break
		}
		keyNum := begin + count
		k := kv.(*ledger.KV).Key
		v := kv.(*ledger.KV).Value
		t.Logf("Retrieved k=%s, v=%s", k, v)
		testutil.AssertEquals(t, k, createTestKey(keyNum))
		testutil.AssertEquals(t, v, createTestValue(keyNum))
		count++
	}
	testutil.AssertEquals(t, count, expectedCount)
}

func TestIteratorWithDeletes(t *testing.T) {
	cID := "cID"
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()
	s, _ := txMgr.NewTxSimulator()
	for i := 1; i <= 10; i++ {
		k := createTestKey(i)
		v := createTestValue(i)
		t.Logf("Adding k=[%s], v=[%s]", k, v)
		s.SetState(cID, k, v)
	}
	s.Done()
	// validate and commit RWset
	txRWSet := s.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, "")
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, "")

	s, _ = txMgr.NewTxSimulator()
	s.DeleteState(cID, createTestKey(4))
	s.Done()
	// validate and commit RWset
	txRWSet = s.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err = txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, "")
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, "")

	queryExecuter, _ := txMgr.NewQueryExecutor()
	itr, _ := queryExecuter.GetStateRangeScanIterator(cID, createTestKey(3), createTestKey(6))
	defer itr.Close()
	kv, _ := itr.Next()
	testutil.AssertEquals(t, kv.(*ledger.KV).Key, createTestKey(3))
	kv, _ = itr.Next()
	testutil.AssertEquals(t, kv.(*ledger.KV).Key, createTestKey(5))
}

func TestTxValidationWithItr(t *testing.T) {
	cID := "cID"
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()

	// simulate tx1
	s1, _ := txMgr.NewTxSimulator()
	for i := 1; i <= 10; i++ {
		k := createTestKey(i)
		v := createTestValue(i)
		t.Logf("Adding k=[%s], v=[%s]", k, v)
		s1.SetState(cID, k, v)
	}
	s1.Done()
	// validate and commit RWset
	txRWSet := s1.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, "")
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, "")

	// simulate tx2 that reads key_001 and key_002
	s2, _ := txMgr.NewTxSimulator()
	itr, _ := s2.GetStateRangeScanIterator(cID, createTestKey(1), createTestKey(5))
	// read key_001 and key_002
	itr.Next()
	itr.Next()
	itr.Close()
	s2.Done()

	// simulate tx3 that reads key_004 and key_005
	s3, _ := txMgr.NewTxSimulator()
	itr, _ = s3.GetStateRangeScanIterator(cID, createTestKey(4), createTestKey(6))
	// read key_001 and key_002
	itr.Next()
	itr.Next()
	itr.Close()
	s3.Done()

	// simulate tx4 before committing tx2 and tx3. Modifies a key read by tx3
	s4, _ := txMgr.NewTxSimulator()
	s4.DeleteState(cID, createTestKey(5))
	s4.Done()

	// validate and commit RWset for tx4
	txRWSet = s4.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err = txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	txMgr.Commit()

	//RWSet tx3 should not be invalid now
	isValid, err = txMgr.validateTx(s3.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertNoError(t, err, fmt.Sprintf("Error in validateTx(): %s", err))
	testutil.AssertSame(t, isValid, false)

	// tx2 should still be valid
	isValid, _ = txMgr.validateTx(s2.(*LockBasedTxSimulator).getTxReadWriteSet())
	testutil.AssertSame(t, isValid, true)
}

func TestGetSetMultipeKeys(t *testing.T) {
	cID := "cID"
	env := newTestEnv(t)
	defer env.Cleanup()
	txMgr := NewLockBasedTxMgr(env.conf)
	defer txMgr.Shutdown()

	// simulate tx1
	s1, _ := txMgr.NewTxSimulator()
	multipleKeyMap := make(map[string][]byte)
	for i := 1; i <= 10; i++ {
		k := createTestKey(i)
		v := createTestValue(i)
		multipleKeyMap[k] = v
	}
	s1.SetStateMultipleKeys(cID, multipleKeyMap)
	s1.Done()
	// validate and commit RWset
	txRWSet := s1.(*LockBasedTxSimulator).getTxReadWriteSet()
	isValid, err := txMgr.validateTx(txRWSet)
	testutil.AssertNoError(t, err, "")
	testutil.AssertSame(t, isValid, true)
	txMgr.addWriteSetToBatch(txRWSet)
	err = txMgr.Commit()
	testutil.AssertNoError(t, err, "")

	qe, _ := txMgr.NewQueryExecutor()
	defer qe.Done()
	multipleKeys := []string{}
	for k := range multipleKeyMap {
		multipleKeys = append(multipleKeys, k)
	}
	values, _ := qe.GetStateMultipleKeys(cID, multipleKeys)
	testutil.AssertEquals(t, len(values), 10)
	for i, v := range values {
		testutil.AssertEquals(t, v, multipleKeyMap[multipleKeys[i]])
	}

	s2, _ := txMgr.NewTxSimulator()
	defer s2.Done()
	values, _ = s2.GetStateMultipleKeys(cID, multipleKeys[5:7])
	testutil.AssertEquals(t, len(values), 2)
	for i, v := range values {
		testutil.AssertEquals(t, v, multipleKeyMap[multipleKeys[i+5]])
	}
}

func createTestKey(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf("key_%03d", i)
}

func createTestValue(i int) []byte {
	return []byte(fmt.Sprintf("value_%03d", i))
}
