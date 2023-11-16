package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/dwilkolek/browse-together-api/dto"
	"github.com/redis/go-redis/v9"
)

const internalCommsPrefix string = "internal-"
const communicationChannelPrefix string = "pubsub-"
const memberIdPrefix string = "memberId-"

type RedisEventQueue struct {
	sessionId         string
	redisClient       *redis.Client
	sessionClosedChan chan struct{}
	cache             map[int64]dto.PositionStateDTO
	mu                sync.Mutex
	outdated          bool
	closed            bool
}

func (q *RedisEventQueue) Initalize() {
	fmt.Println("Init")
	go func() {
		fmt.Println("Init 2")
		// Subscribe to the channel
		pubSub := q.redisClient.Subscribe(context.Background(), communicationChannelPrefix+q.sessionId)
		defer pubSub.Close()

		pubSubInternal := q.redisClient.Subscribe(context.Background(), internalCommsPrefix+q.sessionId)
		defer pubSubInternal.Close()

		// Channel to receive subscription messages
		subscriptionChannel := pubSub.Channel()

		subscriptionChannelInternal := pubSubInternal.Channel()
		// Start listening for messages
		for {
			fmt.Println("Something happens")
			select {
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

					fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
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
					fmt.Printf("Received message from %s: %s\n", msg.Channel, msg.Payload)
					parts := strings.Split(msg.Payload, ";")
					if parts[0] == "MEM_LEFT" {
						memberId, err := strconv.ParseInt(parts[1], 10, 64)
						if err != nil {
							log.Println("Failed to conver str to memberId", err)
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
	q.redisClient.Publish(context.Background(), communicationChannelPrefix+q.sessionId, data)

}
func (q *RedisEventQueue) MemberLeft(memberId int64) {
	if q.closed {
		return
	}
	q.redisClient.Publish(context.Background(), internalCommsPrefix+q.sessionId, fmt.Sprintf("MEM_LEFT;%d", memberId))

}
func (q *RedisEventQueue) CloseSession() {
	if q.closed {
		return
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.redisClient.Publish(context.Background(), internalCommsPrefix+q.sessionId, "CLOSED")
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
