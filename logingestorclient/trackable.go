package logingestorclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/YasarKaan/go-kit/loggerutils"
)

type contextKey string

const trackingContextKey contextKey = "tracking_info"

type TrackingInfo struct {
	MissionId uuid.UUID
	StartTime time.Time
}

// ContextWithTrackingInfo inserts TrackingInfo into Go context.
func ContextWithTrackingInfo(ctx context.Context, info *TrackingInfo) context.Context {
	return context.WithValue(ctx, trackingContextKey, info)
}

// GetTrackingInfo extracts TrackingInfo from Go context.
func GetTrackingInfo(ctx context.Context) *TrackingInfo {
	if val, ok := ctx.Value(trackingContextKey).(*TrackingInfo); ok {
		return val
	}
	return nil
}

type Session struct {
	SessionId   uuid.UUID `json:"sessionId"`
	OperationId string    `json:"operationId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSession   bool      `json:"isSession"`
	StartTime   int64     `json:"startTime"` // Epoch milliseconds
	TotalStep   int       `json:"totalStep"`
	AuthHeader  string    `json:"authHeader"`
}

type MissionStep struct {
	OperationId string    `json:"operationId"`
	MissionId   uuid.UUID `json:"missionId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSession   bool      `json:"isSession"`
	StartTime   int64     `json:"startTime"`
	EndTime     int64     `json:"endTime"`
	StepNumber  int       `json:"stepNumber"`
	AuthHeader  string    `json:"authHeader"`
	Token       string    `json:"token,omitempty"`
}

type SessionAndMission struct {
	SessionId          uuid.UUID `json:"sessionId"`
	SessionName        string    `json:"sessionName"`
	SessionDescription string    `json:"sessionDescription"`
	IsSession          bool      `json:"isSession"`
	StartTime          int64     `json:"startTime"`
	TotalStep          int       `json:"totalStep"`
	MissionId          uuid.UUID `json:"missionId"`
	MissionName        string    `json:"missionName"`
	MissionDescription string    `json:"missionDescription"`
	EndTime            int64     `json:"endTime"`
	StepNumber         int       `json:"stepNumber"`
	OperationId        string    `json:"operationId"`
	AuthHeader         string    `json:"authHeader"`
	Token              string    `json:"token,omitempty"`
}

type UnfinalizedMissionStep struct {
	MissionId   uuid.UUID `json:"missionId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSession   bool      `json:"isSession"`
	StartTime   int64     `json:"startTime"`
	StepNumber  int       `json:"stepNumber"`
}

var (
	exchangeName string
	queueName    string
	routingKey   string
	conn         *amqp.Connection
	ch           *amqp.Channel
	connMutex    sync.Mutex

	notFinishedMissions = make(map[uuid.UUID]*UnfinalizedMissionStep)
	missionsMutex       sync.RWMutex
)

// Initialize initializes RabbitMQ connection, exchange, and queue bindings.
func Initialize(host string, port int, exchange, queue, rKey string) error {
	connMutex.Lock()
	defer connMutex.Unlock()

	exchangeName = exchange
	queueName = queue
	routingKey = rKey

	url := fmt.Sprintf("amqp://guest:guest@%s:%d/", host, port)
	var err error
	conn, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchange: direct, non-durable, auto-delete (matches Java: durable=false, autoDelete=true)
	err = ch.ExchangeDeclare(exchangeName, "direct", false, true, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue: non-durable, non-exclusive, non-auto-delete (matches Java: durable=false, autoDelete=false)
	_, err = ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue
	err = ch.QueueBind(queueName, routingKey, exchangeName, false, nil)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// CreateSessionAndMission publishes initial session details to RabbitMQ.
func CreateSessionAndMission(ctx context.Context, sessionName, sessionDescription string, totalStep int, missionName, missionDescription, authHeader, operationId string) error {
	connMutex.Lock()
	channelOk := ch != nil && !ch.IsClosed()
	connMutex.Unlock()

	if !channelOk {
		loggerutils.Error("RabbitMQ channel is not available")
		return fmt.Errorf("rabbitmq channel is not available")
	}

	info := GetTrackingInfo(ctx)
	var startTime int64
	if info != nil {
		startTime = info.StartTime.UnixNano() / int64(time.Millisecond)
	} else {
		startTime = time.Now().UnixNano() / int64(time.Millisecond)
	}

	sessionAndMission := &SessionAndMission{
		SessionId:          uuid.New(),
		SessionName:        sessionName,
		SessionDescription: sessionDescription,
		IsSession:          true,
		StartTime:          startTime,
		TotalStep:          totalStep,
		MissionId:          uuid.New(),
		MissionName:        missionName,
		MissionDescription: missionDescription,
		EndTime:            time.Now().UnixNano() / int64(time.Millisecond),
		StepNumber:         1,
		OperationId:        operationId,
		AuthHeader:         authHeader,
	}

	body, err := json.Marshal(sessionAndMission)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// StartTheMission creates an unfinalized mission tracking record and returns its UUID.
// Callers should save this tracking info in their context.
func StartTheMission(ctx context.Context, name, description string, stepNumber int) (uuid.UUID, context.Context) {
	missionId := uuid.New()
	startTime := time.Now()

	unfinalized := &UnfinalizedMissionStep{
		MissionId:   missionId,
		Name:        name,
		Description: description,
		IsSession:   false,
		StartTime:   startTime.UnixNano() / int64(time.Millisecond),
		StepNumber:  stepNumber,
	}

	missionsMutex.Lock()
	notFinishedMissions[missionId] = unfinalized
	missionsMutex.Unlock()

	newInfo := &TrackingInfo{
		MissionId: missionId,
		StartTime: startTime,
	}
	newCtx := ContextWithTrackingInfo(ctx, newInfo)

	return missionId, newCtx
}

// FinalizeTheMission retrieves mission metadata from context, finalizes it, and publishes it.
func FinalizeTheMission(ctx context.Context, operationId, authHeader string) error {
	connMutex.Lock()
	channelOk := ch != nil && !ch.IsClosed()
	connMutex.Unlock()

	if !channelOk {
		loggerutils.Error("RabbitMQ channel is not available")
		return fmt.Errorf("rabbitmq channel is not available")
	}

	info := GetTrackingInfo(ctx)
	if info == nil {
		return fmt.Errorf("no tracking info found in context")
	}

	missionId := info.MissionId

	missionsMutex.Lock()
	unfinalized, ok := notFinishedMissions[missionId]
	if ok {
		delete(notFinishedMissions, missionId)
	}
	missionsMutex.Unlock()

	if !ok {
		return fmt.Errorf("mission %s not found or already finalized", missionId)
	}

	missionStep := &MissionStep{
		OperationId: operationId,
		MissionId:   missionId,
		Name:        unfinalized.Name,
		Description: unfinalized.Description,
		IsSession:   unfinalized.IsSession,
		StartTime:   unfinalized.StartTime,
		EndTime:     time.Now().UnixNano() / int64(time.Millisecond),
		StepNumber:  unfinalized.StepNumber,
		AuthHeader:  authHeader,
	}

	body, err := json.Marshal(missionStep)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

// Close closes RabbitMQ channel and connection.
func Close() {
	connMutex.Lock()
	defer connMutex.Unlock()

	if ch != nil && !ch.IsClosed() {
		_ = ch.Close()
	}
	if conn != nil && !conn.IsClosed() {
		_ = conn.Close()
	}
}
