package cache

import (
	"context"
	"time"
)

type asyncRequest struct {
	ctx     context.Context
	key     string
	payload []byte
	ttl     time.Duration
}

func (s *Service) runAsyncWorker(id int) {
	defer s.wg.Done()
	for req := range s.asyncQueue {
		if !s.Enabled() {
			continue
		}
		if req.ctx == nil {
			req.ctx = context.Background()
		}
		if err := s.client.Set(req.ctx, req.key, req.payload, req.ttl).Err(); err != nil {
			s.logger.Warn("async cache set failed",
				"key", req.key,
				"error", err,
				"worker", id)
		}
	}
}
