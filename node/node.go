package main

import (
	"context"
	"log"
	"miniproject3/gRPC"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

//Client1 connects to a random server.
//Client1 sends a Bid grpc request to the connected server1.
//Server1 stores the highest bid and bidder
//The severs1´s client sends a bid request with the same data to server2 and server3, that it´s dialing on
//Sever2 and Server3 stores that data, only if the lamport time from server1 is larger than their own lamporttime. If theirs is larger send their own data back.

//15 and 17 do the same
//Når du kommer ind i node og skal bid brug r.name
//Send request = send data skal ind i sever delen af bid()

//Client1 sender Bid Michael 100
//Server 1 modtager

var Address string
var Addresses []string
var TimeStamp int64
var Name string
var HighestBid int64
var HighestBidder string
var Ongoing bool
var Nodes []gRPC.AuctionServiceClient
var Listening chan bool
var WaitForOwnRequest sync.WaitGroup

type Node struct {
	gRPC.UnimplementedAuctionServiceServer
}

func init() {
	Addresses = []string{":9000", ":9001", ":9002"}
	Listening = make(chan bool)
	Ongoing = true
	WaitForOwnRequest = sync.WaitGroup{}
}

func main() {
	go Listen(0)

	<-Listening
	go Connect()

	for {
		AuctionDuration(600) //Duration of the auction in seconds
		break
	}
}

func Listen(counter int) {
	TimeStamp++
	listen, err := net.Listen("tcp", Addresses[counter])
	if err != nil {
		counter++
		Listen(counter)
	} else {
		Address = Addresses[counter]
		Name, _ = os.Hostname()
		Name = Name + Address
		Listening <- true

		s := grpc.NewServer()
		gRPC.RegisterAuctionServiceServer(s, &Node{})
		log.Printf("Listening at port %v", listen.Addr())
		if err := s.Serve(listen); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}
}

//Server client should dial to every other Address expect itself
func Connect() {
	TimeStamp++
	retry := false
	for i, adr := range Addresses {
		if adr != Address {
			conn, err := grpc.Dial(adr, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(3*time.Second))

			if err != nil {
				retry = true //Error with the dialing
			} else {
				c := gRPC.NewAuctionServiceClient(conn)
				Nodes = append(Nodes, c)         //Store connections
				Addresses = remove(Addresses, i) //Remove connected Address from slice

				log.Println("Connected to localhost" + adr)
			}
		}
	}
	if retry {
		Connect()
	}
}


func (n *Node) Bid(ctx context.Context, request *gRPC.BidRequest) (*gRPC.BidResponse, error) {

	TimeStamp = CompareTimeStamp(request.TimeStamp)
	TimeStamp++

	if Ongoing {
		if request.Bid > HighestBid {
			HighestBid = request.Bid
			HighestBidder = request.Name
			log.Printf("-"+"%v is placing a bid of %d DKK "+LogTimestamp(), request.Name, request.Bid)
			SendBidData(ctx, request)
			return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 1}, nil
		} else if HighestBid >= request.Bid {
			return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 2}, nil
		}
	}
	return &gRPC.BidResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Status: 3}, nil
}

//Sender bid videre til de andre nodes
func SendBidData(ctx context.Context, request *gRPC.BidRequest) {
	WaitForOwnRequest.Add(1)
	WaitForResponses := sync.WaitGroup{}

	for _, node := range Nodes {
		WaitForResponses.Add(1)
		func() {
			TimeStamp++
			_, err := node.Bid(context.Background(), &gRPC.BidRequest{Name: request.Name, Bid: request.Bid, TimeStamp: request.TimeStamp})
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func (n *Node) Result(ctx context.Context, request *gRPC.ResultRequest) (*gRPC.ResultResponse, error) {

	TimeStamp = CompareTimeStamp(request.TimeStamp)
	TimeStamp++

	log.Printf("-"+"%v has requested for the result"+LogTimestamp(), request.Name)

	if Ongoing {
		return &gRPC.ResultResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Ongoing: true}, nil
	} else {
		return &gRPC.ResultResponse{Name: Name, Bid: HighestBid, TimeStamp: TimeStamp, Ongoing: false}, nil
	}
}

//fix me. Ved ikke hvordan dette skal laves
func SendResultData(ctx context.Context, request *gRPC.ResultRequest) {
	WaitForOwnRequest.Add(1)
	WaitForResponses := sync.WaitGroup{}

	for _, node := range Nodes {
		WaitForResponses.Add(1)
		func() {
			TimeStamp++
			response, err := node.Result(context.Background(), &gRPC.ResultRequest{Name: request.Name, TimeStamp: request.TimeStamp})
			if err != nil {
				log.Println(err)
			}
			if response.TimeStamp > TimeStamp {
				//fix
			}
		}()
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
	log.Println("**The auction has ended!**")
	log.Printf("**%v is the winner of the auction with a highest bid of %d**", HighestBidder, HighestBid)
}

func LogTimestamp() string {
	strTimestamp := strconv.FormatInt(TimeStamp, 10)
	return " ~ [" + strTimestamp + "]"
}

// Remove items from a slice
func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
