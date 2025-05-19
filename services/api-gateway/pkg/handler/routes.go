package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

// AnalysisRequest modeli
type AnalysisRequest struct {
	WebsocketKlineOptions struct {
		Symbol   string `json:"symbol" binding:"required"`
		Interval string `json:"interval" binding:"required"`
		Exchange string `json:"exchange" binding:"required"`
	} `json:"websocketKlineOptions"`
	Indicators []struct {
		Indicator  string                 `json:"indicator" binding:"required"`
		Parameters map[string]interface{} `json:"parameters"`
		Operator   string                 `json:"operator" binding:"required"`
		Threshold  float64                `json:"threshold" binding:"required"`
	} `json:"indicators" binding:"required,dive"`
}

// Handler tutacağı Kafka writer ve topic
type Handler struct {
	writer *kafka.Writer
	topic  string
}

// RegisterRoutes Gin router’ına endpoint’leri ekler
func RegisterRoutes(r *gin.Engine, writer *kafka.Writer, topic string) {
	h := &Handler{writer: writer, topic: topic}

	// Healthz
	r.GET("/healthz", h.healthz)

	// StreamAnalysis
	r.POST("/streamanalysis", h.streamAnalysis)
}

func (h *Handler) healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *Handler) streamAnalysis(c *gin.Context) {
	// 1) JSON bind & validation
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[streamAnalysis] bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[streamAnalysis] received: %+v", req)

	// 2) Marshal ve Kafka’ya publish
	payload, _ := json.Marshal(req)
	msg := kafka.Message{Value: payload}
	if err := h.writer.WriteMessages(context.Background(), msg); err != nil {
		log.Printf("[streamAnalysis] kafka publish error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "publish failed"})
		return
	}
	log.Printf("[streamAnalysis] published to topic %s", h.topic)

	// 3) Client’a cevap
	c.JSON(http.StatusAccepted, gin.H{"status": "processing"})
}
