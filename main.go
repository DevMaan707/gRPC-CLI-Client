package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gRPC/proto/gRPC/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const grpcAddress = "localhost:50052"

func main() {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Go gRPC server: %v", err)
	}
	defer conn.Close()

	client := proto.NewStorageServiceClient(conn)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nSelect Operation:")
		fmt.Println("1. Set")
		fmt.Println("2. Get")
		fmt.Println("3. BulkGet")
		fmt.Println("4. Display Schema")
		fmt.Println("5. Stress Test (Batch Requests)")
		fmt.Println("0. Exit")

		fmt.Print("Enter your choice: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			doInteractiveSet(client, reader)
		case "2":
			doInteractiveGet(client, reader)
		case "3":
			doInteractiveBulkGet(client, reader)
		case "4":
			displaySchema()
		case "5":
			stressTest(client, reader)
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}

func doInteractiveSet(client proto.StorageServiceClient, reader *bufio.Reader) {
	fmt.Print("Enter Key: ")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	fmt.Print("Enter Value: ")
	value, _ := reader.ReadString('\n')
	value = strings.TrimSpace(value)

	req := &proto.SetRequest{
		Key: key,
		Value: &proto.SetRequest_StringValue{
			StringValue: value,
		},
		Metadata: &proto.Metadata{
			Tags: map[string]string{"tag1": "example-tag"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.Set(ctx, req)
	latency := time.Since(start)

	if err != nil {
		log.Fatalf("Set request failed: %v", err)
	}

	fmt.Printf("Set Response: Success=%v, Latency=%v\n", resp.Success, latency)
}

func doInteractiveGet(client proto.StorageServiceClient, reader *bufio.Reader) {
	fmt.Print("Enter Key: ")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	req := &proto.GetRequest{
		Key: key,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.Get(ctx, req)
	latency := time.Since(start)

	if err != nil {
		log.Fatalf("Get request failed: %v", err)
	}

	var value string
	switch v := resp.Value.(type) {
	case *proto.GetResponse_StringValue:
		value = v.StringValue
	case *proto.GetResponse_IntValue:
		value = fmt.Sprintf("%d", v.IntValue)
	default:
		value = "unknown type"
	}

	fmt.Printf("Get Response: Value=%v, Metadata=%v, Latency=%v\n", value, resp.Metadata, latency)
}

func doInteractiveBulkGet(client proto.StorageServiceClient, reader *bufio.Reader) {
	fmt.Print("Enter keys (comma-separated): ")
	keysInput, _ := reader.ReadString('\n')
	keys := strings.Split(strings.TrimSpace(keysInput), ",")

	var requests []*proto.GetRequest
	for _, key := range keys {
		requests = append(requests, &proto.GetRequest{Key: strings.TrimSpace(key)})
	}

	req := &proto.BulkGetRequest{
		Requests: requests,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.BulkGet(ctx, req)
	latency := time.Since(start)

	if err != nil {
		log.Fatalf("BulkGet request failed: %v", err)
	}

	fmt.Println("BulkGet Response:")
	for _, r := range resp.Responses {
		var value string
		switch v := r.Value.(type) {
		case *proto.GetResponse_StringValue:
			value = v.StringValue
		case *proto.GetResponse_IntValue:
			value = fmt.Sprintf("%d", v.IntValue)
		default:
			value = "unknown type"
		}
		fmt.Printf("- Value=%v, Metadata=%v\n", value, r.Metadata)
	}
	fmt.Printf("Latency: %v\n", latency)
}

func displaySchema() {
	fmt.Println("\nProto Schema:")
	fmt.Println("SetRequest:")
	fmt.Println("  - Key: string")
	fmt.Println("  - Value: (string/int)")
	fmt.Println("  - Metadata: map[string]string")
	fmt.Println("GetRequest:")
	fmt.Println("  - Key: string")
	fmt.Println("BulkGetRequest:")
	fmt.Println("  - Requests: []GetRequest")
}

func stressTest(client proto.StorageServiceClient, reader *bufio.Reader) {
	fmt.Print("Enter number of requests: ")
	numRequestsInput, _ := reader.ReadString('\n')
	numRequests, err := strconv.Atoi(strings.TrimSpace(numRequestsInput))
	if err != nil || numRequests <= 0 {
		fmt.Println("Invalid number. Please try again.")
		return
	}

	fmt.Println("Performing stress test...")
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		req := &proto.SetRequest{
			Key: fmt.Sprintf("key-%d", i),
			Value: &proto.SetRequest_StringValue{
				StringValue: fmt.Sprintf("value-%d", i),
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		_, err := client.Set(ctx, req)
		cancel()

		if err != nil {
			log.Printf("Set request %d failed: %v", i, err)
		}
	}

	latency := time.Since(start)
	fmt.Printf("Stress test completed: %d requests, Total Latency=%v\n", numRequests, latency)
}
