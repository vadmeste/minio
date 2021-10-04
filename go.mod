module github.com/minio/minio

go 1.16

require (
	cloud.google.com/go/storage v1.10.0
	github.com/Azure/azure-pipeline-go v0.2.2
	github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/Shopify/sarama v1.27.2
	github.com/VividCortex/ewma v1.1.1
	github.com/alecthomas/participle v0.2.1
	github.com/bcicen/jstream v1.0.1
	github.com/beevik/ntp v0.3.0
	github.com/bits-and-blooms/bloom/v3 v3.0.1
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/cheggaaa/pb v1.0.29
	github.com/colinmarc/hdfs/v2 v2.2.0
	github.com/coredns/coredns v1.4.0
	github.com/dchest/siphash v1.2.1
	github.com/djherbis/atime v1.0.0
	github.com/dswarbrick/smart v0.0.0-20190505152634-909a45200d6d
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.3.0
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/fatih/color v1.12.0
	github.com/go-ldap/ldap/v3 v3.2.4
	github.com/go-openapi/loads v0.20.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/jcmturner/gokrb5/v8 v8.4.2
	github.com/json-iterator/go v1.1.11
	github.com/klauspost/compress v1.13.5
	github.com/klauspost/cpuid/v2 v2.0.6
	github.com/klauspost/pgzip v1.2.5
	github.com/klauspost/readahead v1.3.1
	github.com/klauspost/reedsolomon v1.9.13
	github.com/lib/pq v1.9.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/miekg/dns v1.1.35
	github.com/minio/cli v1.22.0
	github.com/minio/console v0.10.2
	github.com/minio/csvparser v1.0.0
	github.com/minio/highwayhash v1.0.2
	github.com/minio/kes v0.14.0
	github.com/minio/madmin-go v1.1.6
	github.com/minio/minio-go/v7 v7.0.15-0.20210928020726-a58653d41dd8
	github.com/minio/parquet-go v1.0.0
	github.com/minio/pkg v1.1.3
	github.com/minio/selfupdate v0.3.1
	github.com/minio/sha256-simd v1.0.0
	github.com/minio/simdjson-go v0.2.1
	github.com/minio/sio v0.3.0
	github.com/minio/zipindex v0.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/montanaflynn/stats v0.6.6
	github.com/nats-io/nats-server/v2 v2.3.2
	github.com/nats-io/nats-streaming-server v0.21.2 // indirect
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	github.com/nats-io/stan.go v0.8.3
	github.com/ncw/directio v1.0.5
	github.com/nsqio/go-nsq v1.0.8
	github.com/philhofer/fwd v1.1.1
	github.com/pierrec/lz4 v2.6.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/procfs v0.7.3
	github.com/rs/cors v1.7.0
	github.com/secure-io/sio-go v0.3.1
	github.com/shirou/gopsutil/v3 v3.21.7
	github.com/streadway/amqp v1.0.0
	github.com/tinylib/msgp v1.1.6-0.20210521143832-0becd170c402
	github.com/valyala/bytebufferpool v1.0.0
	github.com/valyala/tcplisten v0.0.0-20161114210144-ceec8f93295a
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c
	github.com/yargevad/filepathx v1.0.0
	go.etcd.io/etcd/api/v3 v3.5.0-beta.4
	go.etcd.io/etcd/client/v3 v3.5.0-beta.4
	go.opencensus.io v0.22.5 // indirect
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.1-0.20210329175301-c23abee72d19
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/sys v0.0.0-20210819135213-f52c844e1c1c
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/tools v0.1.1 // indirect
	google.golang.org/api v0.31.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/minio/madmin-go => ../madmin-go
	github.com/minio/pkg => ../pkg
)
