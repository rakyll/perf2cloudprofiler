// Program perf2cloudprofiler converts perf
// output to pprof and uploads it to Google Cloud Profiler.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/google/pprof/profile"

	"cloud.google.com/go/compute/metadata"

	"google.golang.org/api/option"
	gtransport "google.golang.org/api/transport/grpc"
	pb "google.golang.org/genproto/googleapis/devtools/cloudprofiler/v2"
)

var (
	client pb.ProfilerServiceClient

	project string
	zone    string
	input   string
	target  string
)

const (
	apiAddr = "cloudprofiler.googleapis.com:443"
	scope   = "https://www.googleapis.com/auth/monitoring.write"
)

func main() {
	ctx := context.Background()
	flag.StringVar(&project, "project", "", "")
	flag.StringVar(&zone, "zone", "", "")
	flag.StringVar(&input, "i", "perf.data", "")
	flag.StringVar(&target, "target", "", "")
	flag.Parse()

	if project == "" {
		id, err := metadata.ProjectID()
		if err != nil {
			log.Fatalf("Cannot resolve the GCP project from the metadata server: %v", err)
		}
		project = id
	}
	if zone == "" {
		z, err := metadata.Zone()
		if err != nil {
			log.Fatalf("Cannot resolve the GCP zone from the metadata server: %v", err)
		}
		zone = z
	}

	if target == "" {
		target = input
	}

	opts := []option.ClientOption{
		option.WithEndpoint(apiAddr),
		option.WithScopes(scope),
	}
	conn, err := gtransport.Dial(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}
	client = pb.NewProfilerServiceClient(conn)

	pprofBytes, err := convert(input)
	if err != nil {
		log.Fatalf("Cannot convert perf data to pprof: %v", err)
	}

	if err := upload(ctx, pprofBytes); err != nil {
		log.Fatalf("Cannot upload to Google Cloud Profiler: %v", err)
	}
	fmt.Printf("https://console.cloud.google.com/profiler/%s;type=%s?project=%s\n", url.PathEscape(target), pb.ProfileType_CPU, project)
}

func convert(perfFile string) (pprofBytes []byte, err error) {
	tmpFile, err := ioutil.TempFile("", "perf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	cmd := exec.Command("perf_to_profile",
		"-i", perfFile,
		"-o", tmpFile.Name(),
		"-f")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(tmpFile)
}

func upload(ctx context.Context, payload []byte) error {
	// Reset time.
	resetted, err := resetTime(payload)
	if err != nil {
		log.Printf("Cannot reset the profile's time: %v", err)
	}

	req := pb.CreateOfflineProfileRequest{
		Parent: "projects/" + project,
		Profile: &pb.Profile{
			ProfileType: pb.ProfileType_CPU,
			Deployment: &pb.Deployment{
				ProjectId: project,
				Target:    target,
				Labels: map[string]string{
					"zone": zone,
				},
			},
			ProfileBytes: resetted,
		},
	}

	// TODO(jbd): Is there a way without having
	// to load the profile all in memory?
	_, err = client.CreateOfflineProfile(ctx, &req)
	return err
}

func resetTime(pprofBytes []byte) ([]byte, error) {
	p, err := profile.ParseData(pprofBytes)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse the profile: %v", err)
	}
	p.TimeNanos = time.Now().UnixNano()

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
