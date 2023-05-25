package model

import (
	audit "github.com/GalushkoArt/GoAuditService/pkg/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Actions string

type Entities string

const (
	ACTION_SIGN_UP Actions = "SIGN_UP"
	ACTION_SIGN_IN Actions = "SIGN_IN"
	ACTION_REFRESH Actions = "REFRESH"
	ACTION_GET     Actions = "GET"
	ACTION_CREATE  Actions = "CREATE"
	ACTION_UPDATE  Actions = "UPDATE"
	ACTION_DELETE  Actions = "DELETE"

	ENTITY_USER   Entities = "USER"
	ENTITY_SYMBOL Entities = "SYMBOL"
)

var (
	itemActionsToRequest = map[Actions]audit.LogRequest_Actions{
		ACTION_SIGN_UP: audit.LogRequest_SIGN_UP,
		ACTION_SIGN_IN: audit.LogRequest_SIGN_IN,
		ACTION_REFRESH: audit.LogRequest_REFRESH,
		ACTION_CREATE:  audit.LogRequest_CREATE,
		ACTION_UPDATE:  audit.LogRequest_UPDATE,
		ACTION_GET:     audit.LogRequest_GET,
		ACTION_DELETE:  audit.LogRequest_DELETE,
	}
	itemEntityToRequest = map[Entities]audit.LogRequest_Entities{
		ENTITY_USER:   audit.LogRequest_USER,
		ENTITY_SYMBOL: audit.LogRequest_SYMBOL,
	}
	requestActionsToItem = map[audit.LogRequest_Actions]Actions{
		audit.LogRequest_SIGN_UP: ACTION_SIGN_UP,
		audit.LogRequest_SIGN_IN: ACTION_SIGN_IN,
		audit.LogRequest_REFRESH: ACTION_REFRESH,
		audit.LogRequest_CREATE:  ACTION_CREATE,
		audit.LogRequest_UPDATE:  ACTION_UPDATE,
		audit.LogRequest_GET:     ACTION_GET,
		audit.LogRequest_DELETE:  ACTION_DELETE,
	}
	requestEntityToItem = map[audit.LogRequest_Entities]Entities{
		audit.LogRequest_USER:   ENTITY_USER,
		audit.LogRequest_SYMBOL: ENTITY_SYMBOL,
	}
)

type LogItem struct {
	Action    Actions   `bson:"action"`
	Entity    Entities  `bson:"entity"`
	EntityID  string    `bson:"entityID"`
	Timestamp time.Time `bson:"timestamp"`
}

func LogRequestToItem(request *audit.LogRequest) LogItem {
	return LogItem{
		Action:    requestActionsToItem[request.Action],
		Entity:    requestEntityToItem[request.Entity],
		EntityID:  request.EntityId,
		Timestamp: request.Timestamp.AsTime(),
	}
}

func LogItemToLogRequest(item *LogItem) audit.LogRequest {
	return audit.LogRequest{
		Action:    itemActionsToRequest[item.Action],
		Entity:    itemEntityToRequest[item.Entity],
		EntityId:  item.EntityID,
		Timestamp: timestamppb.New(item.Timestamp),
	}
}
