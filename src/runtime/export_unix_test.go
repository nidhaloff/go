// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package runtime

func sigismember(mask *sigset, i int) bool {
	clear := *mask
	sigdelset(&clear, i)
	return clear != *mask
}

func Sigisblocked(i int) bool {
	var sigmask sigset
	sigprocmask(_SIG_SETMASK, nil, &sigmask)
	return sigismember(&sigmask, i)
}

type M = m

var waitForSigusr1 struct {
	park note
	mp   *m
}

// WaitForSigusr1 blocks until a SIGUSR1 is received. It calls ready
// when it is set up to receive SIGUSR1. The ready function should
// cause a SIGUSR1 to be sent.
//
// Once SIGUSR1 is received, it returns the ID of the current M and
// the ID of the M the SIGUSR1 was received on. If no SIGUSR1 is
// received for timeoutNS nanoseconds, it returns -1.
func WaitForSigusr1(ready func(mp *M), timeoutNS int64) (int64, int64) {
	mp := getg().m
	testSigusr1 = func(gp *g) bool {
		waitForSigusr1.mp = gp.m
		notewakeup(&waitForSigusr1.park)
		return true
	}
	ready(mp)
	ok := notetsleepg(&waitForSigusr1.park, timeoutNS)
	noteclear(&waitForSigusr1.park)
	gotM := waitForSigusr1.mp
	waitForSigusr1.mp = nil
	testSigusr1 = nil

	if !ok {
		return -1, -1
	}
	return mp.id, gotM.id
}

// SendSigusr1 sends SIGUSR1 to mp.
func SendSigusr1(mp *M) {
	signalM(mp, _SIGUSR1)
}
