package rpcserver

import (
	"time"

	"github.com/Ermi9s/Anubis/model"
	"github.com/Ermi9s/Anubis/repository"
)





type RpcServer struct {
	Repository *repository.Repository
}

func NewRpcServer(repository *repository.Repository) *RpcServer {
	return &RpcServer{
		Repository: repository,
	}
}

type Args struct {
	Action      *string
	Status      *string
	ActorID     *string
	ActorType   *string
	StartTime   *time.Time
	EndTime     *time.Time
	ServiceName *string
	Page        int
	PageSize    int
	SortBy      string
	SortOrder   string
}

type Response struct {
	Data       []model.AuditEvent 
	Pagination model.Pagination 
}


func (rpc *RpcServer) FindLog(args *Args, reponse *Response) error {
	filter := model.AuditEventFilter{
		Action: args.Action,
		Status: args.Status,
		ActorID: args.ActorID,
		ActorType: args.ActorType,
		StartTime: args.StartTime,
		EndTime: args.EndTime,
		ServiceName: args.ServiceName,
		Page: args.Page,
		PageSize: args.PageSize,
		SortBy: args.SortBy,
		SortOrder: args.SortOrder,
	}

	paginatedResponse, err := rpc.Repository.FindAudit(filter)
	if err != nil {
		return err
	}
	
	reponse.Data = paginatedResponse.Data
	reponse.Pagination = paginatedResponse.Pagination

	return  nil
}


