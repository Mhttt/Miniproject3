package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "miniproject3/gRPC"

	"google.golang.org/grpc"
)

var addresses []string
var address string
var number int

func init() {
	addresses = []string{":9000", ":9001", ":9002"}
	rand.Seed(time.Now().UnixNano()) // Need this to generate random numbers
}

func main() {
	go connect()

	for {

	}
}

func connect() {

	address = addresses[rand.Intn(3)] 
	fmt.Printf("This is the random number: %d", number)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		connect()
	}

	c := pb.NewAuctionServiceClient(conn)
	defer conn.Close()
	log.Printf("Successfully connected to port: %v", address)
	fmt.Println("To bid on the auction enter the following: 'bid <name> <bid>'")
	fmt.Println("To get the current status of the auction enter: 'result'")

	for {
		var action string = ""
		var name string = ""
		var bid int64 = -1
		fmt.Scanln(&action, &name, &bid)

		if action == "bid" {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			r, err := c.Bid(ctx, &pb.BidRequest{Name: name, Bid: bid, TimeStamp: 0})
			if err == nil {

				if r.Status == 1 {
					log.Printf("%v Has successfully added a bid of: %d DKK", r.Name, r.Bid)
				}
				if r.Status == 2 {
					log.Printf("Failed to place bid. You must bid higher than the currentbid.")
				}
				if r.Status == 3 {
					log.Printf("The auction is closed.")
				}
				cancel()
			} else {
				log.Printf("Error: %v", err)
				cancel()
			}
		} else if action == "result" {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			r, err := c.Result(ctx, &pb.ResultRequest{TimeStamp: 0, Name: name})
			if err == nil {
				if !r.Ongoing {
					log.Printf("The auction is over! The winner was: %v , with a bid of %d DKK", r.Name, r.Bid)
					cancel()
				} else {
					log.Printf("The highest bid is currently %d from %v", r.Bid, r.Name)
					cancel()
				}
			}
		} else {
			log.Print("An error ocurred. Please enter a valid action")
		}
	}
}
