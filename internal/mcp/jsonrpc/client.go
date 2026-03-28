package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
)

// RPCError is a JSON-RPC error object returned by the server.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("json-rpc error %d: %s", e.Code, e.Message)
}

type envelope struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RequestHandler handles server-initiated JSON-RPC requests.
type RequestHandler func(ctx context.Context, method string, params json.RawMessage) (result any, rpcErr *RPCError, handled bool)

// Client is a minimal transport-agnostic JSON-RPC 2.0 client.
type Client struct {
	reader         *bufio.Reader
	writer         io.Writer
	closer         io.Closer
	requestHandler RequestHandler
	nextID         int64
	writeMu        sync.Mutex
	mu             sync.Mutex
	pendingStr     map[string]chan envelope
	done           chan struct{}
	readErr        error
}

// NewClient constructs a JSON-RPC client and starts its read loop.
func NewClient(reader io.Reader, writer io.Writer, closer io.Closer, requestHandler RequestHandler) *Client {
	client := &Client{
		reader:         bufio.NewReaderSize(reader, 256*1024),
		writer:         writer,
		closer:         closer,
		requestHandler: requestHandler,
		pendingStr:     map[string]chan envelope{},
		done:           make(chan struct{}),
	}
	go client.readLoop()
	return client
}

// Call performs a request and decodes the result into result.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	id := atomic.AddInt64(&c.nextID, 1)
	responseCh := make(chan envelope, 1)
	idStr := fmt.Sprintf("%d", id)

	c.mu.Lock()
	c.pendingStr[idStr] = responseCh
	c.mu.Unlock()

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	}

	payload, err := json.Marshal(request)
	if err != nil {
		c.removePending(idStr)
		return fmt.Errorf("marshal request %s: %w", method, err)
	}

	c.writeMu.Lock()
	err = writeMessage(c.writer, payload)
	c.writeMu.Unlock()
	if err != nil {
		c.removePending(idStr)
		return fmt.Errorf("write request %s: %w", method, err)
	}

	select {
	case response := <-responseCh:
		if response.Error != nil {
			return response.Error
		}
		if result == nil || len(response.Result) == 0 {
			return nil
		}
		if err := json.Unmarshal(response.Result, result); err != nil {
			return fmt.Errorf("decode response for %s: %w", method, err)
		}
		return nil
	case <-ctx.Done():
		c.removePending(idStr)
		return ctx.Err()
	case <-c.done:
		c.removePending(idStr)
		if c.readErr != nil {
			return c.readErr
		}
		return io.EOF
	}
}

// Notify sends a notification without expecting a response.
func (c *Client) Notify(ctx context.Context, method string, params any) error {
	_ = ctx

	request := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		request["params"] = params
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal notification %s: %w", method, err)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := writeMessage(c.writer, payload); err != nil {
		return fmt.Errorf("write notification %s: %w", method, err)
	}
	return nil
}

// Close closes the underlying transport.
func (c *Client) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}

func (c *Client) readLoop() {
	defer close(c.done)

	for {
		payload, err := readMessage(c.reader)
		if err != nil {
			c.failPending(err)
			return
		}

		var response envelope
		if err := json.Unmarshal(payload, &response); err != nil {
			c.failPending(fmt.Errorf("decode json-rpc message: %w", err))
			return
		}
		// Server-initiated request (has method + id)
		if response.Method != "" && len(response.ID) > 0 {
			go c.handleIncomingRequest(response)
			continue
		}
		// Notification (has method but no id) — skip
		if len(response.ID) == 0 {
			continue
		}

		// Response to one of our requests — match by string-normalized ID
		idStr := strings.TrimSpace(string(response.ID))
		idStr = strings.Trim(idStr, `"`) // handle servers that return id as string

		c.mu.Lock()
		ch := c.pendingStr[idStr]
		delete(c.pendingStr, idStr)
		c.mu.Unlock()
		if ch != nil {
			ch <- response
			close(ch)
		}
	}
}

func (c *Client) failPending(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readErr = err
	for id := range c.pendingStr {
		delete(c.pendingStr, id)
	}
}

func (c *Client) removePending(idStr string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pendingStr, idStr)
}

func (c *Client) handleIncomingRequest(request envelope) {
	result := any(map[string]any{})
	var rpcErr *RPCError
	if c.requestHandler != nil {
		handledResult, handledErr, handled := c.requestHandler(context.Background(), request.Method, request.Params)
		if handled {
			result = handledResult
			rpcErr = handledErr
		} else {
			rpcErr = &RPCError{Code: -32601, Message: "method not found"}
		}
		if result == nil {
			result = map[string]any{}
		}
		if handledErr != nil {
			result = nil
		}
		c.respond(request.ID, result, rpcErr)
		return
	}
	c.respond(request.ID, nil, &RPCError{Code: -32601, Message: "method not found"})
}

func (c *Client) respond(id json.RawMessage, result any, rpcErr *RPCError) {
	response := map[string]any{"jsonrpc": "2.0", "id": rawID(id)}
	if rpcErr != nil {
		response["error"] = rpcErr
	} else {
		response["result"] = result
	}
	payload, err := json.Marshal(response)
	if err != nil {
		return
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = writeMessage(c.writer, payload)
}

func rawID(id json.RawMessage) any {
	if len(id) == 0 {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(id, &decoded); err != nil {
		return nil
	}
	return decoded
}
