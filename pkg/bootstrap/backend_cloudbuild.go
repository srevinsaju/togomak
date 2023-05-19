package bootstrap

import (
	"cloud.google.com/go/storage"
	goctx "context"
	"encoding/json"
	"fmt"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"google.golang.org/api/cloudbuild/v1"
	"io"
	"strings"
	"time"
)

func CloudBuild(rootCtx *context.Context, data schema.SchemaConfig) {
	togomakCtx := rootCtx.AddChild("backend", "cloudbuild")
	togomakCtx.Logger.Warnf("togomak + cloudbuild support is experimental, some features may not work as expected.")

	ctx := goctx.Background()
	service, err := cloudbuild.NewService(ctx)

	// Create a Google Cloud Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to create Google Cloud Storage client: %v", err)
	}
	defer client.Close()

	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to create Cloud Build service client: %v", err)
	}

	build := &cloudbuild.Build{
		Steps: []*cloudbuild.BuildStep{},
	}
	for i, stage := range data.Stages {
		togomakCtx.Logger.Infof("togomak:cloudbuild:stage %d: %s\n", i, stage.Id)
		container := stage.Container
		if stage.Container == "" {
			// TODO: make this configurable
			container = "ubuntu:latest"
		}
		var environment []string
		for k, v := range stage.Environment {
			environment = append(environment, k+"="+v)
		}
		build.Steps = append(build.Steps, &cloudbuild.BuildStep{
			Name:         container,
			Args:         stage.Args,
			Id:           stage.Id,
			Script:       stage.Script,
			Dir:          stage.Dir,
			Env:          environment,
			AllowFailure: data.Options.FailLazy,
			WaitFor:      stage.DependsOn,
		})

	}

	projectID := data.Backend.CloudBuild["project_id"].(string)
	resp, err := service.Projects.Builds.Create(projectID, build).Do()
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to create build: %v", err)
	}

	togomakCtx.Logger.Infof("togomak:cloudbuild created with ID: %s\n", resp.Name)
	metadataJson, err := resp.Metadata.MarshalJSON()
	if err != nil {
		panic(err)
	}
	metadata := map[string]interface{}{}
	err = json.Unmarshal(metadataJson, &metadata)
	if err != nil {
		panic(err)
	}
	togomakCtx.Logger.Infof("togomak:cloudbuild see live logs at %s\n", metadata["build"].(map[string]interface{})["logUrl"].(string))

	// parse the log bucket information
	logsBucket := metadata["build"].(map[string]interface{})["logsBucket"].(string)
	id := metadata["build"].(map[string]interface{})["id"].(string)
	logsBucket = strings.TrimPrefix(logsBucket, "gs://")
	logsBucket = strings.TrimSuffix(logsBucket, "/")
	logsObject := fmt.Sprintf("log-%s.txt", id)

	// Poll the operation until it is done.
	for {
		time.Sleep(2 * time.Second)
		resp, err := service.Operations.Get(resp.Name).Do()
		if err != nil {
			togomakCtx.Logger.Fatalf("failed to get build status: %v", err)
			break
		}
		if resp.Done {
			time.Sleep(1 * time.Second)
			if resp.Error != nil {
				togomakCtx.Logger.Fatalf("build error: %v", resp.Error)
			}
			metadataJson, err := resp.Metadata.MarshalJSON()
			if err != nil {
				panic(err)
			}
			metadata := map[string]interface{}{}
			err = json.Unmarshal(metadataJson, &metadata)
			togomakCtx.Logger.Infof("build status: %s\n", metadata["build"].(map[string]interface{})["status"].(string))

			m, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", string(m[:]))
			break

		}
	}

	// Download the object from bucket.
	togomakCtx.Logger.Infof("Downloading logs from %s/%s\n", logsBucket, logsObject)

	rc, err := client.Bucket(logsBucket).Object(logsObject).NewReader(ctx)
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to create object reader: %v", err)

	}
	defer rc.Close()

	// Write the object's content to stdout, rc is reader
	// that implements io.Reader interface.
	togomakCtx.Logger.Infof("Streaming logs from %s/%s\n", logsBucket, logsObject)

	// Copy the file's contents to standard output (terminal)
	_, err = io.Copy(togomakCtx.Logger.Writer(), rc)
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to read object: %v", err)
	}

}
