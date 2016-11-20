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

package simplebft

import "reflect"

func (s *SBFT) maybeSendCommit() {
	if s.cur.sentCommit || len(s.cur.prep) < s.noFaultyQuorum()-1 {
		return
	}
	s.sendCommit()
}

func (s *SBFT) sendCommit() {
	s.cur.sentCommit = true
	c := s.cur.subject
	s.sys.Persist("commit", &c)
	s.broadcast(&Msg{&Msg_Commit{&c}})
}

func (s *SBFT) handleCommit(c *Subject, src uint64) {
	if c.Seq.Seq < s.cur.subject.Seq.Seq {
		// old message
		return
	}

	if !reflect.DeepEqual(c, &s.cur.subject) {
		log.Warningf("commit does not match expected subject %v %x, got %v %x",
			s.cur.subject.Seq, s.cur.subject.Digest, c.Seq, c.Digest)
		return
	}
	if _, ok := s.cur.commit[src]; ok {
		log.Infof("duplicate commit for %v from %d", *c.Seq, src)
		return
	}
	s.cur.commit[src] = c
	s.maybeExecute()
}
