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

	"cloud.google.com/go/compute/metadata"
	"github.com/google/pprof/profile"
	"google.golang.org/api/option"
	gtransport "google.golang.org/api/transport/grpc"
	pb "google.golang.org/genproto/googleapis/devtools/cloudprofiler/v2"
)

var (
	client pb.ProfilerServiceClient

	project  string
	zone     string
	instance string
	input    string
	target   string
)

const (
	apiAddr       = "cloudprofiler.googleapis.com:443"
	scope         = "https://www.googleapis.com/auth/monitoring.write"
	perfToProfile = "perf_to_profile"
)

func main() {
	ctx := context.Background()
	flag.StringVar(&project, "project", "", "")
	flag.StringVar(&zone, "zone", "", "")
	flag.StringVar(&instance, "instance", "", "")
	flag.StringVar(&input, "i", "perf.data", "")
	flag.StringVar(&target, "target", "", "")
	flag.Usage = usageAndExit
	flag.Parse()

	// TODO(jbd): Automatically detect input. Don't convert if pprof.

	if project == "" {
		id, err := metadata.ProjectID()
		if err != nil {
			log.Fatalf("Cannot resolve the GCP project from the metadata server: %v", err)
		}
		project = id
	}
	if zone == "" {
		// Ignore error. If we cannot resolve the instance name,
		// it would be too aggressive to fatal exit.
		zone, _ = metadata.Zone()
	}

	if instance == "" {
		// Ignore error. If we cannot resolve the instance name,
		// it would be too aggressive to fatal exit.
		instance, _ = metadata.InstanceName()
	}

	if target == "" {
		target = input
	}

	conn, err := gtransport.Dial(ctx,
		option.WithEndpoint(apiAddr),
		option.WithScopes(scope))
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

func convert(file string) (pprofBytes []byte, err error) {
	tmpFile, err := ioutil.TempFile("", "perf")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	cmd := exec.Command(perfToProfile,
		"-i", file,
		"-o", tmpFile.Name(),
		"-f")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(tmpFile)
}

func upload(ctx context.Context, payload []byte) error {
	// Reset time, otherwise old profiles wont be shown
	// at Cloud profiler due to data retention limits.
	resetted, err := resetTime(payload)
	if err != nil {
		log.Printf("Cannot reset the profile's time: %v", err)
	}

	req := &pb.CreateOfflineProfileRequest{
		Parent: "projects/" + project,
		Profile: &pb.Profile{
			// TODO(jbd): Guess the profile type from the input.
			ProfileType: pb.ProfileType_CPU,
			Deployment: &pb.Deployment{
				ProjectId: project,
				Target:    target,
				Labels: map[string]string{
					"zone":     zone,
					"instance": instance,
				},
			},
			ProfileBytes: resetted,
		},
	}

	// TODO(jbd): Is there a way without having
	// to load the profile all in memory?
	_, err = client.CreateOfflineProfile(ctx, req)
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

// TODO(jbd): Check binary dependencies and install if not available.

const usageText = `perf2cloudprofiler [-i perf.data]

Other options:
-target   Target Cloud Profiler profile name to upload data to.
-project  Google Cloud project name, tries to automatically
          resolve if none is set.
-zone     Google Cloud zone, tries to automatically resolve if
          none is set.
-instance Google Compute Engine instance name, tries to resolve
          automatically if none is set.`

func usageAndExit() {
	fmt.Println(usageText)
	os.Exit(1)
}
