package handlers

import (
	"context"
	"go-developer-course-shortener/internal/app/service"
	pb "go-developer-course-shortener/proto"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	service *service.Service
}

// GrpcHandler contains service for current Repository.
type GrpcHandler struct {
	service *service.Service
}

// NewGrpcHandler returns a new GrpcHandler for the Repository.
func NewGrpcHandler(service *service.Service) *ShortenerServer {
	return &ShortenerServer{service: service}
}

// TODO: add tests for all grpc handlers

func (s *ShortenerServer) AddBatch(ctx context.Context, in *pb.AddBatchRequest) (*pb.AddBatchResponse, error) {
	// TODO: implement it
	for i := 0; i < len(in.Links); i++ {

	}

	var response pb.AddBatchResponse
	return &response, nil
}

func (s *ShortenerServer) AddLinkJSON(ctx context.Context, in *pb.AddLinkJSONRequest) (*pb.AddLinkJSONResponse, error) {
	// TODO: implement it

	var response pb.AddLinkJSONResponse
	return &response, nil
}

func (s *ShortenerServer) AddLink(ctx context.Context, in *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	// TODO: implement it

	var response pb.AddLinkResponse
	return &response, nil
}

func (s *ShortenerServer) DeleteLink(ctx context.Context, in *pb.DeleteLinkRequest) (*pb.DeleteLinkResponse, error) {
	// TODO: implement it

	var response pb.DeleteLinkResponse
	return &response, nil
}

func (s *ShortenerServer) GetUserLinks(ctx context.Context, in *pb.GetUserLinksRequest) (*pb.GetUserLinksResponse, error) {
	// TODO: implement it

	var response pb.GetUserLinksResponse
	return &response, nil
}

func (s *ShortenerServer) GetOriginalByShort(ctx context.Context, in *pb.GetOriginalByShortRequest) (*pb.GetOriginalByShortResponse, error) {
	// TODO: implement it

	var response pb.GetOriginalByShortResponse
	return &response, nil
}

func (s *ShortenerServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// TODO: implement it

	var response pb.GetStatsResponse
	return &response, nil
}

func (s *ShortenerServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	// TODO: implement it

	var response pb.PingResponse
	return &response, nil
}
