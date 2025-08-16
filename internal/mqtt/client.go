package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler defines the function signature for handling incoming messages
type MessageHandler func(payload []byte)

// ClientInterface defines the interface for MQTT clients (useful for testing)
type ClientInterface interface {
	Connect(ctx context.Context) error
	Disconnect()
	IsConnected() bool
	Publish(ctx context.Context, topic string, payload []byte) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(ctx context.Context, topic string) error
}

// Client wraps MQTT functionality with explicit configuration
type Client struct {
	host      string
	port      int
	clientID  string
	client    pahomqtt.Client
	connected bool
	mu        sync.RWMutex
}

// ClientOptions configures MQTT client behavior
type ClientOptions struct {
	KeepAlive            time.Duration
	ConnectTimeout       time.Duration
	ReconnectBackoff     time.Duration
	MaxReconnectInterval time.Duration
}

// DefaultClientOptions returns sensible defaults
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		KeepAlive:            30 * time.Second,
		ConnectTimeout:       10 * time.Second,
		ReconnectBackoff:     1 * time.Second,
		MaxReconnectInterval: 30 * time.Second,
	}
}

// NewClient creates a new MQTT client with explicit configuration
func NewClient(host string, port int) *Client {
	clientID := fmt.Sprintf("mqtt-worker-%d", time.Now().UnixNano())
	return NewClientWithID(host, port, clientID)
}

// NewClientWithID creates a new MQTT client with explicit client ID
func NewClientWithID(host string, port int, clientID string) *Client {
	return &Client{
		host:     host,
		port:     port,
		clientID: clientID,
	}
}

// Connect establishes connection to MQTT broker
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return early if already connected
	if c.connected && c.client.IsConnected() {
		return nil
	}

	opts := DefaultClientOptions()

	// Create MQTT client options
	clientOpts := pahomqtt.NewClientOptions()
	clientOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", c.host, c.port))
	clientOpts.SetClientID(c.clientID)
	clientOpts.SetKeepAlive(opts.KeepAlive)
	clientOpts.SetConnectTimeout(opts.ConnectTimeout)
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetMaxReconnectInterval(opts.MaxReconnectInterval)

	// Connection lost handler
	clientOpts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	})

	// On connect handler
	clientOpts.SetOnConnectHandler(func(client pahomqtt.Client) {
		c.mu.Lock()
		c.connected = true
		c.mu.Unlock()
	})

	c.client = pahomqtt.NewClient(clientOpts)

	// Connect with context timeout
	done := make(chan error, 1)
	go func() {
		token := c.client.Connect()
		token.Wait()
		done <- token.Error()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to connect to MQTT broker at %s:%d: %w", c.host, c.port, err)
		}
		c.connected = true
		return nil
	case <-ctx.Done():
		return fmt.Errorf("connection timeout: %w", ctx.Err())
	}
}

// Disconnect closes the MQTT connection
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250) // 250ms timeout for graceful disconnect
	}
	c.connected = false
}

// IsConnected returns the current connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected && c.client != nil && c.client.IsConnected()
}

// Publish sends a message to the specified topic
func (c *Client) Publish(ctx context.Context, topic string, payload []byte) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	// Use QoS 1 for reliable delivery
	const qos = 1
	const retained = false

	done := make(chan error, 1)
	go func() {
		token := c.client.Publish(topic, qos, retained, payload)
		token.Wait()
		done <- token.Error()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("publish timeout: %w", ctx.Err())
	}
}

// Subscribe registers a handler for messages on the specified topic
func (c *Client) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	// Use QoS 1 for reliable delivery
	const qos = 1

	messageHandler := func(client pahomqtt.Client, msg pahomqtt.Message) {
		handler(msg.Payload())
	}

	done := make(chan error, 1)
	go func() {
		token := c.client.Subscribe(topic, qos, messageHandler)
		token.Wait()
		done <- token.Error()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("subscribe timeout: %w", ctx.Err())
	}
}

// Unsubscribe removes subscription from the specified topic
func (c *Client) Unsubscribe(ctx context.Context, topic string) error {
	if !c.IsConnected() {
		return fmt.Errorf("client not connected")
	}

	done := make(chan error, 1)
	go func() {
		token := c.client.Unsubscribe(topic)
		token.Wait()
		done <- token.Error()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("unsubscribe timeout: %w", ctx.Err())
	}
}
