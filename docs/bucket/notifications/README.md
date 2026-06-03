# Obstor Bucket Notification Guide

Events occurring on objects in a bucket can be monitored using bucket event notifications.

> NOTE: Backend mode does not support bucket notifications (except NAS backend).

Various event types supported by Obstor server are

| Supported Object Event Types     |                                            |                                        |
| :----------------------          | ------------------------------------------ | -------------------------------------  |
| `s3:ObjectCreated:Put`           | `s3:ObjectCreated:CompleteMultipartUpload` | `s3:ObjectAccessed:Head`               |
| `s3:ObjectCreated:Post`          | `s3:ObjectRemoved:Delete`                  | `s3:ObjectRemoved:DeleteMarkerCreated` |
| `s3:ObjectCreated:Copy`          | `s3:ObjectAccessed:Get`                    |                                        |
| `s3:ObjectCreated:PutRetention`  | `s3:ObjectCreated:PutLegalHold`            |                                        |
| `s3:ObjectAccessed:GetRetention` | `s3:ObjectAccessed:GetLegalHold`           |                                        |

| Supported Replication Event Types                  |
| :------------                                      |
| `s3:Replication:OperationFailedReplication`        |
| `s3:Replication:OperationCompletedReplication`     |
| `s3:Replication:OperationNotTracked`               |
| `s3:Replication:OperationMissedThreshold`          |
| `s3:Replication:OperationReplicatedAfterThreshold` |

| Supported ILM Transition Event Types |
| :-----                               |
| `s3:ObjectRestore:Post`              |
| `s3:ObjectRestore:Completed`         |

| Supported Global Event Types (Only supported through ListenNotification API) |
| :-----                                                                       |
| `s3:BucketCreated`                                                           |
| `s3:BucketRemoved`                                                           |


Use client tools like `mc` to set and listen for event notifications using the [`event` sub-command](/docs/bucket/notifications). Obstor SDK's [`BucketNotification` APIs](/docs/bucket/notifications) can also be used. The notification message Obstor sends to publish an event is a JSON message with the following [structure](https://docs.aws.amazon.com/AmazonS3/latest/dev/notification-content-structure.html).

Bucket events can be published to the following targets:

| Supported Notification Targets    |                             |                                 |
| :-------------------------------- | --------------------------- | ------------------------------- |
| [`AMQP`](#AMQP)                   | [`Redis`](#Redis)           | [`MySQL`](#MySQL)               |
| [`MQTT`](#MQTT)                   | [`NATS`](#NATS)             | [`Apache Kafka`](#apache-kafka) |
| [`Elasticsearch`](#Elasticsearch) | [`PostgreSQL`](#PostgreSQL) | [`Webhooks`](#webhooks)         |
| [`NSQ`](#NSQ)                     |                             |                                 |

## Prerequisites

- Install and configure Obstor Server from here.
- Install and configure Obstor Client from here.

```bash
$ mc admin config get myobstor | grep notify
notify_webhook        publish bucket notifications to webhook endpoints
notify_amqp           publish bucket notifications to AMQP endpoints
notify_kafka          publish bucket notifications to Kafka endpoints
notify_mqtt           publish bucket notifications to MQTT endpoints
notify_nats           publish bucket notifications to NATS endpoints
notify_nsq            publish bucket notifications to NSQ endpoints
notify_mysql          publish bucket notifications to MySQL databases
notify_postgres       publish bucket notifications to Postgres databases
notify_elasticsearch  publish bucket notifications to Elasticsearch endpoints
notify_redis          publish bucket notifications to Redis datastores
```

> NOTE:
> - '\*' at the end of arg means its mandatory.
> - '\*' at the end of the values, means its the default value for the arg.
> - When configured using environment variables, the `:name` can be specified using this format `OBSTOR_NOTIFY_WEBHOOK_ENABLE_<name>`.

<a name="AMQP"></a>

## Publish Obstor events via AMQP

Install RabbitMQ from [here](https://www.rabbitmq.com/).

### Step 1: Add AMQP endpoint to Obstor

The AMQP configuration is located under the sub-system `notify_amqp` top-level key. Create a configuration key-value pair here for your AMQP instance. The key is a name for your AMQP endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_amqp[:name]  publish bucket notifications to AMQP endpoints

ARGS:
url*           (url)       AMQP server endpoint e.g. `amqp://myuser:mypassword@localhost:5672`
exchange       (string)    name of the AMQP exchange
exchange_type  (string)    AMQP exchange type
routing_key    (string)    routing key for publishing
mandatory      (on|off)    quietly ignore undelivered messages when set to 'off', default is 'on'
durable        (on|off)    persist queue across broker restarts when set to 'on', default is 'off'
no_wait        (on|off)    non-blocking message delivery when set to 'on', default is 'off'
internal       (on|off)    set to 'on' for exchange to be not used directly by publishers, but only when bound to other exchanges
auto_deleted   (on|off)    auto delete queue when set to 'on', when there are no consumers
delivery_mode  (number)    set to '1' for non-persistent or '2' for persistent queue
queue_dir      (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit    (number)    maximum limit for undelivered messages, defaults to '100000'
comment        (sentence)  optionally add a comment to this setting
```

Or environment variables

```
KEY:
notify_amqp[:name]  publish bucket notifications to AMQP endpoints

ARGS:
OBSTOR_NOTIFY_AMQP_ENABLE*        (on|off)    enable notify_amqp target, default is 'off'
OBSTOR_NOTIFY_AMQP_URL*           (url)       AMQP server endpoint e.g. `amqp://myuser:mypassword@localhost:5672`
OBSTOR_NOTIFY_AMQP_EXCHANGE       (string)    name of the AMQP exchange
OBSTOR_NOTIFY_AMQP_EXCHANGE_TYPE  (string)    AMQP exchange type
OBSTOR_NOTIFY_AMQP_ROUTING_KEY    (string)    routing key for publishing
OBSTOR_NOTIFY_AMQP_MANDATORY      (on|off)    quietly ignore undelivered messages when set to 'off', default is 'on'
OBSTOR_NOTIFY_AMQP_DURABLE        (on|off)    persist queue across broker restarts when set to 'on', default is 'off'
OBSTOR_NOTIFY_AMQP_NO_WAIT        (on|off)    non-blocking message delivery when set to 'on', default is 'off'
OBSTOR_NOTIFY_AMQP_INTERNAL       (on|off)    set to 'on' for exchange to be not used directly by publishers, but only when bound to other exchanges
OBSTOR_NOTIFY_AMQP_AUTO_DELETED   (on|off)    auto delete queue when set to 'on', when there are no consumers
OBSTOR_NOTIFY_AMQP_DELIVERY_MODE  (number)    set to '1' for non-persistent or '2' for persistent queue
OBSTOR_NOTIFY_AMQP_QUEUE_DIR      (path)      staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_AMQP_QUEUE_LIMIT    (number)    maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_AMQP_COMMENT        (sentence)  optionally add a comment to this setting
```

Obstor supports persistent event store. The persistent store will backup events when the AMQP broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

To update the configuration, use `mc admin config get notify_amqp` command to get the current configuration for `notify_amqp`.

```bash
$ mc admin config get myobstor/ notify_amqp
notify_amqp:1 delivery_mode="0" exchange_type="" no_wait="off" queue_dir="" queue_limit="0"  url="" auto_deleted="off" durable="off" exchange="" internal="off" mandatory="off" routing_key=""
```

Use `mc admin config set` command to update the configuration for the deployment.Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:amqp` at start-up if there were no errors.

An example configuration for RabbitMQ is shown below:

```bash
$ mc admin config set myobstor/ notify_amqp:1 exchange="bucketevents" exchange_type="fanout" mandatory="false" no_wait="false"  url="amqp://myuser:mypassword@localhost:5672" auto_deleted="false" delivery_mode="0" durable="false" internal="false" routing_key="bucketlogs"
```

Obstor supports all the exchanges available in [RabbitMQ](https://www.rabbitmq.com/). For this setup, we are using `fanout` exchange.

Note that, you can add as many AMQP server endpoint configurations as needed by providing an identifier (like "1" in the example above) for the AMQP instance and an object of per-server configuration parameters.

### Step 2: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded or deleted `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:amqp`. To understand more about ARN please follow [AWS ARN](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) documentation.

```bash
mc mb myobstor/images
mc event add myobstor/images arn:obstor:sqs::1:amqp --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:amqp s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 3: Test on RabbitMQ

The python program below waits on the queue exchange `bucketevents` and prints event notifications on the console. We use [Pika Python Client](https://www.rabbitmq.com/tutorials/tutorial-three-python.html) library to do this.

```python
#!/usr/bin/env python
import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(
        host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='bucketevents',
                         exchange_type='fanout')

result = channel.queue_declare(exclusive=False)
queue_name = result.method.queue

channel.queue_bind(exchange='bucketevents',
                   queue=queue_name)

print(' [*] Waiting for logs. To exit press CTRL+C')

def callback(ch, method, properties, body):
    print(" [x] %r" % body)

channel.basic_consume(callback,
                      queue=queue_name,
                      no_ack=False)

channel.start_consuming()
```

Execute this example python program to watch for RabbitMQ events on the console.

```bash
python rabbit.py
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

You should receive the following event notification via RabbitMQ once the upload completes.

```bash
python rabbit.py
'{"Records":[{"eventVersion":"2.0","eventSource":"aws:s3","awsRegion":"","eventTime":"2016–09–08T22:34:38.226Z","eventName":"s3:ObjectCreated:Put","userIdentity":{"principalId":"obstor"},"requestParameters":{"sourceIPAddress":"10.1.10.150:44576"},"responseElements":{},"s3":{"s3SchemaVersion":"1.0","configurationId":"Config","bucket":{"name":"images","ownerIdentity":{"principalId":"obstor"},"arn":"arn:aws:s3:::images"},"object":{"key":"myphoto.jpg","size":200436,"sequencer":"147279EAF9F40933"}}}],"level":"info","msg":"","time":"2016–09–08T15:34:38–07:00"}'
```

<a name="MQTT"></a>

## Publish Obstor events MQTT

Install an MQTT Broker from [here](https://mosquitto.org/).

### Step 1: Add MQTT endpoint to Obstor

The MQTT configuration is located as `notify_mqtt` key. Create a configuration key-value pair here for your MQTT instance. The key is a name for your MQTT endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_mqtt[:name]  publish bucket notifications to MQTT endpoints

ARGS:
broker*              (uri)       MQTT server endpoint e.g. `tcp://localhost:1883`
topic*               (string)    name of the MQTT topic to publish
username             (string)    MQTT username
password             (string)    MQTT password
qos                  (number)    set the quality of service priority, defaults to '0'
keep_alive_interval  (duration)  keep-alive interval for MQTT connections in s,m,h,d
reconnect_interval   (duration)  reconnect interval for MQTT connections in s,m,h,d
queue_dir            (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit          (number)    maximum limit for undelivered messages, defaults to '100000'
comment              (sentence)  optionally add a comment to this setting
```

or environment variables

```
KEY:
notify_mqtt[:name]  publish bucket notifications to MQTT endpoints

ARGS:
OBSTOR_NOTIFY_MQTT_ENABLE*              (on|off)    enable notify_mqtt target, default is 'off'
OBSTOR_NOTIFY_MQTT_BROKER*              (uri)       MQTT server endpoint e.g. `tcp://localhost:1883`
OBSTOR_NOTIFY_MQTT_TOPIC*               (string)    name of the MQTT topic to publish
OBSTOR_NOTIFY_MQTT_USERNAME             (string)    MQTT username
OBSTOR_NOTIFY_MQTT_PASSWORD             (string)    MQTT password
OBSTOR_NOTIFY_MQTT_QOS                  (number)    set the quality of service priority, defaults to '0'
OBSTOR_NOTIFY_MQTT_KEEP_ALIVE_INTERVAL  (duration)  keep-alive interval for MQTT connections in s,m,h,d
OBSTOR_NOTIFY_MQTT_RECONNECT_INTERVAL   (duration)  reconnect interval for MQTT connections in s,m,h,d
OBSTOR_NOTIFY_MQTT_QUEUE_DIR            (path)      staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_MQTT_QUEUE_LIMIT          (number)    maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_MQTT_COMMENT              (sentence)  optionally add a comment to this setting
```

Obstor supports persistent event store. The persistent store will backup events when the MQTT broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

To update the configuration, use `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor/ notify_mqtt
notify_mqtt:1 broker="" password="" queue_dir="" queue_limit="0" reconnect_interval="0s"  keep_alive_interval="0s" qos="0" topic="" username=""
```

Use `mc admin config set` command to update the configuration for the deployment. Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:mqtt` at start-up if there were no errors.

```bash
$ mc admin config set myobstor notify_mqtt:1 broker="tcp://localhost:1883" password="" queue_dir="" queue_limit="0" reconnect_interval="0s"  keep_alive_interval="0s" qos="1" topic="obstor" username=""
```

Obstor supports any MQTT server that supports MQTT 3.1 or 3.1.1 and can connect to them over TCP, TLS, or a Websocket connection using `tcp://`, `tls://`, or `ws://` respectively as the scheme for the broker url. See the [Go Client](http://www.eclipse.org/paho/clients/golang/) documentation for more information.

Note that, you can add as many MQTT server endpoint configurations as needed by providing an identifier (like "1" in the example above) for the MQTT instance and an object of per-server configuration parameters.

### Step 2: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded or deleted `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:mqtt`.

```bash
mc mb myobstor/images
mc event add  myobstor/images arn:obstor:sqs::1:mqtt --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:amqp s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 3: Test on MQTT

The python program below waits on mqtt topic `/obstor` and prints event notifications on the console. We use [paho-mqtt](https://pypi.python.org/pypi/paho-mqtt/) library to do this.

```python
#!/usr/bin/env python3
from __future__ import print_function
import paho.mqtt.client as mqtt

# This is the Subscriber

def on_connect(client, userdata, flags, rc):
  print("Connected with result code "+str(rc))
  # qos level is set to 1
  client.subscribe("obstor", 1)

def on_message(client, userdata, msg):
    print(msg.payload)

# client_id is a randomly generated unique ID for the mqtt broker to identify the connection.
client = mqtt.Client(client_id="myclientid",clean_session=False)

client.on_connect = on_connect
client.on_message = on_message

client.connect("localhost",1883,60)
client.loop_forever()
```

Execute this example python program to watch for MQTT events on the console.

```bash
python mqtt.py
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

You should receive the following event notification via MQTT once the upload completes.

```bash
python mqtt.py
{“Records”:[{“eventVersion”:”2.0",”eventSource”:”aws:s3",”awsRegion”:”",”eventTime”:”2016–09–08T22:34:38.226Z”,”eventName”:”s3:ObjectCreated:Put”,”userIdentity”:{“principalId”:”obstor”},”requestParameters”:{“sourceIPAddress”:”10.1.10.150:44576"},”responseElements”:{},”s3":{“s3SchemaVersion”:”1.0",”configurationId”:”Config”,”bucket”:{“name”:”images”,”ownerIdentity”:{“principalId”:”obstor”},”arn”:”arn:aws:s3:::images”},”object”:{“key”:”myphoto.jpg”,”size”:200436,”sequencer”:”147279EAF9F40933"}}}],”level”:”info”,”msg”:””,”time”:”2016–09–08T15:34:38–07:00"}
```

<a name="Elasticsearch"></a>

## Publish Obstor events via Elasticsearch

Install [Elasticsearch](https://www.elastic.co/downloads/elasticsearch) server.

This notification target supports two formats: _namespace_ and _access_.

When the _namespace_ format is used, Obstor synchronizes objects in the bucket with documents in the index. For each event in the Obstor, the server creates a document with the bucket and object name from the event as the document ID. Other details of the event are stored in the body of the document. Thus if an existing object is over-written in Obstor, the corresponding document in the Elasticsearch index is updated. If an object is deleted, the corresponding document is deleted from the index.

When the _access_ format is used, Obstor appends events as documents in an Elasticsearch index. For each event, a document with the event details, with the timestamp of document set to the event's timestamp is appended to an index. The ID of the documented is randomly generated by Elasticsearch. No documents are deleted or modified in this format.

The steps below show how to use this notification target in `namespace` format. The other format is very similar and is omitted for brevity.

### Step 1: Ensure minimum requirements are met

Obstor requires a 5.x series version of Elasticsearch. This is the latest major release series. Elasticsearch provides version upgrade migration guidelines [here](https://www.elastic.co/guide/en/elasticsearch/reference/current/setup-upgrade.html).

### Step 2: Add Elasticsearch endpoint to Obstor

The Elasticsearch configuration is located in the `notify_elasticsearch` key. Create a configuration key-value pair here for your Elasticsearch instance. The key is a name for your Elasticsearch endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_elasticsearch[:name]  publish bucket notifications to Elasticsearch endpoints

ARGS:
url*         (url)                Elasticsearch server's address, with optional authentication info
index*       (string)             Elasticsearch index to store/update events, index is auto-created
format*      (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
queue_dir    (path)               staging dir for undelivered messages e.g. '/home/events'
queue_limit  (number)             maximum limit for undelivered messages, defaults to '100000'
username     (string)             username for Elasticsearch basic-auth
password     (string)             password for Elasticsearch basic-auth
comment      (sentence)           optionally add a comment to this setting
```

or environment variables

```
KEY:
notify_elasticsearch[:name]  publish bucket notifications to Elasticsearch endpoints

ARGS:
OBSTOR_NOTIFY_ELASTICSEARCH_ENABLE*      (on|off)             enable notify_elasticsearch target, default is 'off'
OBSTOR_NOTIFY_ELASTICSEARCH_URL*         (url)                Elasticsearch server's address, with optional authentication info
OBSTOR_NOTIFY_ELASTICSEARCH_INDEX*       (string)             Elasticsearch index to store/update events, index is auto-created
OBSTOR_NOTIFY_ELASTICSEARCH_FORMAT*      (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
OBSTOR_NOTIFY_ELASTICSEARCH_QUEUE_DIR    (path)               staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_ELASTICSEARCH_QUEUE_LIMIT  (number)             maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_ELASTICSEARCH_USERNAME     (string)             username for Elasticsearch basic-auth
OBSTOR_NOTIFY_ELASTICSEARCH_PASSWORD     (string)             password for Elasticsearch basic-auth
OBSTOR_NOTIFY_ELASTICSEARCH_COMMENT      (sentence)           optionally add a comment to this setting
```

For example: `http://localhost:9200` or with authentication info `http://elastic:MagicWord@127.0.0.1:9200`.

Obstor supports persistent event store. The persistent store will backup events when the Elasticsearch broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

If Elasticsearch has authentication enabled, the credentials can be supplied to Obstor via the `url` parameter formatted as `PROTO://USERNAME:PASSWORD@ELASTICSEARCH_HOST:PORT`.

To update the configuration, use `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor/ notify_elasticsearch
notify_elasticsearch:1 queue_limit="0"  url="" format="namespace" index="" queue_dir=""
```

Use `mc admin config set` command to update the configuration for the deployment. Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:elasticsearch` at start-up if there were no errors.

```bash
$ mc admin config set myobstor notify_elasticsearch:1 queue_limit="0"  url="http://127.0.0.1:9200" format="namespace" index="obstor_events" queue_dir="" username="" password=""
```

Note that, you can add as many Elasticsearch server endpoint configurations as needed by providing an identifier (like "1" in the example above) for the Elasticsearch instance and an object of per-server configuration parameters.

### Step 3: Enable bucket notification using Obstor client

We will now enable bucket event notifications on a bucket named `images`. Whenever a JPEG image is created/overwritten, a new document is added or an existing document is updated in the Elasticsearch index configured above. When an existing object is deleted, the corresponding document is deleted from the index. Thus, the rows in the Elasticsearch index, reflect the `.jpg` objects in the `images` bucket.

To configure this bucket notification, we need the ARN printed by Obstor in the previous step. Additional information about ARN is available [here](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html).

With the `mc` tool, the configuration is very simple to add. Let us say that the Obstor server is aliased as `myobstor` in our mc configuration. Execute the following:

```bash
mc mb myobstor/images
mc event add  myobstor/images arn:obstor:sqs::1:elasticsearch --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:elasticsearch s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 4: Test on Elasticsearch

Upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

Use curl to view contents of `obstor_events` index.

```json
$ curl  "http://localhost:9200/obstor_events/_search?pretty=true"
{
  "took" : 40,
  "timed_out" : false,
  "_shards" : {
    "total" : 5,
    "successful" : 5,
    "failed" : 0
  },
  "hits" : {
    "total" : 1,
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "obstor_events",
        "_type" : "event",
        "_id" : "images/myphoto.jpg",
        "_score" : 1.0,
        "_source" : {
          "Records" : [
            {
              "eventVersion" : "2.0",
              "eventSource" : "obstor:s3",
              "awsRegion" : "",
              "eventTime" : "2017-03-30T08:00:41Z",
              "eventName" : "s3:ObjectCreated:Put",
              "userIdentity" : {
                "principalId" : "obstor"
              },
              "requestParameters" : {
                "sourceIPAddress" : "127.0.0.1:38062"
              },
              "responseElements" : {
                "x-amz-request-id" : "14B09A09703FC47B",
                "x-obstor-origin-endpoint" : "http://192.168.86.115:9000"
              },
              "s3" : {
                "s3SchemaVersion" : "1.0",
                "configurationId" : "Config",
                "bucket" : {
                  "name" : "images",
                  "ownerIdentity" : {
                    "principalId" : "obstor"
                  },
                  "arn" : "arn:aws:s3:::images"
                },
                "object" : {
                  "key" : "myphoto.jpg",
                  "size" : 6474,
                  "eTag" : "a3410f4f8788b510d6f19c5067e60a90",
                  "sequencer" : "14B09A09703FC47B"
                }
              },
              "source" : {
                "host" : "127.0.0.1",
                "port" : "38062",
                "userAgent" : "Obstor (linux; amd64) minio-go/2.0.3 mc/2017-02-15T17:57:25Z"
              }
            }
          ]
        }
      }
    ]
  }
}
```

This output shows that a document has been created for the event in Elasticsearch.

Here we see that the document ID is the bucket and object name. In case `access` format was used, the document ID would be automatically generated by Elasticsearch.

<a name="Redis"></a>

## Publish Obstor events via Redis

Install [Redis](http://redis.io/download) server. For illustrative purposes, we have set the database password as "yoursecret".

This notification target supports two formats: _namespace_ and _access_.

When the _namespace_ format is used, Obstor synchronizes objects in the bucket with entries in a hash. For each entry, the key is formatted as "bucketName/objectName" for an object that exists in the bucket, and the value is the JSON-encoded event data about the operation that created/replaced the object in Obstor. When objects are updated or deleted, the corresponding entry in the hash is also updated or deleted.

When the _access_ format is used, Obstor appends events to a list using [RPUSH](https://redis.io/commands/rpush). Each item in the list is a JSON encoded list with two items, where the first item is a timestamp string, and the second item is a JSON object containing event data about the operation that happened in the bucket. No entries appended to the list are updated or deleted by Obstor in this format.

The steps below show how to use this notification target in `namespace` and `access` format.

### Step 1: Add Redis endpoint to Obstor

The Obstor server configuration file is stored on the backend in json format.The Redis configuration is located in the `redis` key under the `notify` top-level key. Create a configuration key-value pair here for your Redis instance. The key is a name for your Redis endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_redis[:name]  publish bucket notifications to Redis datastores

ARGS:
address*     (address)            Redis server's address. For example: `localhost:6379`
key*         (string)             Redis key to store/update events, key is auto-created
format*      (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
password     (string)             Redis server password
queue_dir    (path)               staging dir for undelivered messages e.g. '/home/events'
queue_limit  (number)             maximum limit for undelivered messages, defaults to '100000'
comment      (sentence)           optionally add a comment to this setting
```

or environment variables

```
KEY:
notify_redis[:name]  publish bucket notifications to Redis datastores

ARGS:
OBSTOR_NOTIFY_REDIS_ENABLE*      (on|off)             enable notify_redis target, default is 'off'
OBSTOR_NOTIFY_REDIS_KEY*         (string)             Redis key to store/update events, key is auto-created
OBSTOR_NOTIFY_REDIS_FORMAT*      (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
OBSTOR_NOTIFY_REDIS_PASSWORD     (string)             Redis server password
OBSTOR_NOTIFY_REDIS_QUEUE_DIR    (path)               staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_REDIS_QUEUE_LIMIT  (number)             maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_REDIS_COMMENT      (sentence)           optionally add a comment to this setting
```

Obstor supports persistent event store. The persistent store will backup events when the Redis broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

To update the configuration, use `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor/ notify_redis
notify_redis:1 address="" format="namespace" key="" password="" queue_dir="" queue_limit="0"
```

Use `mc admin config set` command to update the configuration for the deployment.Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:redis` at start-up if there were no errors.

```bash
$ mc admin config set myobstor/ notify_redis:1 address="127.0.0.1:6379" format="namespace" key="bucketevents" password="yoursecret" queue_dir="" queue_limit="0"
```

Note that, you can add as many Redis server endpoint configurations as needed by providing an identifier (like "1" in the example above) for the Redis instance and an object of per-server configuration parameters.

### Step 2: Enable bucket notification using Obstor client

We will now enable bucket event notifications on a bucket named `images`. Whenever a JPEG image is created/overwritten, a new key is added or an existing key is updated in the Redis hash configured above. When an existing object is deleted, the corresponding key is deleted from the Redis hash. Thus, the rows in the Redis hash, reflect the `.jpg` objects in the `images` bucket.

To configure this bucket notification, we need the ARN printed by Obstor in the previous step. Additional information about ARN is available [here](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html).

With the `mc` tool, the configuration is very simple to add. Let us say that the Obstor server is aliased as `myobstor` in our mc configuration. Execute the following:

```bash
mc mb myobstor/images
mc event add myobstor/images arn:obstor:sqs::1:redis --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:redis s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 3: Test on Redis

Start the `redis-cli` Redis client program to inspect the contents in Redis. Run the `monitor` Redis command. This prints each operation performed on Redis as it occurs.

```bash
redis-cli -a yoursecret
127.0.0.1:6379> monitor
OK
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

In the previous terminal, you will now see the operation that Obstor performs on Redis:

```bash
127.0.0.1:6379> monitor
OK
1490686879.650649 [0 172.17.0.1:44710] "PING"
1490686879.651061 [0 172.17.0.1:44710] "HSET" "obstor_events" "images/myphoto.jpg" "{\"Records\":[{\"eventVersion\":\"2.0\",\"eventSource\":\"obstor:s3\",\"awsRegion\":\"\",\"eventTime\":\"2017-03-28T07:41:19Z\",\"eventName\":\"s3:ObjectCreated:Put\",\"userIdentity\":{\"principalId\":\"obstor\"},\"requestParameters\":{\"sourceIPAddress\":\"127.0.0.1:52234\"},\"responseElements\":{\"x-amz-request-id\":\"14AFFBD1ACE5F632\",\"x-obstor-origin-endpoint\":\"http://192.168.86.115:9000\"},\"s3\":{\"s3SchemaVersion\":\"1.0\",\"configurationId\":\"Config\",\"bucket\":{\"name\":\"images\",\"ownerIdentity\":{\"principalId\":\"obstor\"},\"arn\":\"arn:aws:s3:::images\"},\"object\":{\"key\":\"myphoto.jpg\",\"size\":2586,\"eTag\":\"5d284463f9da279f060f0ea4d11af098\",\"sequencer\":\"14AFFBD1ACE5F632\"}},\"source\":{\"host\":\"127.0.0.1\",\"port\":\"52234\",\"userAgent\":\"Obstor (linux; amd64) minio-go/2.0.3 mc/2017-02-15T17:57:25Z\"}}]}"
```

Here we see that Obstor performed `HSET` on `obstor_events` key.

In case, `access` format was used, then `obstor_events` would be a list, and the Obstor server would have performed an `RPUSH` to append to the list. A consumer of this list would ideally use `BLPOP` to remove list items from the left-end of the list.

<a name="NATS"></a>

## Publish Obstor events via NATS

Install NATS from [here](http://nats.io/).

### Step 1: Add NATS endpoint to Obstor

Obstor supports persistent event store. The persistent store will backup events when the NATS broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

```
KEY:
notify_nats[:name]  publish bucket notifications to NATS endpoints

ARGS:
address*                          (address)   NATS server address e.g. '0.0.0.0:4222'
subject*                          (string)    NATS subscription subject
username                          (string)    NATS username
password                          (string)    NATS password
token                             (string)    NATS token
tls                               (on|off)    set to 'on' to enable TLS
tls_skip_verify                   (on|off)    trust server TLS without verification, defaults to "on" (verify)
ping_interval                     (duration)  client ping commands interval in s,m,h,d. Disabled by default
streaming                         (on|off)    set to 'on', to use streaming NATS server
streaming_async                   (on|off)    set to 'on', to enable asynchronous publish
streaming_max_pub_acks_in_flight  (number)    number of messages to publish without waiting for ACKs
streaming_cluster_id              (string)    unique ID for NATS streaming cluster
cert_authority                    (string)    path to certificate chain of the target NATS server
client_cert                       (string)    client cert for NATS mTLS auth
client_key                        (string)    client cert key for NATS mTLS auth
queue_dir                         (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit                       (number)    maximum limit for undelivered messages, defaults to '100000'
comment                           (sentence)  optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_nats[:name]  publish bucket notifications to NATS endpoints

ARGS:
OBSTOR_NOTIFY_NATS_ENABLE*                           (on|off)    enable notify_nats target, default is 'off'
OBSTOR_NOTIFY_NATS_ADDRESS*                          (address)   NATS server address e.g. '0.0.0.0:4222'
OBSTOR_NOTIFY_NATS_SUBJECT*                          (string)    NATS subscription subject
OBSTOR_NOTIFY_NATS_USERNAME                          (string)    NATS username
OBSTOR_NOTIFY_NATS_PASSWORD                          (string)    NATS password
OBSTOR_NOTIFY_NATS_TOKEN                             (string)    NATS token
OBSTOR_NOTIFY_NATS_TLS                               (on|off)    set to 'on' to enable TLS
OBSTOR_NOTIFY_NATS_TLS_SKIP_VERIFY                   (on|off)    trust server TLS without verification, defaults to "on" (verify)
OBSTOR_NOTIFY_NATS_PING_INTERVAL                     (duration)  client ping commands interval in s,m,h,d. Disabled by default
OBSTOR_NOTIFY_NATS_STREAMING                         (on|off)    set to 'on', to use streaming NATS server
OBSTOR_NOTIFY_NATS_STREAMING_ASYNC                   (on|off)    set to 'on', to enable asynchronous publish
OBSTOR_NOTIFY_NATS_STREAMING_MAX_PUB_ACKS_IN_FLIGHT  (number)    number of messages to publish without waiting for ACKs
OBSTOR_NOTIFY_NATS_STREAMING_CLUSTER_ID              (string)    unique ID for NATS streaming cluster
OBSTOR_NOTIFY_NATS_CERT_AUTHORITY                    (string)    path to certificate chain of the target NATS server
OBSTOR_NOTIFY_NATS_CLIENT_CERT                       (string)    client cert for NATS mTLS auth
OBSTOR_NOTIFY_NATS_CLIENT_KEY                        (string)    client cert key for NATS mTLS auth
OBSTOR_NOTIFY_NATS_QUEUE_DIR                         (path)      staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_NATS_QUEUE_LIMIT                       (number)    maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_NATS_COMMENT                           (sentence)  optionally add a comment to this setting
```

To update the configuration, use `mc admin config get` command to get the current configuration file for the obstor deployment.

```bash
$ mc admin config get myobstor/ notify_nats
notify_nats:1 password="yoursecret" streaming_max_pub_acks_in_flight="10" subject="" address="0.0.0.0:4222"  token="" username="yourusername" ping_interval="0" queue_limit="0" tls="off" tls_skip_verify="off" streaming_async="on" queue_dir="" streaming_cluster_id="test-cluster" streaming_enable="on"
```

Use `mc admin config set` command to update the configuration for the deployment.Restart Obstor server to reflect config changes. `bucketevents` is the subject used by NATS in this example.

```bash
$ mc admin config set myobstor notify_nats:1 password="yoursecret" streaming_max_pub_acks_in_flight="10" subject="" address="0.0.0.0:4222"  token="" username="yourusername" ping_interval="0" queue_limit="0" tls="off" streaming_async="on" queue_dir="" streaming_cluster_id="test-cluster" streaming_enable="on"
```

Obstor server also supports [NATS Streaming mode](https://docs.nats.io/nats-concepts/jetstream/streams) that offers additional functionality like `At-least-once-delivery`, and `Publisher rate limiting`. To configure Obstor server to send notifications to NATS Streaming server, update the Obstor server configuration file as follows:

Read more about sections `cluster_id`, `client_id` on [NATS documentation](https://github.com/nats-io/nats-streaming-server/blob/master/README.md). Section `maxPubAcksInflight` is explained [here](https://github.com/nats-io/stan.go#publisher-rate-limiting).

### Step 2: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded or deleted from `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:nats`. To understand more about ARN please follow [AWS ARN](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) documentation.

```bash
mc mb myobstor/images
mc event add myobstor/images arn:obstor:sqs::1:nats --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:nats s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 3: Test on NATS

If you use NATS server, check out this sample program below to log the bucket notification added to NATS.

```go
package main

// Import Go and NATS packages
import (
  "log"
  "runtime"

  "github.com/nats-io/nats.go"
)

func main() {

  // Create server connection
  natsConnection, _ := nats.Connect("nats://yourusername:yoursecret@localhost:4222")
  log.Println("Connected")

  // Subscribe to subject
  log.Printf("Subscribing to subject 'bucketevents'\n")
  natsConnection.Subscribe("bucketevents", func(msg *nats.Msg) {

    // Handle the message
    log.Printf("Received message '%s\n", string(msg.Data)+"'")
  })

  // Keep the connection alive
  runtime.Goexit()
}
```

```bash
go run nats.go
2016/10/12 06:39:18 Connected
2016/10/12 06:39:18 Subscribing to subject 'bucketevents'
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

The example `nats.go` program prints event notification to console.

```bash
go run nats.go
2016/10/12 06:51:26 Connected
2016/10/12 06:51:26 Subscribing to subject 'bucketevents'
2016/10/12 06:51:33 Received message '{"EventType":"s3:ObjectCreated:Put","Key":"images/myphoto.jpg","Records":[{"eventVersion":"2.0","eventSource":"aws:s3","awsRegion":"","eventTime":"2016-10-12T13:51:33Z","eventName":"s3:ObjectCreated:Put","userIdentity":{"principalId":"obstor"},"requestParameters":{"sourceIPAddress":"[::1]:57106"},"responseElements":{},"s3":{"s3SchemaVersion":"1.0","configurationId":"Config","bucket":{"name":"images","ownerIdentity":{"principalId":"obstor"},"arn":"arn:aws:s3:::images"},"object":{"key":"myphoto.jpg","size":56060,"eTag":"1d97bf45ecb37f7a7b699418070df08f","sequencer":"147CCD1AE054BFD0"}}}],"level":"info","msg":"","time":"2016-10-12T06:51:33-07:00"}
```

If you use NATS Streaming server, check out this sample program below to log the bucket notification added to NATS.

```go
package main

// Import Go and NATS packages
import (
  "fmt"
  "runtime"

  "github.com/nats-io/stan.go"
)

func main() {

  var stanConnection stan.Conn

  subscribe := func() {
    fmt.Printf("Subscribing to subject 'bucketevents'\n")
    stanConnection.Subscribe("bucketevents", func(m *stan.Msg) {

      // Handle the message
      fmt.Printf("Received a message: %s\n", string(m.Data))
    })
  }


  stanConnection, _ = stan.Connect("test-cluster", "test-client", stan.NatsURL("nats://yourusername:yoursecret@0.0.0.0:4222"), stan.SetConnectionLostHandler(func(c stan.Conn, _ error) {
    go func() {
      for {
        // Reconnect if the connection is lost.
        if stanConnection == nil || stanConnection.NatsConn() == nil ||  !stanConnection.NatsConn().IsConnected() {
          stanConnection, _ = stan.Connect("test-cluster", "test-client", stan.NatsURL("nats://yourusername:yoursecret@0.0.0.0:4222"), stan.SetConnectionLostHandler(func(c stan.Conn, _ error) {
            if c.NatsConn() != nil {
              c.NatsConn().Close()
            }
            _ = c.Close()
          }))
          if stanConnection != nil {
            subscribe()
          }

        }
      }

    }()
  }))

  // Subscribe to subject
  subscribe()

  // Keep the connection alive
  runtime.Goexit()
}

```

```bash
go run nats.go
2017/07/07 11:47:40 Connected
2017/07/07 11:47:40 Subscribing to subject 'bucketevents'
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

The example `nats.go` program prints event notification to console.

```
Received a message: {"EventType":"s3:ObjectCreated:Put","Key":"images/myphoto.jpg","Records":[{"eventVersion":"2.0","eventSource":"obstor:s3","awsRegion":"","eventTime":"2017-07-07T18:46:37Z","eventName":"s3:ObjectCreated:Put","userIdentity":{"principalId":"obstor"},"requestParameters":{"sourceIPAddress":"192.168.1.80:55328"},"responseElements":{"x-amz-request-id":"14CF20BD1EFD5B93","x-obstor-origin-endpoint":"http://127.0.0.1:9000"},"s3":{"s3SchemaVersion":"1.0","configurationId":"Config","bucket":{"name":"images","ownerIdentity":{"principalId":"obstor"},"arn":"arn:aws:s3:::images"},"object":{"key":"myphoto.jpg","size":248682,"eTag":"f1671feacb8bbf7b0397c6e9364e8c92","contentType":"image/jpeg","userDefined":{"content-type":"image/jpeg"},"versionId":"1","sequencer":"14CF20BD1EFD5B93"}},"source":{"host":"192.168.1.80","port":"55328","userAgent":"Obstor (linux; amd64) minio-go/2.0.4 mc/DEVELOPMENT.GOGET"}}],"level":"info","msg":"","time":"2017-07-07T11:46:37-07:00"}
```

<a name="PostgreSQL"></a>

## Publish Obstor events via PostgreSQL

> NOTE: Until release RELEASE.2026-04-10T03-34-42Z PostgreSQL notification used to support following options:
>
> ```
> host                (hostname)           Postgres server hostname (used only if `connection_string` is empty)
> port                (port)               Postgres server port, defaults to `5432` (used only if `connection_string` is empty)
> username            (string)             database username (used only if `connection_string` is empty)
> password            (string)             database password (used only if `connection_string` is empty)
> database            (string)             database name (used only if `connection_string` is empty)
> ```
>
> These are now deprecated, if you plan to upgrade to any releases after *RELEASE.2026-04-10T03-34-42Z* make sure
> to migrate to only using *connection_string* option. To migrate, once you have upgraded all the servers use the
> following command to update the existing notification targets.
>
> ```bash
> mc admin config set myobstor/ notify_postgres[:name] connection_string="host=hostname port=2832 username=psqluser password=psqlpass database=bucketevents"
> ```
>
> Please make sure this step is carried out, without this step PostgreSQL notification targets will not work,
> an error message will be shown on the console upon server upgrade/restart, make sure to follow the above
> instructions appropriately. For further questions please join our https://pgg.net/discord

Install [PostgreSQL](https://www.postgresql.org/) database server. For illustrative purposes, we have set the "postgres" user password as `password` and created a database called `obstor_events` to store the events.

This notification target supports two formats: _namespace_ and _access_.

When the _namespace_ format is used, Obstor synchronizes objects in the bucket with rows in the table. It creates rows with two columns: key and value. The key is the bucket and object name of an object that exists in Obstor. The value is JSON encoded event data about the operation that created/replaced the object in Obstor. When objects are updated or deleted, the corresponding row from this table is updated or deleted respectively.

When the _access_ format is used, Obstor appends events to a table. It creates rows with two columns: event_time and event_data. The event_time is the time at which the event occurred in the Obstor server. The event_data is the JSON encoded event data about the operation on an object. No rows are deleted or modified in this format.

The steps below show how to use this notification target in `namespace` format. The other format is very similar and is omitted for brevity.

### Step 1: Ensure minimum requirements are met

Obstor requires PostgreSQL version 17 or above. Obstor uses the [`INSERT ON CONFLICT`](https://www.postgresql.org/docs/9.5/static/sql-insert.html#SQL-ON-CONFLICT) (aka UPSERT) feature, introduced in version 9.5 and the [JSONB](https://www.postgresql.org/docs/9.4/static/datatype-json.html) data-type introduced in version 9.4.

### Step 2: Add PostgreSQL endpoint to Obstor

The PostgreSQL configuration is located in the `notify_postgresql` key. Create a configuration key-value pair here for your PostgreSQL instance. The key is a name for your PostgreSQL endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_postgres[:name]  publish bucket notifications to Postgres databases

ARGS:
connection_string*   (string)             Postgres server connection-string e.g. "host=localhost port=5432 dbname=obstor_events user=postgres password=password sslmode=disable"
table*               (string)             DB table name to store/update events, table is auto-created
format*              (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
queue_dir            (path)               staging dir for undelivered messages e.g. '/home/events'
queue_limit          (number)             maximum limit for undelivered messages, defaults to '100000'
max_open_connections (number)             maximum number of open connections to the database, defaults to '2'
comment              (sentence)           optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_postgres[:name]  publish bucket notifications to Postgres databases

ARGS:
OBSTOR_NOTIFY_POSTGRES_ENABLE*              (on|off)             enable notify_postgres target, default is 'off'
OBSTOR_NOTIFY_POSTGRES_CONNECTION_STRING*   (string)             Postgres server connection-string e.g. "host=localhost port=5432 dbname=obstor_events user=postgres password=password sslmode=disable"
OBSTOR_NOTIFY_POSTGRES_TABLE*               (string)             DB table name to store/update events, table is auto-created
OBSTOR_NOTIFY_POSTGRES_FORMAT*              (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
OBSTOR_NOTIFY_POSTGRES_QUEUE_DIR            (path)               staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_POSTGRES_QUEUE_LIMIT          (number)             maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_POSTGRES_COMMENT              (sentence)           optionally add a comment to this setting
OBSTOR_NOTIFY_POSTGRES_MAX_OPEN_CONNECTIONS (number)             maximum number of open connections to the database, defaults to '2'
```

> NOTE: If the `max_open_connections` key or the environment variable `OBSTOR_NOTIFY_POSTGRES_MAX_OPEN_CONNECTIONS` is set to `0`, There will be no limit set on the number of
> open connections to the database. This setting is generally NOT recommended as the behavior may be inconsistent during recursive deletes in `namespace` format.

Obstor supports persistent event store. The persistent store will backup events when the PostgreSQL connection goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

Note that for illustration here, we have disabled SSL. In the interest of security, for production this is not recommended.
To update the configuration, use `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor notify_postgres
notify_postgres:1 queue_dir="" connection_string="" queue_limit="0"  table="" format="namespace"
```

Use `mc admin config set` command to update the configuration for the deployment. Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:postgresql` at start-up if there were no errors.

```bash
$ mc admin config set myobstor notify_postgres:1 connection_string="host=localhost port=5432 dbname=obstor_events user=postgres password=password sslmode=disable" table="bucketevents" format="namespace"
```

Note that, you can add as many PostgreSQL server endpoint configurations as needed by providing an identifier (like "1" in the example above) for the PostgreSQL instance and an object of per-server configuration parameters.

### Step 3: Enable bucket notification using Obstor client

We will now enable bucket event notifications on a bucket named `images`. Whenever a JPEG image is created/overwritten, a new row is added or an existing row is updated in the PostgreSQL configured above. When an existing object is deleted, the corresponding row is deleted from the PostgreSQL table. Thus, the rows in the PostgreSQL table, reflect the `.jpg` objects in the `images` bucket.

To configure this bucket notification, we need the ARN printed by Obstor in the previous step. Additional information about ARN is available [here](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html).

With the `mc` tool, the configuration is very simple to add. Let us say that the Obstor server is aliased as `myobstor` in our mc configuration. Execute the following:

```bash
# Create bucket named `images` in myobstor
mc mb myobstor/images
# Add notification configuration on the `images` bucket using the MySQL ARN. The --suffix argument filters events.
mc event add myobstor/images arn:obstor:sqs::1:postgresql --suffix .jpg
# Print out the notification configuration on the `images` bucket.
mc event list myobstor/images
mc event list myobstor/images
arn:obstor:sqs::1:postgresql s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 4: Test on PostgreSQL

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

Open PostgreSQL terminal to list the rows in the `bucketevents` table.

```bash
$ psql -h 127.0.0.1 -U postgres -d obstor_events
obstor_events=# select * from bucketevents;

key                 |                      value
--------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 images/myphoto.jpg | {"Records": [{"s3": {"bucket": {"arn": "arn:aws:s3:::images", "name": "images", "ownerIdentity": {"principalId": "obstor"}}, "object": {"key": "myphoto.jpg", "eTag": "1d97bf45ecb37f7a7b699418070df08f", "size": 56060, "sequencer": "147CE57C70B31931"}, "configurationId": "Config", "s3SchemaVersion": "1.0"}, "awsRegion": "", "eventName": "s3:ObjectCreated:Put", "eventTime": "2016-10-12T21:18:20Z", "eventSource": "aws:s3", "eventVersion": "2.0", "userIdentity": {"principalId": "obstor"}, "responseElements": {}, "requestParameters": {"sourceIPAddress": "[::1]:39706"}}]}
(1 row)
```

<a name="MySQL"></a>

## Publish Obstor events via MySQL

> NOTE: Until release RELEASE.2026-04-10T03-34-42Z MySQL notification used to support following options:
>
> ```
> host         (hostname)           MySQL server hostname (used only if `dsn_string` is empty)
> port         (port)               MySQL server port (used only if `dsn_string` is empty)
> username     (string)             database username (used only if `dsn_string` is empty)
> password     (string)             database password (used only if `dsn_string` is empty)
> database     (string)             database name (used only if `dsn_string` is empty)
> ```
>
> These are now deprecated, if you plan to upgrade to any releases after *RELEASE.2026-04-10T03-34-42Z* make sure
> to migrate to only using *dsn_string* option. To migrate, once you have upgraded all the servers use the
> following command to update the existing notification targets.
>
> ```bash
> mc admin config set myobstor/ notify_mysql[:name] dsn_string="mysqluser:mysqlpass@tcp(localhost:2832)/bucketevents"
> ```
>
> Please make sure this step is carried out, without this step MySQL notification targets will not work,
> an error message will be shown on the console upon server upgrade/restart, make sure to follow the above
> instructions appropriately. For further questions please join our https://pgg.net/discord

Install MySQL from [here](https://dev.mysql.com/downloads/mysql/). For illustrative purposes, we have set the root password as `password` and created a database called `obstordb` to store the events.

This notification target supports two formats: _namespace_ and _access_.

When the _namespace_ format is used, Obstor synchronizes objects in the bucket with rows in the table. It creates rows with two columns: key_name and value. The key_name is the bucket and object name of an object that exists in Obstor. The value is JSON encoded event data about the operation that created/replaced the object in Obstor. When objects are updated or deleted, the corresponding row from this table is updated or deleted respectively.

When the _access_ format is used, Obstor appends events to a table. It creates rows with two columns: event_time and event_data. The event_time is the time at which the event occurred in the Obstor server. The event_data is the JSON encoded event data about the operation on an object. No rows are deleted or modified in this format.

The steps below show how to use this notification target in `namespace` format. The other format is very similar and is omitted for brevity.

### Step 1: Ensure minimum requirements are met

Obstor requires MySQL version 8.0 or above. Obstor uses the [JSON](https://dev.mysql.com/doc/refman/8.4/en/json.html) data-type introduced in version 5.7.8. We tested this setup on MySQL 8.4.

### Step 2: Add MySQL server endpoint configuration to Obstor

The MySQL configuration is located in the `notify_mysql` key. Create a configuration key-value pair here for your MySQL instance. The key is a name for your MySQL endpoint, and the value is a collection of key-value parameters described in the table below.

```
KEY:
notify_mysql[:name]  publish bucket notifications to MySQL databases. When multiple MySQL server endpoints are needed, a user specified "name" can be added for each configuration, (e.g."notify_mysql:myinstance").

ARGS:
dsn_string*          (string)             MySQL data-source-name connection string e.g. "<user>:<password>@tcp(<host>:<port>)/<database>"
table*               (string)             DB table name to store/update events, table is auto-created
format*              (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
queue_dir            (path)               staging dir for undelivered messages e.g. '/home/events'
queue_limit          (number)             maximum limit for undelivered messages, defaults to '100000'
max_open_connections (number)             maximum number of open connections to the database, defaults to '2'
comment              (sentence)           optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_mysql[:name]  publish bucket notifications to MySQL databases

ARGS:
OBSTOR_NOTIFY_MYSQL_ENABLE*              (on|off)             enable notify_mysql target, default is 'off'
OBSTOR_NOTIFY_MYSQL_DSN_STRING*          (string)             MySQL data-source-name connection string e.g. "<user>:<password>@tcp(<host>:<port>)/<database>"
OBSTOR_NOTIFY_MYSQL_TABLE*               (string)             DB table name to store/update events, table is auto-created
OBSTOR_NOTIFY_MYSQL_FORMAT*              (namespace*|access)  'namespace' reflects current bucket/object list and 'access' reflects a journal of object operations, defaults to 'namespace'
OBSTOR_NOTIFY_MYSQL_QUEUE_DIR            (path)               staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_MYSQL_QUEUE_LIMIT          (number)             maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_MYSQL_MAX_OPEN_CONNECTIONS (number)             maximum number of open connections to the database, defaults to '2'
OBSTOR_NOTIFY_MYSQL_COMMENT              (sentence)           optionally add a comment to this setting
```

> NOTE: If the `max_open_connections` key or the environment variable `OBSTOR_NOTIFY_MYSQL_MAX_OPEN_CONNECTIONS` is set to `0`, There will be no limit set on the number of
> open connections to the database. This setting is generally NOT recommended as the behavior may be inconsistent during recursive deletes in `namespace` format.

`dsn_string` is required and is of form `"<user>:<password>@tcp(<host>:<port>)/<database>"`

Obstor supports persistent event store. The persistent store will backup events if MySQL connection goes offline and then replays the stored events when the broken connection comes back up. The event store can be configured by setting a directory path in `queue_dir` field, and the maximum number of events, which can be stored in a `queue_dir`, in `queue_limit` field. For example, `queue_dir` can be set to `/home/events` and `queue_limit` can be set to `1000`. By default, the `queue_limit` is set to `100000`.

Before updating the configuration, let's start with `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor/ notify_mysql
notify_mysql:myinstance enable=off format=namespace host= port= username= password= database= dsn_string= table= queue_dir= queue_limit=0
```

Use `mc admin config set` command to update MySQL notification configuration for the deployment with `dsn_string` parameter:

```bash
$ mc admin config set myobstor notify_mysql:myinstance table="obstor_images" dsn_string="root:xxxx@tcp(172.17.0.1:3306)/obstordb"
```

Note that, you can add as many MySQL server endpoint configurations as needed by providing an identifier (like "myinstance" in the example above) for each MySQL instance desired.

Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::myinstance:mysql` at start-up, if there are no errors.

### Step 3: Enable bucket notification using Obstor client

We will now setup bucket notifications on a bucket named `images`. Whenever a JPEG image object is created/overwritten, a new row is added or an existing row is updated in the MySQL table configured above. When an existing object is deleted, the corresponding row is deleted from the MySQL table. Thus, the rows in the MySQL table, reflect the `.jpg` objects in the `images` bucket.

To configure this bucket notification, we need the ARN printed by Obstor in the previous step. Additional information about ARN is available [here](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html).

With the `mc` tool, the configuration is very simple to add. Let us say that the Obstor server is aliased as `myobstor` in our mc configuration. Execute the following:

```bash
# Create bucket named `images` in myobstor
mc mb myobstor/images
# Add notification configuration on the `images` bucket using the MySQL ARN. The --suffix argument filters events.
mc event add myobstor/images arn:obstor:sqs::myinstance:mysql --suffix .jpg
# Print out the notification configuration on the `images` bucket.
mc event list myobstor/images
arn:obstor:sqs::myinstance:mysql s3:ObjectCreated:*,s3:ObjectRemoved:*,s3:ObjectAccessed:* Filter: suffix=”.jpg”
```

### Step 4: Test on MySQL

Open another terminal and upload a JPEG image into `images` bucket:

```bash
mc cp myphoto.jpg myobstor/images
```

Open MySQL terminal and list the rows in the `obstor_images` table.

```bash
$ mysql -h 172.17.0.1 -P 3306 -u root -p obstordb
mysql> select * from obstor_images;
+--------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| key_name           | value                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
+--------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
| images/myphoto.jpg | {"Records": [{"s3": {"bucket": {"arn": "arn:aws:s3:::images", "name": "images", "ownerIdentity": {"principalId": "obstor"}}, "object": {"key": "myphoto.jpg", "eTag": "467886be95c8ecfd71a2900e3f461b4f", "size": 26, "sequencer": "14AC59476F809FD3"}, "configurationId": "Config", "s3SchemaVersion": "1.0"}, "awsRegion": "", "eventName": "s3:ObjectCreated:Put", "eventTime": "2017-03-16T11:29:00Z", "eventSource": "aws:s3", "eventVersion": "2.0", "userIdentity": {"principalId": "obstor"}, "responseElements": {"x-amz-request-id": "14AC59476F809FD3", "x-obstor-origin-endpoint": "http://192.168.86.110:9000"}, "requestParameters": {"sourceIPAddress": "127.0.0.1:38260"}}]} |
+--------------------+----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
1 row in set (0.01 sec)

```

<a name="apache-kafka"></a>

## Publish Obstor events via Kafka

Install Apache Kafka from [here](http://kafka.apache.org/).

### Step 1: Ensure minimum requirements are met

Obstor requires Kafka version 0.10 or 0.9. Internally Obstor uses the [Shopify/sarama](https://github.com/IBM/sarama/) library and so has the same version compatibility as provided by this library.

### Step 2: Add Kafka endpoint to Obstor

Obstor supports persistent event store. The persistent store will backup events when the kafka broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

```
KEY:
notify_kafka[:name]  publish bucket notifications to Kafka endpoints

ARGS:
brokers*         (csv)       comma separated list of Kafka broker addresses
topic            (string)    Kafka topic used for bucket notifications
sasl_username    (string)    username for SASL/PLAIN or SASL/SCRAM authentication
sasl_password    (string)    password for SASL/PLAIN or SASL/SCRAM authentication
sasl_mechanism   (string)    sasl authentication mechanism, default 'PLAIN'
tls_client_auth  (string)    clientAuth determines the Kafka server's policy for TLS client auth
sasl             (on|off)    set to 'on' to enable SASL authentication
tls              (on|off)    set to 'on' to enable TLS
tls_skip_verify  (on|off)    trust server TLS without verification, defaults to "on" (verify)
client_tls_cert  (path)      path to client certificate for mTLS auth
client_tls_key   (path)      path to client key for mTLS auth
queue_dir        (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit      (number)    maximum limit for undelivered messages, defaults to '100000'
version          (string)    specify the version of the Kafka cluster e.g '2.2.0'
comment          (sentence)  optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_kafka[:name]  publish bucket notifications to Kafka endpoints

ARGS:
OBSTOR_NOTIFY_KAFKA_ENABLE*          (on|off)                enable notify_kafka target, default is 'off'
OBSTOR_NOTIFY_KAFKA_BROKERS*         (csv)                   comma separated list of Kafka broker addresses
OBSTOR_NOTIFY_KAFKA_TOPIC            (string)                Kafka topic used for bucket notifications
OBSTOR_NOTIFY_KAFKA_SASL_USERNAME    (string)                username for SASL/PLAIN or SASL/SCRAM authentication
OBSTOR_NOTIFY_KAFKA_SASL_PASSWORD    (string)                password for SASL/PLAIN or SASL/SCRAM authentication
OBSTOR_NOTIFY_KAFKA_SASL_MECHANISM   (plain*|sha256|sha512)  sasl authentication mechanism, default 'plain'
OBSTOR_NOTIFY_KAFKA_TLS_CLIENT_AUTH  (string)                clientAuth determines the Kafka server's policy for TLS client auth
OBSTOR_NOTIFY_KAFKA_SASL             (on|off)                set to 'on' to enable SASL authentication
OBSTOR_NOTIFY_KAFKA_TLS              (on|off)                set to 'on' to enable TLS
OBSTOR_NOTIFY_KAFKA_TLS_SKIP_VERIFY  (on|off)                trust server TLS without verification, defaults to "on" (verify)
OBSTOR_NOTIFY_KAFKA_CLIENT_TLS_CERT  (path)                  path to client certificate for mTLS auth
OBSTOR_NOTIFY_KAFKA_CLIENT_TLS_KEY   (path)                  path to client key for mTLS auth
OBSTOR_NOTIFY_KAFKA_QUEUE_DIR        (path)                  staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_KAFKA_QUEUE_LIMIT      (number)                maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_KAFKA_COMMENT          (sentence)              optionally add a comment to this setting
OBSTOR_NOTIFY_KAFKA_VERSION          (string)                specify the version of the Kafka cluster e.g. '2.2.0'
```

To update the configuration, use `mc admin config get` command to get the current configuration.

```bash
$ mc admin config get myobstor/ notify_kafka
notify_kafka:1 tls_skip_verify="off"  queue_dir="" queue_limit="0" sasl="off" sasl_password="" sasl_username="" tls_client_auth="0" tls="off" brokers="" topic="" client_tls_cert="" client_tls_key="" version=""
```

Use `mc admin config set` command to update the configuration for the deployment. Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:kafka` at start-up if there were no errors.`bucketevents` is the topic used by kafka in this example.

```bash
$ mc admin config set myobstor notify_kafka:1 tls_skip_verify="off"  queue_dir="" queue_limit="0" sasl="off" sasl_password="" sasl_username="" tls_client_auth="0" tls="off" client_tls_cert="" client_tls_key="" brokers="localhost:9092,localhost:9093" topic="bucketevents" version=""
```

### Step 3: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded or deleted from `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:kafka`. To understand more about ARN please follow [AWS ARN](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) documentation.

```bash
mc mb myobstor/images
mc event add  myobstor/images arn:obstor:sqs::1:kafka --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:kafka s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 4: Test on Kafka

We used [kafkacat](https://github.com/edenhill/kafkacat) to print all notifications on the console.

```bash
kafkacat -C -b localhost:9092 -t bucketevents
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp myphoto.jpg myobstor/images
```

`kafkacat` prints the event notification to the console.

```json
kafkacat -b localhost:9092 -t bucketevents
{
  "EventName": "s3:ObjectCreated:Put",
  "Key": "images/myphoto.jpg",
  "Records": [
    {
      "eventVersion": "2.0",
      "eventSource": "obstor:s3",
      "awsRegion": "",
      "eventTime": "2026-09-10T17:41:54Z",
      "eventName": "s3:ObjectCreated:Put",
      "userIdentity": {
        "principalId": "AKIAIOSFODNN7EXAMPLE"
      },
      "requestParameters": {
        "accessKey": "AKIAIOSFODNN7EXAMPLE",
        "region": "",
        "sourceIPAddress": "192.168.56.192"
      },
      "responseElements": {
        "x-amz-request-id": "15C3249451E12784",
        "x-obstor-deployment-id": "751a8ba6-acb2-42f6-a297-4cdf1cf1fa4f",
        "x-obstor-origin-endpoint": "http://192.168.97.83:9000"
      },
      "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "Config",
        "bucket": {
          "name": "images",
          "ownerIdentity": {
            "principalId": "AKIAIOSFODNN7EXAMPLE"
          },
          "arn": "arn:aws:s3:::images"
        },
        "object": {
          "key": "myphoto.jpg",
          "size": 6474,
          "eTag": "430f89010c77aa34fc8760696da62d08-1",
          "contentType": "image/jpeg",
          "userMetadata": {
            "content-type": "image/jpeg"
          },
          "versionId": "1",
          "sequencer": "15C32494527B46C5"
        }
      },
      "source": {
        "host": "192.168.56.192",
        "port": "",
        "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0"
      }
    }
  ]
}
```

<a name="webhooks"></a>

## Publish Obstor events via Webhooks

[Webhooks](https://en.wikipedia.org/wiki/Webhook) are a way to receive information when it happens, rather than continually polling for that data.

### Step 1: Add Webhook endpoint to Obstor

Obstor supports persistent event store. The persistent store will backup events when the webhook goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

```
KEY:
notify_webhook[:name]  publish bucket notifications to webhook endpoints

ARGS:
endpoint*    (url)       webhook server endpoint e.g. http://localhost:8080/obstor/events
auth_token   (string)    opaque string or JWT authorization token
queue_dir    (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit  (number)    maximum limit for undelivered messages, defaults to '100000'
client_cert  (string)    client cert for Webhook mTLS auth
client_key   (string)    client cert key for Webhook mTLS auth
comment      (sentence)  optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_webhook[:name]  publish bucket notifications to webhook endpoints

ARGS:
OBSTOR_NOTIFY_WEBHOOK_ENABLE*      (on|off)    enable notify_webhook target, default is 'off'
OBSTOR_NOTIFY_WEBHOOK_ENDPOINT*    (url)       webhook server endpoint e.g. http://localhost:8080/obstor/events
OBSTOR_NOTIFY_WEBHOOK_AUTH_TOKEN   (string)    opaque string or JWT authorization token
OBSTOR_NOTIFY_WEBHOOK_QUEUE_DIR    (path)      staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_WEBHOOK_QUEUE_LIMIT  (number)    maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_WEBHOOK_COMMENT      (sentence)  optionally add a comment to this setting
OBSTOR_NOTIFY_WEBHOOK_CLIENT_CERT  (string)    client cert for Webhook mTLS auth
OBSTOR_NOTIFY_WEBHOOK_CLIENT_KEY   (string)    client cert key for Webhook mTLS auth
```

```bash
$ mc admin config get myobstor/ notify_webhook
notify_webhook:1 endpoint="" auth_token="" queue_limit="0" queue_dir="" client_cert="" client_key=""
```

Use `mc admin config set` command to update the configuration for the deployment. Here the endpoint is the server listening for webhook notifications. Save the settings and restart the Obstor server for changes to take effect. Note that the endpoint needs to be live and reachable when you restart your Obstor server.

```bash
$ mc admin config set myobstor notify_webhook:1 queue_limit="0"  endpoint="http://localhost:3000" queue_dir=""
```

### Step 2: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded to `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:webhook`. To learn more about ARN please follow [AWS ARN](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) documentation.

```bash
mc mb myobstor/images
mc mb myobstor/images-thumbnail
mc event add myobstor/images arn:obstor:sqs::1:webhook --event put --suffix .jpg
```

Check if event notification is successfully configured by

```bash
mc event list myobstor/images
```

You should get a response like this

```
arn:obstor:sqs::1:webhook   s3:ObjectCreated:*   Filter: suffix=".jpg"
```

### Step 3: Test with Thumbnailer

We used [Thumbnailer](https://github.com/minio/thumbnailer) to listen for Obstor notifications when a new JPEG file is uploaded (HTTP PUT). Triggered by a notification, Thumbnailer uploads a thumbnail of new image to Obstor server. To start with, download and install Thumbnailer.

```bash
git clone https://github.com/minio/thumbnailer/
npm install
```

Then open the Thumbnailer config file at `config/webhook.json` and add the configuration for your Obstor server and then start Thumbnailer by

```bash
NODE_ENV=webhook node thumbnail-webhook.js
```

Thumbnailer starts running at `http://localhost:3000/`. Next, configure the Obstor server to send notifications to this URL (as mentioned in step 1) and use `mc` to set up bucket notifications (as mentioned in step 2). Then upload a JPEG image to Obstor server by

```bash
mc cp ~/images.jpg myobstor/images
.../images.jpg:  8.31 KB / 8.31 KB ┃███████████████████████████████████┃ 100.00% 59.42 KB/s 0s
```

Wait a few moments, then check the bucket’s contents with mc ls — you will see a thumbnail appear.

```bash
mc ls myobstor/images-thumbnail
[2017-02-08 11:39:40 IST]   992B images-thumbnail.jpg
```

<a name="NSQ"></a>

## Publish Obstor events to NSQ

Install an NSQ Daemon from [here](https://nsq.io/). Or use the following Docker
command for starting an nsq daemon:

```bash
docker run --rm -p 4150-4151:4150-4151 nsqio/nsq /nsqd
```

### Step 1: Add NSQ endpoint to Obstor

Obstor supports persistent event store. The persistent store will backup events when the NSQ broker goes offline and replays it when the broker comes back online. The event store can be configured by setting the directory path in `queue_dir` field and the maximum limit of events in the queue_dir in `queue_limit` field. For eg, the `queue_dir` can be `/home/events` and `queue_limit` can be `1000`. By default, the `queue_limit` is set to 100000.

To update the configuration, use `mc admin config get` command to get the current configuration for `notify_nsq`.

```
KEY:
notify_nsq[:name]  publish bucket notifications to NSQ endpoints

ARGS:
nsqd_address*    (address)   NSQ server address e.g. '127.0.0.1:4150'
topic*           (string)    NSQ topic
tls              (on|off)    set to 'on' to enable TLS
tls_skip_verify  (on|off)    trust server TLS without verification, defaults to "on" (verify)
queue_dir        (path)      staging dir for undelivered messages e.g. '/home/events'
queue_limit      (number)    maximum limit for undelivered messages, defaults to '100000'
comment          (sentence)  optionally add a comment to this setting
```

or environment variables
```
KEY:
notify_nsq[:name]  publish bucket notifications to NSQ endpoints

ARGS:
OBSTOR_NOTIFY_NSQ_ENABLE*          (on|off)    enable notify_nsq target, default is 'off'
OBSTOR_NOTIFY_NSQ_NSQD_ADDRESS*    (address)   NSQ server address e.g. '127.0.0.1:4150'
OBSTOR_NOTIFY_NSQ_TOPIC*           (string)    NSQ topic
OBSTOR_NOTIFY_NSQ_TLS              (on|off)    set to 'on' to enable TLS
OBSTOR_NOTIFY_NSQ_TLS_SKIP_VERIFY  (on|off)    trust server TLS without verification, defaults to "on" (verify)
OBSTOR_NOTIFY_NSQ_QUEUE_DIR        (path)      staging dir for undelivered messages e.g. '/home/events'
OBSTOR_NOTIFY_NSQ_QUEUE_LIMIT      (number)    maximum limit for undelivered messages, defaults to '100000'
OBSTOR_NOTIFY_NSQ_COMMENT          (sentence)  optionally add a comment to this setting
```

```bash
$ mc admin config get myobstor/ notify_nsq
notify_nsq:1 nsqd_address="" queue_dir="" queue_limit="0"  tls="off" tls_skip_verify="off" topic=""
```

Use `mc admin config set` command to update the configuration for the deployment. Restart the Obstor server to put the changes into effect. The server will print a line like `SQS ARNs: arn:obstor:sqs::1:nsq` at start-up if there were no errors.

```bash
$ mc admin config set myobstor notify_nsq:1 nsqd_address="127.0.0.1:4150" queue_dir="" queue_limit="0" tls="off" tls_skip_verify="on" topic="obstor"
```

Note that, you can add as many NSQ daemon endpoint configurations as needed by providing an identifier (like "1" in the example above) for the NSQ instance and an object of per-server configuration parameters.

### Step 2: Enable bucket notification using Obstor client

We will enable bucket event notification to trigger whenever a JPEG image is uploaded or deleted `images` bucket on `myobstor` server. Here ARN value is `arn:obstor:sqs::1:nsq`.

```bash
mc mb myobstor/images
mc event add  myobstor/images arn:obstor:sqs::1:nsq --suffix .jpg
mc event list myobstor/images
arn:obstor:sqs::1:nsq s3:ObjectCreated:*,s3:ObjectRemoved:* Filter: suffix=”.jpg”
```

### Step 3: Test on NSQ

The simplest test is to download `nsq_tail` from [nsq github](https://github.com/nsqio/nsq/releases)

```bash
./nsq_tail -nsqd-tcp-address 127.0.0.1:4150 -topic obstor
```

Open another terminal and upload a JPEG image into `images` bucket.

```bash
mc cp gopher.jpg myobstor/images
```

You should receive the following event notification via NSQ once the upload completes.

```json
{"EventName":"s3:ObjectCreated:Put","Key":"images/gopher.jpg","Records":[{"eventVersion":"2.0","eventSource":"obstor:s3","awsRegion":"","eventTime":"2018-10-31T09:31:11Z","eventName":"s3:ObjectCreated:Put","userIdentity":{"principalId":"21EJ9HYV110O8NVX2VMS"},"requestParameters":{"sourceIPAddress":"10.1.1.1"},"responseElements":{"x-amz-request-id":"1562A792DAA53426","x-obstor-origin-endpoint":"http://10.0.3.1:9000"},"s3":{"s3SchemaVersion":"1.0","configurationId":"Config","bucket":{"name":"images","ownerIdentity":{"principalId":"21EJ9HYV110O8NVX2VMS"},"arn":"arn:aws:s3:::images"},"object":{"key":"gopher.jpg","size":162023,"eTag":"5337769ffa594e742408ad3f30713cd7","contentType":"image/jpeg","userMetadata":{"content-type":"image/jpeg"},"versionId":"1","sequencer":"1562A792DAA53426"}},"source":{"host":"","port":"","userAgent":"Obstor (linux; amd64) minio-go/v6.0.8 mc/DEVELOPMENT.GOGET"}}]}
```
