package handlers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go-developer-course-shortener/internal/app/repository"
	"go-developer-course-shortener/internal/app/service"
	"go-developer-course-shortener/internal/configs"
	"go-developer-course-shortener/internal/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"net/http"
	"net/url"
	"testing"

	pb "go-developer-course-shortener/proto"
)

func TestNewGrpcHandler(t *testing.T) {
	tests := []struct {
		name string
		want *ShortenerServer
	}{
		{name: "positive test",
			want: &ShortenerServer{service: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewGrpcHandler(nil), "NewGrpcHandler(%v)", nil)
		})
	}
}

func startGrpcServer() {

	storage := repository.NewInMemoryRepository()

	config, err := configs.ReadConfig()
	if err != nil {
		log.Fatalf("Failed to read server configuration. Error: %v", err.Error())
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup worker pool to handle delete requests
	jobs := make(chan worker.Job, worker.MaxWorkerPoolSize)
	workerPool := worker.NewWorkerPool(storage, jobs)
	go workerPool.Run(context.Background())

	network := net.IPNet{
		IP:   []byte("localhost:8080"),
		Mask: nil,
	}
	svc := service.NewService(storage, jobs, &network, config.BaseURL)

	var grpcSrv *grpc.Server
	go func() {
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
		if err != nil {
			log.Fatalf("GRPC server net.Listen: %v", err)
		}

		grpcSrv = grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptor))

		pb.RegisterShortenerServer(grpcSrv, NewGrpcHandler(svc))

		log.Printf("GRPC server started on %v", config.GrpcPort)
		// start grc server
		if err := grpcSrv.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()
}

func TestShortenerServer_All_Negative(t *testing.T) {
	// start client
	conn, err := grpc.Dial(":9999", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		assert.NoError(t, err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			assert.NoError(t, err)
		}
	}(conn)
	c := pb.NewShortenerClient(conn)
	ctx := context.WithValue(context.Background(), service.UserCtx, "userID")

	// Ping
	_, err = c.Ping(ctx, &pb.PingRequest{})
	assert.NotNil(t, err)

	// AddBatch
	var batchRequest pb.AddBatchRequest
	_, err = c.AddBatch(ctx, &batchRequest)
	assert.NotNil(t, err)

	// AddLinkJSON
	_, err = c.AddLinkJSON(ctx, &pb.AddLinkJSONRequest{Link: "https://github.com/test_repo1"})
	assert.NotNil(t, err)

	// AddLink
	_, err = c.AddLink(ctx, &pb.AddLinkRequest{Link: &pb.OriginalURL{OriginalUrl: "https://github.com/test_repo2"}})
	assert.NotNil(t, err)

	// DeleteLink
	var deleteRequest pb.DeleteLinkRequest
	_, err = c.DeleteLink(ctx, &deleteRequest)
	assert.NotNil(t, err)

	// GetUserLinks
	_, err = c.GetUserLinks(ctx, &pb.GetUserLinksRequest{})
	assert.NotNil(t, err)

	// GetOriginalByShort
	_, err = c.GetOriginalByShort(ctx, &pb.GetOriginalByShortRequest{Short: &pb.ShortURL{ShortUrl: "error_short_url"}})
	assert.NotNil(t, err)

}

func TestShortenerServer_All_Positive(t *testing.T) {
	t.Setenv("GRPC_PORT", "3201")
	// start grpc server
	startGrpcServer()

	// start client
	conn, err := grpc.Dial(":3201", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		assert.NoError(t, err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			assert.NoError(t, err)
		}
	}(conn)
	c := pb.NewShortenerClient(conn)
	ctx := context.WithValue(context.Background(), service.UserCtx, "userID")

	// Ping
	pingResponse, err := c.Ping(ctx, &pb.PingRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusOK), pingResponse.Code)

	// set metadata
	md := metadata.New(map[string]string{service.AccessToken: "userID"})
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	// AddBatch
	var batchRequest pb.AddBatchRequest
	batchRequest.Links = append(batchRequest.Links, &pb.RequestBatchJSON{
		Id:   &pb.CorrelationID{CorrelationId: "grpc_id1"},
		Orig: &pb.OriginalURL{OriginalUrl: "https://github.com/test_repo1"},
	})
	batchRequest.Links = append(batchRequest.Links, &pb.RequestBatchJSON{
		Id:   &pb.CorrelationID{CorrelationId: "grpc_id2"},
		Orig: &pb.OriginalURL{OriginalUrl: "https://github.com/test_repo2"},
	})

	batchResponse, err := c.AddBatch(ctx, &batchRequest)
	log.Printf("batchResponse: %v", batchResponse)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(batchResponse.Links))
	assert.Equal(t, int32(http.StatusCreated), batchResponse.Code)

	// AddLinkJSON
	linkJSONResponse, err := c.AddLinkJSON(ctx, &pb.AddLinkJSONRequest{Link: "https://github.com/test_repo3"})
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusCreated), linkJSONResponse.Code)

	// AddLink
	linkResponse, err := c.AddLink(ctx, &pb.AddLinkRequest{Link: &pb.OriginalURL{OriginalUrl: "https://github.com/test_repo4"}})
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusCreated), linkResponse.Code)

	// DeleteLink
	var deleteRequest pb.DeleteLinkRequest
	deleteRequest.Ids = append(deleteRequest.Ids, &pb.CorrelationID{
		CorrelationId: "grpc_id1",
	})
	deleteRequest.Ids = append(deleteRequest.Ids, &pb.CorrelationID{
		CorrelationId: "grpc_id2",
	})
	deleteLinkResponse, err := c.DeleteLink(ctx, &deleteRequest)
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusAccepted), deleteLinkResponse.Code)

	// GetUserLinks
	userLinksResponse, err := c.GetUserLinks(ctx, &pb.GetUserLinksRequest{})
	log.Printf("userLinksResponse: %v", userLinksResponse)
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusOK), userLinksResponse.Code)

	// GetOriginalByShort (positive test)
	shortURL, err := url.Parse(linkResponse.Short.ShortUrl)
	assert.NoError(t, err)
	link := shortURL.Path
	link = link[1:] // remove '/'
	origResponse, err := c.GetOriginalByShort(ctx, &pb.GetOriginalByShortRequest{Short: &pb.ShortURL{ShortUrl: link}})
	assert.NoError(t, err)
	assert.Equal(t, int32(http.StatusTemporaryRedirect), origResponse.Code)

	// GetOriginalByShort (negative test)
	_, err = c.GetOriginalByShort(ctx, &pb.GetOriginalByShortRequest{Short: &pb.ShortURL{ShortUrl: "error_short_url"}})
	assert.Error(t, err, status.Error(codes.InvalidArgument, err.Error()))

	// GetStats
	statsResponse, err := c.GetStats(ctx, &pb.GetStatsRequest{})
	log.Printf("userLinksResponse: %v", statsResponse)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), statsResponse.GetUsers())
	assert.Equal(t, int32(2), statsResponse.GetUrls())
	assert.Equal(t, int32(http.StatusOK), statsResponse.Code)
}
