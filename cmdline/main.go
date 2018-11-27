package main

import (
	"context"
	"fmt"
	"github.com/fiibbb/gitdb/proto"
	"google.golang.org/grpc"
	"time"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := gitdbpb.NewGitClient(conn)
	resp, err := client.GetObject(context.Background(), &gitdbpb.GetObjectRequest{
		Id: &gitdbpb.ObjectIdentifier{
			Type: gitdbpb.ObjectType_BLOB,
			Path: "foo",
			Time: time.Now().Unix(),
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("`%s`", resp)
}
