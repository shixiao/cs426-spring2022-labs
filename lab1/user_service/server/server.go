package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"

	"cs426.yale.edu/lab1/failure_injection"
	fi "cs426.yale.edu/lab1/failure_injection/proto"
	pb "cs426.yale.edu/lab1/user_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gofakeit "github.com/brianvoe/gofakeit/v6"
)

var (
	port        = flag.Int("port", 8081, "The server port")
	seed        = flag.Int("seed", 42, "Random seed for generating database data")
	sleepNs     = flag.Int64("sleep-ns", 0, "Injected latency on each request")
	failureRate = flag.Int64(
		"failure-rate",
		0,
		"Injected failure rate N (0 means no injection; o/w errors one in N requests",
	)
	responseOmissionRate = flag.Int64(
		"response-omission-rate",
		0,
		"Injected response omission rate N (0 means no injection; o/w errors one in N requests",
	)
	maxBatchSize = flag.Int(
		"batch-size",
		50,
		"Maximum size of batches accepted",
	)
)

const USER_ID_OFFSET = 200000
const VIDEO_ID_OFFSET = 1000

func makeRandomUser(userId uint64, maxUsers int, maxVideos int) *pb.UserInfo {
	user := new(pb.UserInfo)

	user.UserId = userId
	user.Email = gofakeit.Email()
	user.ProfileUrl = fmt.Sprintf("https://user-service.localhost/profile/%d", userId)
	user.Username = gofakeit.Username()

	coeffCount := rand.Intn(10) + 10
	user.UserCoefficients = new(pb.UserCoefficients)
	user.UserCoefficients.Coeffs = make(map[int32]uint64)
	for i := 0; i < coeffCount; i++ {
		user.UserCoefficients.Coeffs[int32(rand.Intn(20))] = uint64(rand.Intn(500))
	}

	subscribeCount := rand.Intn(50) + 10
	for i := 0; i < subscribeCount; i++ {
		subscribed := uint64(rand.Intn(maxUsers) + USER_ID_OFFSET)
		if subscribed != userId {
			user.SubscribedTo = append(user.SubscribedTo, subscribed)
		}
	}

	likeCount := rand.Intn(20) + 5
	for i := 0; i < likeCount; i++ {
		liked := uint64(rand.Intn(maxVideos) + VIDEO_ID_OFFSET)
		user.LikedVideos = append(user.LikedVideos, liked)
	}
	return user
}

func makeRandomUsers() map[uint64]*pb.UserInfo {
	videoCount := rand.Intn(500) + 100
	userCount := rand.Intn(5000) + 5000

	users := make(map[uint64]*pb.UserInfo)

	for userId := USER_ID_OFFSET; userId < USER_ID_OFFSET+userCount; userId++ {
		users[uint64(userId)] = makeRandomUser(uint64(userId), userCount, videoCount)
	}
	return users
}

type userServiceServer struct {
	pb.UnimplementedUserServiceServer
	users map[uint64]*pb.UserInfo
}

func MakeUserServiceServer() *userServiceServer {
	return &userServiceServer{
		users: makeRandomUsers(),
	}
}

func (db *userServiceServer) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	shouldError := failure_injection.MaybeInject()
	if shouldError {
		return nil, status.Error(
			codes.Internal,
			"UserService: (injected) internal error!",
		)
	}

	userIds := req.GetUserIds()
	if userIds == nil || len(userIds) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"UserService: user_ids in GetUserRequest should not be empty",
		)
	}
	if len(userIds) > *maxBatchSize {
		return nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("UserService: user_ids exceeded the max batch size %d", *maxBatchSize),
		)
	}
	users := make([]*pb.UserInfo, 0, len(userIds))
	for _, userId := range req.GetUserIds() {
		info, ok := db.users[userId]
		if ok {
			users = append(users, info)
		} else {
			return nil, status.Error(codes.NotFound, fmt.Sprintf(
				"UserService: user %d cannot be found, it may have been deleted or never existed in the first place.",
				userId,
			))
		}
	}
	return &pb.GetUserResponse{Users: users}, nil
}

func (db *userServiceServer) SetInjectionConfig(
	ctx context.Context,
	req *fi.SetInjectionConfigRequest,
) (*fi.SetInjectionConfigResponse, error) {
	failure_injection.SetInjectionConfigPb(req.Config)
	return &fi.SetInjectionConfigResponse{}, nil
}

func main() {
	flag.Parse()
	rand.Seed(int64(*seed))
	gofakeit.Seed(int64(*seed))
	failure_injection.SetInjectionConfig(*sleepNs, *failureRate, *responseOmissionRate)
	fiConfig := failure_injection.GetInjectionConfig()
	log.Printf(
		"failure injection config: [sleepNs: %d, failureRate: %d, responseOmissionRate: %d",
		fiConfig.SleepNs,
		fiConfig.FailureRate,
		fiConfig.ResponseOmissionRate,
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, MakeUserServiceServer())
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
