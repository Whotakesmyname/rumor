package service

import (
	"container/list"
	"sync"
	"time"
)

// Table is an interface for a table.
// Delete/Remove method has not been considered necessary yet because the table implements a gc mechanism.
type Table interface {
	Get(*Cookie) chan<- *Datagram
	Add(*Cookie, chan<- *Datagram) *Cookie
}

// CookieTable will collect outdated cookie automatically.
type CookieTable struct {
	Map   map[Cookie]chan<- *Datagram
	Queue *list.List // The oldest one at the front and the latest one at the end.
	Lock  *sync.Mutex
}

// QueueMember is the member in a CookieTable's Queue.
type QueueMember struct {
	ptrCookie *Cookie
	timestamp time.Time // UTC
}

// NewCookieTable creates a new cookie table for outgoing requests.
func NewCookieTable() *CookieTable {
	ptrCookieTable := &CookieTable{make(map[Cookie]chan<- *Datagram, 25), list.New(), &sync.Mutex{}}
	go func() {
		var qMember QueueMember
		var prev *list.Element
		var channel chan<- *Datagram
		var isExist bool
		for tNow := range time.Tick(time.Duration(RefreshInternal) * time.Second) {
			tNow = tNow.UTC()

			ptrCookieTable.Lock.Lock()
			prev = nil
			for p := ptrCookieTable.Queue.Front(); p != nil; p = p.Next() {
				// Delete the previous element from the list if exist
				if prev != nil {
					ptrCookieTable.Queue.Remove(prev)
				}

				qMember = p.Value.(QueueMember)
				if tNow.Sub(qMember.timestamp).Seconds() < RequestTimeout {
					break
				}
				channel, isExist = ptrCookieTable.Map[*qMember.ptrCookie]
				if isExist {
					close(channel)
				}
				delete(ptrCookieTable.Map, *qMember.ptrCookie)
				prev = p
			}
			ptrCookieTable.Lock.Unlock()
		}
	}()
	return ptrCookieTable
}

// Add cookie to table, timestamp will be created automatically.
// Return value false means a conflict happens between the cookie and an existing cookie.
func (table *CookieTable) Add(ptrCookie *Cookie, channel chan<- *Datagram) *Cookie {
	table.Lock.Lock()
	defer table.Lock.Unlock()
	_, isExist := table.Map[*ptrCookie]
	if isExist {
		return nil
	}
	table.Map[*ptrCookie] = channel
	table.Queue.PushBack(QueueMember{ptrCookie, time.Now()})
	return ptrCookie
}

// Get the target Cookie channel. If not exist, return nil.
func (table *CookieTable) Get(ptrCookie *Cookie) chan<- *Datagram {
	table.Lock.Lock()
	defer table.Lock.Unlock()
	channel, isExist := table.Map[*ptrCookie]
	if !isExist {
		return nil
	}
	return channel
}
