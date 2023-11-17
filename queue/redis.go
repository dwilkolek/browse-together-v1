package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dwilkolek/browse-together-api/config"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dwilkolek/browse-together-api/dto"
	"github.com/redis/go-redis/v9"
)

const sessionCommunicationChannelPrefix string = "communication-"
const sessionPositionUpdatesChannelPrefix string = "position-"
const memberIdPrefix string = "memberId-"
const snapshotPrefix string = "snapshot-"

type RedisEventQueue struct {
	sessionId         string
	redisClient       *redis.Client
	sessionClosedChan chan struct{}
	cache             map[int64]dto.PositionStateDTO
	mu                sync.Mutex
	outdated          bool
	closed            bool
	initialized       bool
}

func (q *RedisEventQueue) Initialise() {

	q.cache = make(map[int64]dto.PositionStateDTO)
	var cacheTmp map[int64]dto.PositionStateDTO
	if snapshot, err := q.redisClient.Get(context.Background(), snapshotPrefix+q.sessionId).Result(); err == nil {
		if err = json.Unmarshal([]byte(snapshot), &cacheTmp); err == nil {
			q.cache = cacheTmp
		}
	}

	go func() {
		pubSub := q.redisClient.Subscribe(context.Background(), sessionPositionUpdatesChannelPrefix+q.sessionId)
		defer func(pubSub *redis.PubSub) {
			err := pubSub.Close()
			if err != nil {
				log.Default().Println("Error closing position changed channel", err)
			}
		}(pubSub)

		pubSubInternal := q.redisClient.Subscribe(context.Background(), sessionCommunicationChannelPrefix+q.sessionId)
		defer func(pubSubInternal *redis.PubSub) {
			err := pubSubInternal.Close()
			if err != nil {
				log.Default().Println("Error closing session communication channel", err)
			}
		}(pubSubInternal)

		persistCacheTicker := time.NewTicker(time.Minute).C
		subscriptionChannel := pubSub.Channel()
		subscriptionChannelInternal := pubSubInternal.Channel()

		for {
			select {
			case <-persistCacheTicker:
				{
					func() {
						q.mu.Lock()
						defer q.mu.Unlock()
						q.cache = validPositionStates(q.cache)
						if snapshot, err := json.Marshal(q.cache); err == nil {
							q.redisClient.Set(context.Background(), snapshotPrefix+q.sessionId, snapshot, time.Hour)
						}
					}()
				}
			case msg, ok := <-subscriptionChannel:
				{
					if !ok {
						continue
					}

					var positionState dto.PositionStateDTO
					err := json.Unmarshal([]byte(msg.Payload), &positionState)
					if err != nil {
						log.Printf("Failed to unmarshal PositionStateDTO: %s\n", err)
					}
					if config.DEBUG {
						fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
					}
					func() {
						q.mu.Lock()
						defer q.mu.Unlock()
						q.outdated = true
						q.cache[positionState.MemberId] = positionState
					}()
				}
			case msg, ok := <-subscriptionChannelInternal:
				{
					if !ok {
						break
					}
					if config.DEBUG {
						fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
					}
					parts := strings.Split(msg.Payload, ";")
					if parts[0] == "MEM_LEFT" {
						memberId, err := strconv.ParseInt(parts[1], 10, 64)
						if err != nil {
							log.Println("Failed to convert str to memberId", err)
							continue
						}
						func(memberId int64) {
							q.mu.Lock()
							defer q.mu.Unlock()
							delete(q.cache, memberId)
							q.outdated = true
						}(memberId)
						continue
					}
					if parts[0] == "CLOSED" {
						func() {
							q.mu.Lock()
							defer q.mu.Unlock()
							q.sessionClosedChan <- struct{}{}
							q.closed = true
							close(q.sessionClosedChan)
						}()
						return
					}
				}
			}

		}
	}()
}

func (q *RedisEventQueue) RefreshNeeded() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.outdated && !q.closed
}
func (q *RedisEventQueue) GetSnapshot() map[int64]dto.PositionStateDTO {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.outdated = false
	return q.cache
}
func (q *RedisEventQueue) SessionMemberPositionChange(update dto.PositionStateDTO) {
	if q.closed {
		return
	}
	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("Failed to marshal PositionStateDTO: %s\n", err)
	}
	q.redisClient.Publish(context.Background(), sessionPositionUpdatesChannelPrefix+q.sessionId, data)

}
func (q *RedisEventQueue) MemberLeft(memberId int64) {
	if q.closed {
		return
	}
	q.redisClient.Publish(context.Background(), sessionCommunicationChannelPrefix+q.sessionId, fmt.Sprintf("MEM_LEFT;%d", memberId))

}
func (q *RedisEventQueue) CloseSession() {
	if q.closed {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.redisClient.Publish(context.Background(), sessionCommunicationChannelPrefix+q.sessionId, "CLOSED")
}
func (q *RedisEventQueue) NextMemberId() int64 {
	memberId, err := q.redisClient.Incr(context.Background(), memberIdPrefix+q.sessionId).Result()
	if err != nil {
		panic(err)
	}
	return memberId
}

func (q *RedisEventQueue) OnSessionClosed() <-chan struct{} {
	return q.sessionClosedChan
}
