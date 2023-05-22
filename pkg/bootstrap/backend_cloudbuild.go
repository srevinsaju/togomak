package bootstrap

import (
	"archive/tar"
	"bufio"
	"bytes"
	"cloud.google.com/go/storage"
	"compress/gzip"
	goctx "context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/flosch/pongo2/v6"
	"github.com/google/uuid"
	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/templating"
	"github.com/srevinsaju/togomak/pkg/x"
	"google.golang.org/api/cloudbuild/v1"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func createTarball(directory string) ([]byte, error) {
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	// Read the .gitignore file, if present

	gcloudIgnoreRules, err := gitignore.CompileIgnoreFile(filepath.Join(directory, ".gcloudignore"))

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file matches any patterns in .gitignore
		if gcloudIgnoreRules != nil && gcloudIgnoreRules.MatchesPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(tw, file); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func readPartialObject(ctx goctx.Context, client *storage.Client, bucketName, objectName string) (reader *storage.Reader, err error) {
	attrs, err := client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	offset := int64(0)
	if attrs.Size > 1024 {
		offset = attrs.Size - 1024
	}

	return client.Bucket(bucketName).Object(objectName).NewRangeReader(ctx, offset, -1)
}

func LogStreamer(client *storage.Client, logsBucket string, logsObject string, writer io.Writer, ctx goctx.Context, wg *sync.WaitGroup) {
	var reader *storage.Reader
	var err error
	backoff.Retry(func() error {
		reader, err = client.Bucket(logsBucket).Object(logsObject).NewReader(ctx)
		if err != nil {
			if err == storage.ErrObjectNotExist {
				writer.Write([]byte(fmt.Sprintf("logs not available yet, retrying...\n")))
				return err
			}
			writer.Write([]byte(fmt.Sprintf("failed to read logs: %v\n", err)))
			return err
		}
		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 12))
	defer reader.Close()
	writer.Write([]byte(fmt.Sprintf("Reading logs from gs://%s/%s\n", logsBucket, logsObject)))

	for {
		reader, err := readPartialObject(ctx, client, logsBucket, logsObject)
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Failed to read logs: %v\n", err)))
		}

		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			line := scanner.Text()
			n, err := writer.Write([]byte(line))
			writer.Write([]byte("\n"))
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("Failed to read logs: %v\n", err)))
			}
			if n != len(line) {
				writer.Write([]byte(fmt.Sprintf("Failed to read logs: %v\n", io.ErrShortWrite)))
			}

			if strings.HasPrefix(line, "DONE") || strings.HasPrefix(line, "ERROR") {
				wg.Done()
				return
			}
		}

		if scanner.Err() != nil {
			writer.Write([]byte(fmt.Sprintf("Failed to read logs: %v\n", scanner.Err())))
		}

		time.Sleep(5 * time.Second) // Adjust the sleep duration as per your requirements
	}

}

func CloudBuild(rootCtx *context.Context, data schema.SchemaConfig) error {
	togomakCtx := rootCtx.AddChild("backend", "cloudbuild")
	togomakCtx.Logger.Warnf("togomak + cloudbuild support is experimental, some features may not work as expected.")
	projectID := data.Backend.CloudBuild["project_id"].(string)
	secretEnv, secretEnvExists := data.Backend.CloudBuild["secret_env"].([]string)
	availableSecrets, availableSecretsExists := data.Backend.CloudBuild["availableSecrets"].(map[string]interface{})

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

	bucket, ok := data.Backend.CloudBuild["bucket"].(string)
	if !ok {
		bucket = fmt.Sprintf("%s_cloudbuild", projectID)
	}
	source, ok := data.Backend.CloudBuild["source-prefix"].(string)
	if !ok {
		// timestamp in precise nanoseconds
		source = fmt.Sprintf("togomak-%d-%s.tgz", time.Now().UnixNano(), uuid.New().String())
	}
	togomakCtx.Logger.Infof("Uploading current directory [.] as [gs://%s/%s]'\n", bucket, source)
	// Upload the current directory as a tarball
	wc := client.Bucket(bucket).Object(source).NewWriter(ctx)
	wc.ContentType = "application/x-tar"
	wc.Metadata = map[string]string{
		"x-goog-meta-togomak": "true",
	}
	// create a tarball of the current directory
	tarball, err := createTarball(".")
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to create tarball: %v", err)
	}
	n, err := wc.Write(tarball)
	if err != nil {
		togomakCtx.Logger.Fatalf("Failed to upload tarball to google cloud storage: %v", err)
	}
	togomakCtx.Logger.Infof("Uploaded %d bytes to gs://%s/%s\n", n, bucket, source)
	x.Must(wc.Close())
	var environ []string
	for k, v := range data.Environment {
		tpl, err := pongo2.FromString(v)
		if err != nil {
			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}
		parsedV, err := templating.Execute(tpl, rootCtx.Data.AsMap())
		parsedV = strings.TrimSpace(parsedV)

		if err != nil {
			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}

		environ = append(environ, fmt.Sprintf("%s=%s", k, parsedV))
	}

	if !secretEnvExists {
		secretEnv = []string{}
	}

	build := &cloudbuild.Build{
		Steps: []*cloudbuild.BuildStep{
			{
				Name:      "ghcr.io/srevinsaju/togomak",
				Env:       environ,
				SecretEnv: secretEnv,
			},
		},
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Object: source,
				Bucket: bucket,
			},
		},
		AvailableSecrets: &cloudbuild.Secrets{
			SecretManager: []*cloudbuild.SecretManagerSecret{},
		},
	}
	if availableSecretsExists {
		togomakCtx.Logger.Info("found available secrets")
		secretManager, secretManagerExists := availableSecrets["secretManager"].([]map[string]string)

		if secretManagerExists {
			togomakCtx.Logger.Info("found secrets on secret manager")
			for i := range secretManager {
				build.AvailableSecrets.SecretManager = append(build.AvailableSecrets.SecretManager,
					&cloudbuild.SecretManagerSecret{
						VersionName: secretManager[i]["versionName"],
						Env:         secretManager[i]["env"],
					})
			}
		} else {
			togomakCtx.Logger.Info("no secrets found on secret manager")
		}

	} else {
		togomakCtx.Logger.Info("no available secrets found")
	}

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

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go LogStreamer(client, logsBucket, logsObject, togomakCtx.Logger.Writer(), ctx, wg)

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
				togomakCtx.Logger.Warnf("build error: %v (%v)", resp.Error.Message, resp.Error.Details)
				togomakCtx.Logger.Warnf("waiting for logs to be complete...")
				wg.Wait()
				metadata := map[string]interface{}{}
				metadataJson, err := resp.MarshalJSON()
				if err != nil {
					panic(err)
				}
				m, err := json.MarshalIndent(metadata, "", "  ")
				if err != nil {
					panic(err)
				}
				err = json.Unmarshal(metadataJson, &metadata)
				fmt.Printf("%s\n", string(m[:]))
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
	wg.Wait()
	return nil
	//
	//// Download the object from bucket.
	//togomakCtx.Logger.Infof("Downloading logs from %s/%s\n", logsBucket, logsObject)
	//
	//rc, err := client.Bucket(logsBucket).Object(logsObject).NewReader(ctx)
	//if err != nil {
	//	togomakCtx.Logger.Fatalf("Failed to create object reader: %v", err)
	//
	//}
	//defer rc.Close()
	//
	//// Write the object's content to stdout, rc is reader
	//// that implements io.Reader interface.
	//togomakCtx.Logger.Infof("Streaming logs from %s/%s\n", logsBucket, logsObject)
	//
	//// Copy the file's contents to standard output (terminal)
	//_, err = io.Copy(togomakCtx.Logger.Writer(), rc)
	//if err != nil {
	//	togomakCtx.Logger.Fatalf("Failed to read object: %v", err)
	//}

}
