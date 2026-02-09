package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// BatchTelemetry represents MQTT batch telemetry data
type BatchTelemetry struct {
	BatchID    string      `json:"batch_id"`
	Timestamp  int64       `json:"timestamp"`
	Vehicles   []Telemetry `json:"vehicles"`
	BatchSize  int         `json:"batch_size"`
}

// TelemetryBatchSender handles batch telemetry sending
type TelemetryBatchSender struct {
	BatchSize    int
	BatchTimeout time.Duration
	batches      map[string][]Telemetry
	lastSend     map[string]time.Time
}

// NewTelemetryBatchSender creates a new batch sender
func NewTelemetryBatchSender(batchSize int, timeout time.Duration) *TelemetryBatchSender {
	return &TelemetryBatchSender{
		BatchSize:    batchSize,
		BatchTimeout: timeout,
		batches:      make(map[string][]Telemetry),
		lastSend:     make(map[string]time.Time),
	}
}

// AddTelemetry adds telemetry to batch
func (tbs *TelemetryBatchSender) AddTelemetry(topic string, telemetry Telemetry) (bool, *BatchTelemetry) {
	tbs.batches[topic] = append(tbs.batches[topic], telemetry)
	
	// Check if batch is full or timeout reached
	now := time.Now()
	lastSend, exists := tbs.lastSend[topic]
	
	batchReady := len(tbs.batches[topic]) >= tbs.BatchSize ||
		(exists && now.Sub(lastSend) >= tbs.BatchTimeout)
	
	if batchReady && len(tbs.batches[topic]) > 0 {
		batch := tbs.createBatch(topic)
		tbs.batches[topic] = nil
		tbs.lastSend[topic] = now
		return true, batch
	}
	
	return false, nil
}

// createBatch creates a batch telemetry message
func (tbs *TelemetryBatchSender) createBatch(topic string) *BatchTelemetry {
	telemetries := tbs.batches[topic]
	
	return &BatchTelemetry{
		BatchID:    generateBatchID(),
		Timestamp:  time.Now().Unix(),
		Vehicles:   telemetries,
		BatchSize:  len(telemetries),
	}
}

// generateBatchID generates a unique batch ID
func generateBatchID() string {
	return fmt.Sprintf("batch_%d", time.Now().UnixNano())
}

// SendBatchTelemetry sends batch telemetry via MQTT
func SendBatchTelemetry(client mqtt.Client, topic string, batch *BatchTelemetry) {
	data, err := json.Marshal(batch)
	if err != nil {
		log.Printf("Failed to marshal batch telemetry: %v", err)
		return
	}

	token := client.Publish(topic, 0, false, data)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to publish batch telemetry: %v", token.Error())
	}
}