package media

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/egfanboy/mediapire-manager/internal/node"
	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type downloaderQueue struct {
	// mapping between nodeId and the mediaIds from that node to download
	queue map[node.NodeConfig][]uuid.UUID

	content map[node.NodeConfig][]byte
}

func (q *downloaderQueue) processNode(ctx context.Context, n node.NodeConfig) error {
	ids, ok := q.queue[n]

	if !ok {
		return errors.New("Nope")
	}

	client := mhApi.NewClient(ctx)

	content, _, err := client.DownloadMedia(n, ids)
	if err != nil {
		return err
	}

	q.content[n] = content

	return nil
}

func (q *downloaderQueue) ProcessQueue(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	wg := sync.WaitGroup{}

	// get media from all nodes at the same time
	for nodeInQueue := range q.queue {
		wg.Add(1)

		go func(n node.NodeConfig) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return // Error somewhere, terminate
			default:
			}

			err := q.processNode(ctx, n)
			if err != nil {
				log.Err(err).Msgf("An error occured when downloading content from node %s", n.NodeHost)
				cancel()
			}
		}(nodeInQueue)

	}

	wg.Wait()

	if ctx.Err() != nil && ctx.Err() == context.Canceled {
		return errors.New("failed to download selected content from nodes")
	}

	return nil
}

func (q *downloaderQueue) GetContent(ctx context.Context) ([]byte, error) {

	numberOfNodes := len(q.queue)

	result := new(bytes.Buffer)

	zipWriter := zip.NewWriter(result)

	for _, nodeContent := range q.content {
		// this is the only node, just return the content
		if numberOfNodes == 1 {
			return nodeContent, nil
		}

		br := bytes.NewReader(nodeContent)

		zipReader, err := zip.NewReader(br, br.Size())
		if err != nil {
			return nil, err
		}

		for _, file := range zipReader.File {
			zipItem, err := file.Open()
			if err != nil {
				return nil, err
			}

			w, err := zipWriter.Create(file.Name)
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(w, zipItem)
			if err != nil {
				return nil, err
			}

		}

	}

	err := zipWriter.Close()

	return result.Bytes(), err
}

type mediaDownloader struct{}

type populatedDownloadItem struct {
	Node    node.NodeConfig
	MediaId uuid.UUID
}

func (mediaDownloader) Download(ctx context.Context, items []populatedDownloadItem) ([]byte, error) {
	nodeMappings := make(map[node.NodeConfig][]uuid.UUID)

	for _, item := range items {
		if mapping, ok := nodeMappings[item.Node]; ok {
			nodeMappings[item.Node] = append(mapping, item.MediaId)
		} else {
			nodeMappings[item.Node] = []uuid.UUID{item.MediaId}
		}
	}

	queue := &downloaderQueue{queue: nodeMappings, content: map[node.NodeConfig][]byte{}}

	err := queue.ProcessQueue(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to process download queue")

		return nil, err
	}

	content, err := queue.GetContent(ctx)
	if err != nil {
		log.Err(err).Msg("Failed to merge downloaded content")

		return nil, err
	}

	return content, nil
}
