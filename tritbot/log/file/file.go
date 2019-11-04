package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes"
	"github.com/mhutchinson/tritter/tritbot/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	listenAddr = "localhost:50052"
)

var (
	logFile = flag.String("log_file", "/tmp/tritter.log", "file path for message log")
)

type fileLogger struct {
	log.UnimplementedLoggerServer
	f *os.File
}

// newFileLogger creates a fileLogger from the flags.
func newFileLogger() *fileLogger {
	// Open the log file for writing.
	f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		glog.Fatalf("could not open log file: %v", err)
	}
	return &fileLogger{f: f}
}

// Log implements log.LoggerServer.Log.
func (l *fileLogger) Log(ctx context.Context, in *log.LogRequest) (*log.LogResponse, error) {
	msg := in.GetMessage()
	t, err := ptypes.Timestamp(msg.GetTimestamp())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid timestamp: %v", err)
	}
	if _, err := l.f.WriteString(fmt.Sprintf("%v: [%v] %v\n", t.Format(time.RFC3339), msg.GetUser(), msg.GetMessage())); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to log message: %v", err)
	}

	return &log.LogResponse{}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	log.RegisterLoggerServer(s, newFileLogger())
	glog.Infof("Serving file logger on %v, writing log to %v", listenAddr, *logFile)
	if err := s.Serve(lis); err != nil {
		glog.Fatalf("failed to serve: %v", err)
	}
}
