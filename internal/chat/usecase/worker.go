package usecase

import (
	"context"
	"log"

	"video-call/internal/models"
)

type MessageWriterIface interface {
	Enqueue(msg models.Message)
}

type UseCase interface {
	CreateMessage(ctx context.Context, message models.Message) error
}

type MessageWriter struct {
	queue   chan models.Message
	nWorker int
	uc      UseCase
}

func NewMessageWriter(uc UseCase, workerNum int, queueSize int) *MessageWriter {
	mw := &MessageWriter{
		queue:   make(chan models.Message, queueSize),
		nWorker: workerNum,
		uc:      uc,
	}
	for i := 0; i < mw.nWorker; i++ {
		go mw.worker()
	}
	return mw
}

func (mw *MessageWriter) worker() {
	for msg := range mw.queue {
		if err := mw.uc.CreateMessage(context.Background(), msg); err != nil {
			log.Printf("[DB Worker] Failed to save message: %v", err)
		}
	}
}

func (mw *MessageWriter) Enqueue(msg models.Message) {
	mw.queue <- msg
}
