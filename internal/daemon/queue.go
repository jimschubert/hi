package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type RequestType string

const (
	RequestTypeText      RequestType = "text"
	RequestTypeMultiline RequestType = "multiline"
	RequestTypeChoice    RequestType = "choice"
	RequestTypeConfirm   RequestType = "confirm"
	RequestTypeNotify    RequestType = "notify"
)

type PendingRequest struct {
	ID          string
	Type        RequestType
	AgentName   string
	Title       string
	Prompt      string
	DefaultVal  string
	Choices     []string
	MultiSelect bool
	CreatedAt   time.Time

	responseCh chan Response
}

type Response struct {
	TextValue    string
	BoolValue    bool
	ChoiceValues []string
	Cancelled    bool
	Error        error
}

// Queue is a non-blocking asynchronous queue
// see https://jsschools.com/golang/advanced-go-channel-patterns-for-building-robust-d/
// see https://webdevstation.com/posts/simple-queue-implementation-in-golang/
// see https://codezup.com/building-distributed-task-queue-go-goroutines-channels/
// see https://github.com/twiny/tinyq/blob/main/queue.go
type Queue struct {
	mu       sync.Mutex
	items    []*PendingRequest
	notifyCh chan struct{} // a "tick" on every enqueue/dequeue
}

func NewQueue() *Queue {
	return &Queue{notifyCh: make(chan struct{}, 8)}
}

func (q *Queue) Enqueue(ctx context.Context, req *PendingRequest) (Response, error) {
	req.ID = uuid.NewString()
	req.CreatedAt = time.Now()
	req.responseCh = make(chan Response, 1)

	q.mu.Lock()
	q.items = append(q.items, req)
	q.mu.Unlock()
	q.signal()

	select {
	case resp := <-req.responseCh:
		q.remove(req.ID)
		q.signal()
		return resp, resp.Error
	case <-ctx.Done():
		q.remove(req.ID)
		q.signal()
		return Response{Cancelled: true}, ctx.Err()
	}
}

func (q *Queue) Respond(id string, resp Response) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, r := range q.items {
		if r.ID == id {
			r.responseCh <- resp
			return true
		}
	}
	return false
}

func (q *Queue) Peek() []*PendingRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*PendingRequest, len(q.items))
	copy(out, q.items)
	return out
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *Queue) Notify() <-chan struct{} {
	return q.notifyCh
}

func (q *Queue) remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, r := range q.items {
		if r.ID == id {
			// NOTE: FIFO, as you'd expect from a notification-driven system
			q.items = append(q.items[:i], q.items[i+1:]...)
			return
		}
	}
}

func (q *Queue) signal() {
	select {
	case q.notifyCh <- struct{}{}:
	default:
	}
}
