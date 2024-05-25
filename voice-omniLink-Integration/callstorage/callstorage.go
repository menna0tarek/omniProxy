package callstorage

import (
	"fmt"
	"sync"
	"time"

	"example.com/m/domain"
	"github.com/google/uuid"
)

type callStorage struct {
	calls          map[string]*Call
	callsFIFO      []*Call
	callCtrlManger chan string
	ttl            int
	sync.RWMutex
}

type Call struct {
	callId         string
	InboundChannel chan domain.CallbackRequestData
	RedirectURLMap map[string]string
	EndTime        time.Time
	isFinished     bool
	Attempts       int
}

func New(ttl int) *callStorage {
	cs := &callStorage{}
	cs.calls = make(map[string]*Call)
	cs.callsFIFO = make([]*Call, 0)
	cs.callCtrlManger = make(chan string)
	cs.ttl = ttl
	//	go cs.controlTTL() // TODO consider it back.
	return cs
}

func (cs *callStorage) Close() {
	close(cs.callCtrlManger)
}

func (ca *Call) ReplaceUrl(url string) string {
	token := uuid.NewString()
	ca.RedirectURLMap[token] = url
	return token
}

func (ca *Call) GetUrlByToken(token string) string {
	return ca.RedirectURLMap[token]
}

// Store the call in the call storage  once we receive a new request.
func (cs *callStorage) Store(callId string, mobile string, subId string) (*Call, error) {
	call := &Call{
		callId:         callId,
		InboundChannel: make(chan domain.CallbackRequestData, 1),
		RedirectURLMap: make(map[string]string),
		EndTime:        time.Now().Add(10 * time.Second),
		Attempts:       1,
	}

	cs.calls[callId] = call
	cs.callsFIFO = append(cs.callsFIFO, call)
	return call, nil
}

func (cs *callStorage) Get(callId string) (*Call, error) {
	if out, ok := cs.calls[callId]; ok && out != nil {
		return out, nil
	}
	return nil, nil
}

// Remove the call from the call storage after handling the callback.
func (cs *callStorage) Remove(callId string) error {
	if out, ok := cs.calls[callId]; ok && out != nil {
		out.isFinished = true
		close(out.InboundChannel)
	}
	return fmt.Errorf("Call with id %s not found", callId)
}

func RemoveIndex(s []*Call, index int) []*Call {
	return append(s[:index], s[index+1:]...)
}
func (cs *callStorage) controlTTL() {
	for {
		for i, c := range cs.callsFIFO {
			if time.Now().After(c.EndTime) {
				cs.Remove(c.callId)
				delete(cs.calls, c.callId)
				cs.Lock()
				cs.callsFIFO = RemoveIndex(cs.callsFIFO, i)
				cs.Unlock()
			} else {
				break
			}
		}
		time.Sleep(time.Duration(cs.ttl))
	}
}
