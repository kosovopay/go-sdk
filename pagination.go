package kosovopay

import (
	"context"
	"encoding/json"
	"fmt"
)

// listEnvelope is the wire shape of every list response.
type listEnvelope struct {
	Object  string                   `json:"object"`
	Data    []map[string]interface{} `json:"data"`
	HasMore bool                     `json:"has_more"`
	URL     string                   `json:"url"`
}

// PageIter is a generic cursor-paginated iterator. Call Next repeatedly until
// it returns false; check Err() afterwards.
//
//	iter := client.Payments.Iter(ctx, params)
//	for iter.Next() {
//	    p := iter.Payment()
//	    _ = p
//	}
//	if err := iter.Err(); err != nil { ... }
type PageIter[T any] struct {
	ctx     context.Context
	client  *Client
	path    string
	query   map[string]string
	fromMap func(map[string]interface{}) T
	buf     []T
	pos     int
	hasMore bool
	started bool
	err     error
	lastID  string
}

// newPageIter constructs a PageIter. fromMap converts a raw JSON row to T.
func newPageIter[T any](ctx context.Context, c *Client, path string, query map[string]string, fromMap func(map[string]interface{}) T) *PageIter[T] {
	if query == nil {
		query = map[string]string{}
	}
	return &PageIter[T]{
		ctx:     ctx,
		client:  c,
		path:    path,
		query:   query,
		fromMap: fromMap,
		hasMore: true,
	}
}

// Next advances the iterator, fetching the next page when the buffer is
// exhausted. Returns false when all items have been yielded or an error occurs.
func (it *PageIter[T]) Next() bool {
	if it.err != nil {
		return false
	}
	it.pos++
	if it.pos-1 < len(it.buf) {
		return true
	}
	// Buffer exhausted — fetch next page if there is one.
	if it.started && !it.hasMore {
		return false
	}
	it.started = true
	if it.lastID != "" {
		it.query["starting_after"] = it.lastID
	}
	page, err := it.fetchPage()
	if err != nil {
		it.err = err
		return false
	}
	it.buf = page.items
	it.hasMore = page.hasMore
	it.lastID = page.lastID
	it.pos = 1
	return len(it.buf) > 0
}

// Value returns the current item. Must only be called after Next returns true.
func (it *PageIter[T]) Value() T {
	return it.buf[it.pos-1]
}

// Err returns any error encountered during iteration.
func (it *PageIter[T]) Err() error {
	return it.err
}

type pageResult[T any] struct {
	items   []T
	hasMore bool
	lastID  string
}

func (it *PageIter[T]) fetchPage() (pageResult[T], error) {
	resp, err := it.client.resty.R().
		SetContext(it.ctx).
		SetQueryParams(it.query).
		Get(apiPrefix + it.path)
	if err != nil {
		return pageResult[T]{}, fmt.Errorf("kosovopay: request failed: %w", err)
	}
	if resp.StatusCode() >= 400 {
		var envelope map[string]interface{}
		_ = json.Unmarshal(resp.Body(), &envelope)
		return pageResult[T]{}, mapError(envelope, resp.StatusCode(), 0)
	}
	var env listEnvelope
	if err := json.Unmarshal(resp.Body(), &env); err != nil {
		return pageResult[T]{}, fmt.Errorf("kosovopay: failed to decode list response: %w", err)
	}
	items := make([]T, 0, len(env.Data))
	lastID := ""
	for _, row := range env.Data {
		items = append(items, it.fromMap(row))
		if id, ok := row["id"].(string); ok && id != "" {
			lastID = id
		}
	}
	return pageResult[T]{items: items, hasMore: env.HasMore, lastID: lastID}, nil
}

// All collects every item across all pages into a slice. Use for small result
// sets; for large ones prefer the iterator to stream items.
func All[T any](iter *PageIter[T]) ([]T, error) {
	var out []T
	for iter.Next() {
		out = append(out, iter.Value())
	}
	return out, iter.Err()
}
