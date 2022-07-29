package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc/peer"
	"io"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"time"

	log "github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/google/gnxi/gnmi"
	"github.com/hrk091/gnmi-fake/pkg/modeldata"
	"github.com/hrk091/gnmi-fake/pkg/modeldata/gostruct"

	"github.com/google/gnxi/utils/credentials"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

var (
	bindAddr   = flag.String("bind_address", ":9339", "Bind to address:port or just :port")
	configFile = flag.String("config", "", "IETF JSON file for target startup config")
)

type server struct {
	*gnmi.Server
	subChannels map[string]chan *pb.Notification
}

func newServer(model *gnmi.Model, config []byte) (*server, error) {
	s, err := gnmi.NewServer(model, config, nil)
	if err != nil {
		return nil, err
	}
	return &server{Server: s, subChannels: map[string]chan *pb.Notification{}}, nil
}

// Get overrides the Get func of gnmi.Target to provide user auth.
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Get request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Get request: %v", msg)
	return s.Server.Get(ctx, req)
}

// Set overrides the Set func of gnmi.Target to provide user auth.
func (s *server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	msg, ok := credentials.AuthorizeUser(ctx)
	if !ok {
		log.Infof("denied a Set request: %v", msg)
		return nil, status.Error(codes.PermissionDenied, msg)
	}
	log.Infof("allowed a Set request: %v", msg)
	resp, err := s.Server.Set(ctx, req)
	if err != nil {
		return nil, err
	}
	noti := &pb.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    req.GetPrefix(),
		Update:    append(req.GetUpdate(), req.GetReplace()...),
		Delete:    req.GetDelete(),
	}
	for _, ch := range s.subChannels {
		ch <- noti
	}
	return resp, err
}

// Subscribe overrides the Subscribe func of gnmi.Target to provide user auth.
func (s *server) Subscribe(stream pb.GNMI_SubscribeServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get peer from context")
	}

	id := pr.Addr.String()
	s.subChannels[id] = make(chan *pb.Notification)
	defer func() {
		cancel()
		delete(s.subChannels, id)
	}()
	log.Infof("new subscription: id=%d", id)

	if stream == nil {
		return status.Error(codes.FailedPrecondition, "cannot start client: stream is nil")
	}

	query, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Infof("Received initial query %v", query)

	subscribe := query.GetSubscribe()
	if subscribe == nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("first message must be SubscriptionList: %q", query))
	}

	go func() {
		for {
			log.Infof("send loop started")
			ch, ok := s.subChannels[id]
			if !ok {
				log.Infof("channel is already closed. send loop stopped")
				return
			}

			log.Info("waiting update notification...")
			select {
			case noti := <-ch:
				log.Infof("update notification: %v", noti)
				resp := &pb.SubscribeResponse{
					Response: &pb.SubscribeResponse_Update{
						Update: noti,
					},
				}
				if err := stream.Send(resp); err != nil {
					return
				}
			case <-ctx.Done():
				log.Info("cancelled. send loop stopped")
				return
			}
		}
	}()

	for {
		log.Infof("recv loop started")
		log.Info("waiting client request...")
		r, err := stream.Recv()

		log.Infof("client request received: %v, %v", r, err)
		if err == io.EOF {
			log.Infof("EOF")
			return nil
		} else if err != nil {
			log.Error(err)
			return err
		}
	}
}

func main() {
	model := gnmi.NewModel(modeldata.ModelData,
		reflect.TypeOf((*gostruct.Device)(nil)),
		gostruct.SchemaTree["Device"],
		gostruct.Unmarshal,
		gostruct.ΛEnum)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Supported models:\n")
		for _, m := range model.SupportedModels() {
			fmt.Fprintf(os.Stderr, "  %s\n", m)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Set("logtostderr", "true")
	flag.Parse()

	opts := credentials.ServerCredentials()
	g := grpc.NewServer(opts...)

	var configData []byte
	if *configFile != "" {
		var err error
		configData, err = ioutil.ReadFile(*configFile)
		if err != nil {
			log.Exitf("error in reading config file: %v", err)
		}
	}
	s, err := newServer(model, configData)
	if err != nil {
		log.Exitf("error in creating gnmi target: %v", err)
	}
	pb.RegisterGNMIServer(g, s)
	reflection.Register(g)

	log.Infof("starting to listen on %s", *bindAddr)
	listen, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Exitf("failed to listen: %v", err)
	}

	log.Info("starting to serve")
	if err := g.Serve(listen); err != nil {
		log.Exitf("failed to serve: %v", err)
	}
}
