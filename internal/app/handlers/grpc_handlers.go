package handlers

import (
	"context"
	"github.com/google/uuid"
	"go-developer-course-shortener/internal/app/rand"
	"go-developer-course-shortener/internal/app/service"
	"go-developer-course-shortener/internal/app/types"
	pb "go-developer-course-shortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
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

func (s *ShortenerServer) AddBatch(ctx context.Context, in *pb.AddBatchRequest) (*pb.AddBatchResponse, error) {
	userID := service.ExtractUserIDFromContext(ctx)
	request := in.Links

	batchLinks := make(types.BatchLinks, len(request)) // allocate required capacity for the links
	for i, v := range request {
		id := string(rand.GenerateRandom(shortLinkLength))
		shortURL := service.MakeShortURL(s.service.BaseURL, id)
		batchLinks[i] = types.BatchLink{CorrelationID: v.GetId().CorrelationId, ShortURL: shortURL, OriginalURL: v.GetOrig().OriginalUrl}
	}

	var res types.ResponseBatch
	res, err := s.service.SaveBatchURLS(userID, batchLinks)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var response pb.AddBatchResponse
	for _, v := range res {
		response.Links = append(response.Links, &pb.ResponseBatchJSON{
			Id:    &pb.CorrelationID{CorrelationId: v.CorrelationID},
			Short: &pb.ShortURL{ShortUrl: v.ShortURL},
		})
	}

	response.Code = int32(http.StatusCreated)
	return &response, nil
}

func (s *ShortenerServer) AddLinkJSON(ctx context.Context, in *pb.AddLinkJSONRequest) (*pb.AddLinkJSONResponse, error) {
	log.Printf("Long URL (AddLinkJSON): %v", in.GetLink())
	longURL, err := service.ParseURL(in.GetLink())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID := service.ExtractUserIDFromContext(ctx)

	id := string(rand.GenerateRandom(shortLinkLength))
	shortURL := service.MakeShortURL(s.service.BaseURL, id)
	log.Printf("Short URL (AddLinkJSON): %v", shortURL)

	err = s.service.SaveURL(userID, shortURL, longURL)
	code, err := service.CheckDBViolation(err)
	if err != nil {
		return nil, status.Error(codes.Code(code), err.Error())
	}

	if code == http.StatusConflict {
		shortURL, err = s.service.GetShortURLByOriginalURL(longURL)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.AddLinkJSONResponse{
		Code:   int32(code),
		Result: shortURL}, nil
}

func (s *ShortenerServer) AddLink(ctx context.Context, in *pb.AddLinkRequest) (*pb.AddLinkResponse, error) {
	log.Printf("Long URL (AddLink): %v", in.GetLink().OriginalUrl)
	longURL, err := service.ParseURL(in.GetLink().OriginalUrl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID := service.ExtractUserIDFromContext(ctx)

	id := string(rand.GenerateRandom(shortLinkLength))
	shortURL := service.MakeShortURL(s.service.BaseURL, id)
	log.Printf("Short URL (AddLink): %v", shortURL)

	err = s.service.SaveURL(userID, shortURL, longURL)
	code, err := service.CheckDBViolation(err)
	if err != nil {
		return nil, status.Error(codes.Code(code), err.Error())
	}

	if code == http.StatusConflict {
		shortURL, err = s.service.GetShortURLByOriginalURL(longURL)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.AddLinkResponse{
		Code:  int32(code),
		Short: &pb.ShortURL{ShortUrl: shortURL},
	}, nil
}

func (s *ShortenerServer) DeleteLink(ctx context.Context, in *pb.DeleteLinkRequest) (*pb.DeleteLinkResponse, error) {
	userID := service.ExtractUserIDFromContext(ctx)
	log.Printf("Delete all links for userID (DeleteLink): %s", userID)

	shortURLS := make([]string, len(in.Ids)) // allocate required capacity for the links
	for i, id := range in.Ids {
		shortURLS[i] = service.MakeShortURL(s.service.BaseURL, id.CorrelationId)
	}
	log.Printf("Request shortURLS (DeleteLink): %+v", shortURLS)

	err := s.service.DeleteURLS(userID, shortURLS)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteLinkResponse{Code: http.StatusAccepted}, nil
}

func (s *ShortenerServer) GetUserLinks(ctx context.Context, in *pb.GetUserLinksRequest) (*pb.GetUserLinksResponse, error) {
	userID := service.ExtractUserIDFromContext(ctx)

	log.Printf("Get all links for userID (GetUserLinks): %s", userID)
	links, err := s.service.GetUserStorage(userID)
	if err != nil {
		return &pb.GetUserLinksResponse{Code: int32(http.StatusNoContent), Links: nil}, err
	}

	var response pb.GetUserLinksResponse

	for _, v := range links {
		response.Links = append(response.Links, &pb.Link{
			Short: &pb.ShortURL{ShortUrl: v.ShortURL},
			Orig:  &pb.OriginalURL{OriginalUrl: v.OriginalURL},
		})
	}

	response.Code = int32(http.StatusOK)
	return &response, nil
}

func (s *ShortenerServer) GetOriginalByShort(ctx context.Context, in *pb.GetOriginalByShortRequest) (*pb.GetOriginalByShortResponse, error) {

	strID := in.Short.ShortUrl
	log.Printf("ShortUrl (GetOriginalByShort): `%s`", strID)

	originalLink, err := s.service.GetURL(service.MakeShortURL(s.service.BaseURL, strID))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	log.Printf("Original URL (GetOriginalByShort): %s deleted: %v", originalLink.OriginalURL, originalLink.Deleted)

	var code int32
	if !originalLink.Deleted {
		code = http.StatusTemporaryRedirect
	} else {
		code = http.StatusGone
	}

	return &pb.GetOriginalByShortResponse{
		Code: code,
		Link: &pb.OriginalLink{
			Orig:    &pb.OriginalURL{OriginalUrl: originalLink.OriginalURL},
			Deleted: originalLink.Deleted,
		},
	}, nil
}

func (s *ShortenerServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// get user ip (check "X-Real-IP" metadata)
	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			token = values[0]
		}
	}
	userIP := net.ParseIP(token)

	stats, err := s.service.GetInternalStats(userIP)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	return &pb.GetStatsResponse{
		Code:  int32(http.StatusOK),
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}

func (s *ShortenerServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	userID := service.ExtractUserIDFromContext(ctx)
	log.Printf("userID (Ping): %v\n", userID)

	var code int32
	if !s.service.Ping() {
		code = http.StatusInternalServerError
	} else {
		code = http.StatusOK
	}

	return &pb.PingResponse{Code: code}, nil
}

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	userID := uuid.NewString()

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(service.AccessToken)
		if len(values) > 0 {
			userID = values[0]
			log.Printf("UnaryInterceptor userID from context: '%s'", userID)
		}
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md.Append(service.AccessToken, string(userID))
	}
	newCtx := metadata.NewIncomingContext(ctx, md)

	return handler(newCtx, req)
}
