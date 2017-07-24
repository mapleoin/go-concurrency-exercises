//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"sync/atomic"
	"time"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

const seconds_per_user = 10

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	// if u.IsPremium {
	// 	process()
	// 	return true
	// }

	if atomic.LoadInt64(&u.TimeUsed) >= seconds_per_user {
		return false
	}

	ticker := time.NewTicker(time.Second)

	process_done := make(chan bool)
	go func() {
		process()
		process_done <- true
	}()

	for {
		select {
		case <-process_done:
			return true
		// every second add up the time used to the user object
		// and check if they've exceeded the alloted time
		case <-ticker.C:
			atomic.AddInt64(&u.TimeUsed, 1)
			if atomic.LoadInt64(&u.TimeUsed) >= seconds_per_user {
				return false
			}
		}
	}
}

func main() {
	RunMockServer()
}
