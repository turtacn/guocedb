package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// This import will not work until the .proto file is compiled with protoc.
	// mgmtv1 "github.com/turtacn/guocedb/api/protobuf/mgmt/v1"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The gRPC server address")
)

func main() {
	flag.Parse()

	// For now, we will just have a simple "status" command.
	// A real CLI would use a library like Cobra for subcommands.
	cmd := "status"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	switch cmd {
	case "status":
		getStatus()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Usage: guocedb-cli [command]")
		fmt.Println("Available commands: status")
	}
}

func getStatus() {
	fmt.Printf("Connecting to server at %s...\n", *serverAddr)

	// Set up a connection to the server.
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// This part is commented out because the gRPC generated code doesn't exist yet.
	/*
		c := mgmtv1.NewManagementServiceClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := c.GetStatus(ctx, &mgmtv1.GetStatusRequest{})
		if err != nil {
			log.Fatalf("could not get status: %v", err)
		}

		fmt.Println("Server Status:")
		fmt.Printf("  Version: %s\n", r.GetVersion())
		fmt.Printf("  Status: %s\n", r.GetStatus())
		fmt.Printf("  Active Connections: %d\n", r.GetActiveConnections())
	*/

	// Placeholder message since we can't actually call the RPC yet.
	fmt.Println("\nNote: gRPC client code is commented out as it requires generated protobuf files.")
	fmt.Println("This demonstrates the basic connection logic.")
}
