package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lovoo/goka"
)

var (
	BlockedUsersGroup goka.Group = "blocked-users"
)

type UserBlockEvent struct {
	BlockedUserId int
	IsBlocked     bool
}

type UserBlockEventCodec struct{}

func (uc *UserBlockEventCodec) Encode(value any) ([]byte, error) {
	if _, isUserBlock := value.(*UserBlockEvent); !isUserBlock {
		return nil, fmt.Errorf("expected type is *UserBlockEvent, received %T", value)
	}
	return json.Marshal(value)
}

func (uc *UserBlockEventCodec) Decode(data []byte) (any, error) {
	var (
		c   UserBlockEvent
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

type UserBlockValue struct {
	Blocks map[int]bool
}

type UserBlockValueCodec struct{}

func (uc *UserBlockValueCodec) Encode(value any) ([]byte, error) {
	if _, isUserBlock := value.(*UserBlockValue); !isUserBlock {
		return nil, fmt.Errorf("expected type is *UserBlockValue, received %T", value)
	}
	return json.Marshal(value)
}

func (uc *UserBlockValueCodec) Decode(data []byte) (any, error) {
	var (
		c   UserBlockValue
		err error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %v", err)
	}
	return &c, nil
}

func block(ctx goka.Context, msg interface{}) {
	var blockEvent *UserBlockEvent
	var ok bool
	var blockValue *UserBlockValue

	if val := ctx.Value(); val != nil {
		blockValue = val.(*UserBlockValue)
	} else {
		blockValue = &UserBlockValue{Blocks: make(map[int]bool)}
	}

	if blockEvent, ok = msg.(*UserBlockEvent); !ok || blockEvent == nil {
		return
	}

	if blockEvent.IsBlocked {
		blockValue.Blocks[blockEvent.BlockedUserId] = true
	} else {
		delete(blockValue.Blocks, blockEvent.BlockedUserId)
	}

	ctx.SetValue(blockValue)
	log.Printf("[proc] key: %s,  msg: %+v, data in group_table %v \n", ctx.Key(), blockEvent, blockValue)
}

func RunUserBlocker(brokers []string, usersToBlockStream goka.Stream) {
	g := goka.DefineGroup(BlockedUsersGroup,
		goka.Input(usersToBlockStream, new(UserBlockEventCodec), block),
		goka.Persist(new(UserBlockValueCodec)),
	)
	p, err := goka.NewProcessor(brokers, g)

	if err != nil {
		log.Fatal(err)
	}
	err = p.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
