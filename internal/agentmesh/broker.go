package agentmesh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventBroker defines the interface for event mesh communication
type EventBroker interface {
	// Publish emits an event to a topic
	Publish(ctx context.Context, topic string, event Event) error

	// Subscribe registers a handler for topic events
	Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error)

	// Request sends a request and waits for response
	Request(ctx context.Context, topic string, payload interface{}, timeout time.Duration) (*Event, error)

	// SubscribeWithFilter subscribes with topic wildcards
	SubscribeWithFilter(ctx context.Context, pattern string, handler EventHandler) (Subscription, error)

	// Close shuts down the broker
	Close() error
}

// InternalBroker implements EventBroker using in-memory message passing
type InternalBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	history     []Event
	maxHistory  int
	closed      bool
}

type subscription struct {
	id      string
	topic   string
	pattern string
	handler EventHandler
	cancel  context.CancelFunc
}

// NewInternalBroker creates a new internal event broker
func NewInternalBroker() *InternalBroker {
	return &InternalBroker{
		subscribers: make(map[string][]*subscription),
		maxHistory:  1000,
	}
}

// Publish emits an event to a topic
func (b *InternalBroker) Publish(ctx context.Context, topic string, event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return fmt.Errorf("broker is closed")
	}

	// Set event metadata
	event.Topic = topic
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	// Store in history
	b.history = append(b.history, event)
	if len(b.history) > b.maxHistory {
		b.history = b.history[1:]
	}

	// Notify subscribers
	if subs, ok := b.subscribers[topic]; ok {
		for _, sub := range subs {
			go func(s *subscription, e Event) {
				if err := s.handler(ctx, e); err != nil {
					fmt.Printf("Error in event handler for topic %s: %v\n", topic, err)
				}
			}(sub, event)
		}
	}

	// Notify wildcard subscribers
	for _, subs := range b.subscribers {
		for _, sub := range subs {
			if sub.pattern != "" && matchesPattern(topic, sub.pattern) {
				go func(s *subscription, e Event) {
					if err := s.handler(ctx, e); err != nil {
						fmt.Printf("Error in wildcard handler for pattern %s: %v\n", s.pattern, err)
					}
				}(sub, event)
			}
		}
	}

	return nil
}

// Subscribe registers a handler for topic events
func (b *InternalBroker) Subscribe(ctx context.Context, topic string, handler EventHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return Subscription{}, fmt.Errorf("broker is closed")
	}

	_, cancel := context.WithCancel(ctx)
	sub := &subscription{
		id:      fmt.Sprintf("sub-%d", time.Now().UnixNano()),
		topic:   topic,
		handler: handler,
		cancel:  cancel,
	}

	b.subscribers[topic] = append(b.subscribers[topic], sub)

	return Subscription{
		ID:      sub.id,
		Topic:   topic,
		Handler: handler,
		Cancel:  cancel,
	}, nil
}

// Request sends a request and waits for response
func (b *InternalBroker) Request(ctx context.Context, topic string, payload interface{}, timeout time.Duration) (*Event, error) {
	responseCh := make(chan Event, 1)
	correlationID := fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Subscribe for response
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	sub, err := b.Subscribe(subCtx, topic+".response", func(ctx context.Context, event Event) error {
		if event.CorrelationID == correlationID {
			responseCh <- event
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe for response: %w", err)
	}
	defer sub.Cancel()

	// Publish request
	event := Event{
		CorrelationID: correlationID,
		Payload: map[string]interface{}{
			"request": payload,
		},
	}

	if err := b.Publish(ctx, topic, event); err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-responseCh:
		return &resp, nil
	case <-subCtx.Done():
		return nil, fmt.Errorf("request timed out after %v", timeout)
	}
}

// SubscribeWithFilter subscribes with topic wildcards
func (b *InternalBroker) SubscribeWithFilter(ctx context.Context, pattern string, handler EventHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return Subscription{}, fmt.Errorf("broker is closed")
	}

	_, cancel := context.WithCancel(ctx)
	sub := &subscription{
		id:      fmt.Sprintf("sub-%d", time.Now().UnixNano()),
		pattern: pattern,
		handler: handler,
		cancel:  cancel,
	}

	// Store under a special key for wildcard subscriptions
	b.subscribers["_wildcard_"+pattern] = append(b.subscribers["_wildcard_"+pattern], sub)

	return Subscription{
		ID:      sub.id,
		Topic:   pattern,
		Handler: handler,
		Cancel:  cancel,
	}, nil
}

// Close shuts down the broker
func (b *InternalBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	// Cancel all subscriptions
	for _, subs := range b.subscribers {
		for _, sub := range subs {
			sub.cancel()
		}
	}

	return nil
}

// matchesPattern checks if a topic matches a wildcard pattern
func matchesPattern(topic, pattern string) bool {
	// Simple wildcard matching: * matches any segment
	// e.g., "omnigraph.*" matches "omnigraph.state.changed"
	if pattern == "*" {
		return true
	}

	// Exact match
	if topic == pattern {
		return true
	}

	// Prefix wildcard
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(topic) >= len(prefix) && topic[:len(prefix)] == prefix
	}

	return false
}
