/*
 * MinIO Cloud Storage, (C) 2018 MinIO, Inc.
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package target

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/obstor/obstor/pkg/event"
	xnet "github.com/obstor/obstor/pkg/net"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Elastic constants
const (
	ElasticFormat     = "format"
	ElasticURL        = "url"
	ElasticIndex      = "index"
	ElasticQueueDir   = "queue_dir"
	ElasticQueueLimit = "queue_limit"
	ElasticUsername   = "username"
	ElasticPassword   = "password"

	EnvElasticEnable     = "OBSTOR_NOTIFY_ELASTICSEARCH_ENABLE"
	EnvElasticFormat     = "OBSTOR_NOTIFY_ELASTICSEARCH_FORMAT"
	EnvElasticURL        = "OBSTOR_NOTIFY_ELASTICSEARCH_URL"
	EnvElasticIndex      = "OBSTOR_NOTIFY_ELASTICSEARCH_INDEX"
	EnvElasticQueueDir   = "OBSTOR_NOTIFY_ELASTICSEARCH_QUEUE_DIR"
	EnvElasticQueueLimit = "OBSTOR_NOTIFY_ELASTICSEARCH_QUEUE_LIMIT"
	EnvElasticUsername   = "OBSTOR_NOTIFY_ELASTICSEARCH_USERNAME"
	EnvElasticPassword   = "OBSTOR_NOTIFY_ELASTICSEARCH_PASSWORD"
)

// ElasticsearchArgs - Elasticsearch target arguments.
type ElasticsearchArgs struct {
	Enable     bool            `json:"enable"`
	Format     string          `json:"format"`
	URL        xnet.URL        `json:"url"`
	Index      string          `json:"index"`
	QueueDir   string          `json:"queueDir"`
	QueueLimit uint64          `json:"queueLimit"`
	Transport  *http.Transport `json:"-"`
	Username   string          `json:"username"`
	Password   string          `json:"password"`
}

// Validate ElasticsearchArgs fields
func (a ElasticsearchArgs) Validate() error {
	if !a.Enable {
		return nil
	}
	if a.URL.IsEmpty() {
		return errors.New("empty URL")
	}
	if a.Format != "" {
		f := strings.ToLower(a.Format)
		if f != event.NamespaceFormat && f != event.AccessFormat {
			return errors.New("format value unrecognized")
		}
	}
	if a.Index == "" {
		return errors.New("empty index value")
	}

	if (a.Username == "" && a.Password != "") || (a.Username != "" && a.Password == "") {
		return errors.New("username and password should be set in pairs")
	}

	return nil
}

// ElasticsearchTarget - Elasticsearch target.
type ElasticsearchTarget struct {
	id         event.TargetID
	args       ElasticsearchArgs
	client     *elasticsearch.Client
	store      Store
	loggerOnce func(ctx context.Context, err error, id interface{}, errKind ...interface{})
}

// ID - returns target ID.
func (target *ElasticsearchTarget) ID() event.TargetID {
	return target.id
}

// HasQueueStore - Checks if the queueStore has been configured for the target
func (target *ElasticsearchTarget) HasQueueStore() bool {
	return target.store != nil
}

// IsActive - Return true if target is up and active
func (target *ElasticsearchTarget) IsActive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if target.client == nil {
		client, err := newClient(target.args)
		if err != nil {
			return false, err
		}
		target.client = client
	}
	res, err := target.client.Ping(target.client.Ping.WithContext(ctx))
	if err != nil {
		if isConnErr(err) || isContextErr(err) || xnet.IsNetworkOrHostDown(err, false) {
			return false, errNotConnected
		}
		return false, err
	}
	defer func() { _ = res.Body.Close() }()
	return res.StatusCode < http.StatusBadRequest, nil
}

// isConnErr returns true if the error is a network connection error.
func isConnErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// isContextErr returns true if the error is a context cancellation or deadline exceeded error.
func isContextErr(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Save - saves the events to the store if queuestore is configured, which will be replayed when the elasticsearch connection is active.
func (target *ElasticsearchTarget) Save(eventData event.Event) error {
	if target.store != nil {
		return target.store.Put(eventData)
	}
	err := target.send(eventData)
	if isConnErr(err) || isContextErr(err) || xnet.IsNetworkOrHostDown(err, false) {
		return errNotConnected
	}
	return err
}

// send - sends the event to the target.
func (target *ElasticsearchTarget) send(eventData event.Event) error {

	var key string

	exists := func() (bool, error) {
		req := esapi.ExistsRequest{
			Index:      target.args.Index,
			DocumentID: key,
		}
		res, err := req.Do(context.Background(), target.client)
		if err != nil {
			return false, err
		}
		defer func() { _ = res.Body.Close() }()
		return !res.IsError(), nil
	}

	remove := func() error {
		docExists, err := exists()
		if err == nil && docExists {
			req := esapi.DeleteRequest{
				Index:      target.args.Index,
				DocumentID: key,
			}
			res, err := req.Do(context.Background(), target.client)
			if err != nil {
				return err
			}
			defer func() { _ = res.Body.Close() }()
			if res.IsError() {
				return fmt.Errorf("delete failed: %s", res.String())
			}
		}
		return err
	}

	update := func() error {
		data, err := json.Marshal(map[string]interface{}{"Records": []event.Event{eventData}})
		if err != nil {
			return err
		}
		req := esapi.IndexRequest{
			Index:      target.args.Index,
			DocumentID: key,
			Body:       bytes.NewReader(data),
		}
		res, err := req.Do(context.Background(), target.client)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()
		if res.IsError() {
			return fmt.Errorf("index failed: %s", res.String())
		}
		return nil
	}

	add := func() error {
		data, err := json.Marshal(map[string]interface{}{"Records": []event.Event{eventData}})
		if err != nil {
			return err
		}
		req := esapi.IndexRequest{
			Index: target.args.Index,
			Body:  bytes.NewReader(data),
		}
		res, err := req.Do(context.Background(), target.client)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()
		if res.IsError() {
			return fmt.Errorf("index failed: %s", res.String())
		}
		return nil
	}

	if target.args.Format == event.NamespaceFormat {
		objectName, err := url.QueryUnescape(eventData.S3.Object.Key)
		if err != nil {
			return err
		}

		key = eventData.S3.Bucket.Name + "/" + objectName
		if eventData.EventName == event.ObjectRemovedDelete {
			err = remove()
		} else {
			err = update()
		}
		return err
	}

	if target.args.Format == event.AccessFormat {
		return add()
	}

	return nil
}

// Send - reads an event from store and sends it to Elasticsearch.
func (target *ElasticsearchTarget) Send(eventKey string) error {
	var err error
	if target.client == nil {
		target.client, err = newClient(target.args)
		if err != nil {
			return err
		}
	}

	eventData, eErr := target.store.Get(eventKey)
	if eErr != nil {
		// The last event key in a successful batch will be sent in the channel atmost once by the replayEvents()
		// Such events will not exist and wouldve been already been sent successfully.
		if os.IsNotExist(eErr) {
			return nil
		}
		return eErr
	}

	if err := target.send(eventData); err != nil {
		if isConnErr(err) || isContextErr(err) || xnet.IsNetworkOrHostDown(err, false) {
			return errNotConnected
		}
		return err
	}

	// Delete the event from store.
	return target.store.Del(eventKey)
}

// Close - does nothing and available for interface compatibility.
func (target *ElasticsearchTarget) Close() error {
	return nil
}

// createIndex - creates the index if it does not exist.
func createIndex(client *elasticsearch.Client, args ElasticsearchArgs) error {
	res, err := client.Indices.Exists([]string{args.Index})
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode == http.StatusNotFound {
		res, err = client.Indices.Create(args.Index)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()
		if res.IsError() {
			return fmt.Errorf("index %v not created: %s", args.Index, res.String())
		}
	}
	return nil
}

// newClient - creates a new elasticsearch client with args provided.
func newClient(args ElasticsearchArgs) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses:  []string{args.URL.String()},
		Transport:  args.Transport,
		MaxRetries: 10,
	}
	if args.Username != "" && args.Password != "" {
		cfg.Username = args.Username
		cfg.Password = args.Password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		if isConnErr(err) || isContextErr(err) || xnet.IsNetworkOrHostDown(err, false) {
			return nil, errNotConnected
		}
		return nil, err
	}

	// Ping to verify connection
	res, err := client.Ping()
	if err != nil {
		if isConnErr(err) || isContextErr(err) || xnet.IsNetworkOrHostDown(err, false) {
			return nil, errNotConnected
		}
		return nil, err
	}
	_ = res.Body.Close()

	if err = createIndex(client, args); err != nil {
		return nil, err
	}
	return client, nil
}

// NewElasticsearchTarget - creates new Elasticsearch target.
func NewElasticsearchTarget(id string, args ElasticsearchArgs, doneCh <-chan struct{}, loggerOnce func(ctx context.Context, err error, id interface{}, kind ...interface{}), test bool) (*ElasticsearchTarget, error) {
	target := &ElasticsearchTarget{
		id:         event.TargetID{ID: id, Name: "elasticsearch"},
		args:       args,
		loggerOnce: loggerOnce,
	}

	if args.QueueDir != "" {
		queueDir := filepath.Join(args.QueueDir, storePrefix+"-elasticsearch-"+id)
		target.store = NewQueueStore(queueDir, args.QueueLimit)
		if err := target.store.Open(); err != nil {
			target.loggerOnce(context.Background(), err, target.ID())
			return target, err
		}
	}

	var err error
	target.client, err = newClient(args)
	if err != nil {
		if target.store == nil || err != errNotConnected {
			target.loggerOnce(context.Background(), err, target.ID())
			return target, err
		}
	}

	if target.store != nil && !test {
		// Replays the events from the store.
		eventKeyCh := replayEvents(target.store, doneCh, target.loggerOnce, target.ID())
		// Start replaying events from the store.
		go sendEvents(target, eventKeyCh, doneCh, target.loggerOnce)
	}

	return target, nil
}
