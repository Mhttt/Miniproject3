package main

import (
	"context"
	"log"
	"miniproject3/gRPC"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

var address string
var addresses []string
var TimeStamp int64
var Name string
var HighestBid int64
var HighestBidder string
var Ongoing bool
var duration chan (bool)

type Node struct {
	gRPC.UnimplementedAuctionServiceServer
}

func init() {
	addresses = []string{":9000", ":9001", ":9002"}
	Ongoing = true
	duration = make(chan bool)
}

func main() {
	go Listen(0)

	for {
		AuctionDuration(15) //Duration of the auction in seconds
		break
	}
}

func Listen(counter int) {
	TimeStamp++
	listen, err := net.Listen("tcp", addresses[counter])
	if err != nil {
		counter++
		Listen(counter)
	} else {
		address = addresses[counter]
		Name, _ = os.Hostname()
		Name = Name + address

		s := grpc.NewServer()
		gRPC.RegisterAuctionServiceServer(s, &Node{})
		log.Printf("Listening at port %v", listen.Addr())
		if err := s.Serve(listen); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}

}

func (n *Node) Bid(ctx context.Context, request *gRPC.BidRequest) (*gRPC.BidResponse, error) {

	TimeStamp = CompareTimeStamp(request.TimeStamp)
	TimeStamp++

	if Ongoing {
		if request.Bid > HighestBid {
			HighestBid = request.Bid
			HighestBidder = request.Name
			return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 1}, nil
		} else if HighestBid >= request.Bid {
			return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 2}, nil
		}
	}
	return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 3}, nil
}

func (n *Node) Result(ctx context.Context, request *gRPC.ResultRequest) (*gRPC.ResultResponse, error) {
	TimeStamp = CompareTimeStamp(request.TimeStamp)
	TimeStamp++

	if Ongoing {
		return &gRPC.ResultResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Ongoing: true}, nil
	} else {
		return &gRPC.ResultResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Ongoing: false}, nil
	}
}

func CompareTimeStamp(requestTimetamp int64) int64 {
	if requestTimetamp > TimeStamp {
		return requestTimetamp
	} else {
		return TimeStamp
	}
}

func AuctionDuration(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
	log.Print("The auction has ended!")
}
