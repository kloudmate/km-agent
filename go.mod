module github.com/kloudmate/km-agent

go 1.23.4

require (
	github.com/kardianos/service v1.2.2
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	go.opentelemetry.io/collector/component v0.117.0
	go.opentelemetry.io/collector/confmap v1.23.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.22.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.22.0
	go.opentelemetry.io/collector/confmap/provider/httpprovider v1.22.0
	go.opentelemetry.io/collector/confmap/provider/httpsprovider v1.21.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.22.0
	go.opentelemetry.io/collector/connector v0.116.0
	go.opentelemetry.io/collector/connector/forwardconnector v0.115.0
	go.opentelemetry.io/collector/exporter v0.116.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.115.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.115.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.115.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.115.0
	go.opentelemetry.io/collector/extension v0.116.0
	go.opentelemetry.io/collector/extension/memorylimiterextension v0.115.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.115.0
	go.opentelemetry.io/collector/featuregate v1.22.0
	go.opentelemetry.io/collector/otelcol v0.116.0
	go.opentelemetry.io/collector/processor v0.116.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.115.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.115.0
	go.opentelemetry.io/collector/receiver v0.117.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.115.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.115.0
	go.uber.org/goleak v1.3.0
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.4.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.117.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.116.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.116.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus-community/windows_exporter v0.27.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/urfave/cli/v2 v2.27.5 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.opentelemetry.io/collector/filter v0.117.0 // indirect
	go.opentelemetry.io/collector/scraper v0.117.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.117.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.23.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.117.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.116.0
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.61.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shirou/gopsutil/v4 v4.24.11 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/collector v0.115.0 // indirect
	go.opentelemetry.io/collector/client v1.21.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.117.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.117.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.115.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.21.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.115.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.115.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.21.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.21.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.21.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.117.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.21.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.115.0 // indirect
	go.opentelemetry.io/collector/connector/connectorprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.116.0 // indirect
	go.opentelemetry.io/collector/consumer v1.23.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.117.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/consumererrorprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/consumer/consumerprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.117.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/exporterhelperprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.116.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.115.0 // indirect
	go.opentelemetry.io/collector/extension/experimental/storage v0.115.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.116.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.116.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.116.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.115.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.115.0 // indirect
	go.opentelemetry.io/collector/pdata v1.23.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.117.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.117.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.117.0 // indirect
	go.opentelemetry.io/collector/pipeline/pipelineprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/processor/processorprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.116.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverprofiles v0.115.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.117.0 // indirect
	go.opentelemetry.io/collector/semconv v0.117.0 // indirect
	go.opentelemetry.io/collector/service v0.116.0
	go.opentelemetry.io/contrib/bridges/otelzap v0.6.0 // indirect
	go.opentelemetry.io/contrib/config v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.31.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.56.0 // indirect
	go.opentelemetry.io/otel v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.31.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.31.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.54.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.31.0 // indirect
	go.opentelemetry.io/otel/log v0.8.0 // indirect
	go.opentelemetry.io/otel/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/sdk v1.32.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.7.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.32.0 // indirect
	go.opentelemetry.io/otel/trace v1.32.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0
	golang.org/x/text v0.21.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241104194629-dd2ea8efbc28 // indirect
	google.golang.org/grpc v1.69.2
	google.golang.org/protobuf v1.36.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace go.opentelemetry.io/collector => github.com/open-telemetry/opentelemetry-collector v0.115.0

replace go.opentelemetry.io/collector/service => github.com/open-telemetry/opentelemetry-collector/service v0.115.0

replace go.opentelemetry.io/collector/connector => github.com/open-telemetry/opentelemetry-collector/connector v0.115.0

replace go.opentelemetry.io/collector/connector/connectortest => github.com/open-telemetry/opentelemetry-collector/connector/connectortest v0.115.0

replace go.opentelemetry.io/collector/component => github.com/open-telemetry/opentelemetry-collector/component v0.115.0

replace go.opentelemetry.io/collector/component/componenttest => github.com/open-telemetry/opentelemetry-collector/component/componenttest v0.115.0

replace go.opentelemetry.io/collector/pdata => github.com/open-telemetry/opentelemetry-collector/pdata v1.21.0

replace go.opentelemetry.io/collector/pdata/testdata => github.com/open-telemetry/opentelemetry-collector/pdata/testdata v0.115.0

replace go.opentelemetry.io/collector/pdata/pprofile => github.com/open-telemetry/opentelemetry-collector/pdata/pprofile v0.115.0

replace go.opentelemetry.io/collector/extension/zpagesextension => github.com/open-telemetry/opentelemetry-collector/extension/zpagesextension v0.115.0

replace go.opentelemetry.io/collector/extension => github.com/open-telemetry/opentelemetry-collector/extension v0.115.0

replace go.opentelemetry.io/collector/extension/experimental/storage => github.com/open-telemetry/opentelemetry-collector/extension/experimental/storage v0.115.0

replace go.opentelemetry.io/collector/exporter => github.com/open-telemetry/opentelemetry-collector/exporter v0.115.0

replace go.opentelemetry.io/collector/confmap => github.com/open-telemetry/opentelemetry-collector/confmap v1.21.0

replace go.opentelemetry.io/collector/config/configtelemetry => github.com/open-telemetry/opentelemetry-collector/config/configtelemetry v0.115.0

replace go.opentelemetry.io/collector/processor => github.com/open-telemetry/opentelemetry-collector/processor v0.115.0

replace go.opentelemetry.io/collector/consumer => github.com/open-telemetry/opentelemetry-collector/consumer v1.21.0

replace go.opentelemetry.io/collector/semconv => github.com/open-telemetry/opentelemetry-collector/semconv v0.115.0

replace go.opentelemetry.io/collector/receiver => github.com/open-telemetry/opentelemetry-collector/receiver v0.115.0

replace go.opentelemetry.io/collector/featuregate => github.com/open-telemetry/opentelemetry-collector/featuregate v1.21.0

replace go.opentelemetry.io/collector/config/configretry => github.com/open-telemetry/opentelemetry-collector/config/configretry v1.21.0

replace go.opentelemetry.io/collector/config/confighttp => github.com/open-telemetry/opentelemetry-collector/config/confighttp v0.115.0

replace go.opentelemetry.io/collector/config/internal => github.com/open-telemetry/opentelemetry-collector/config/internal v0.115.0

replace go.opentelemetry.io/collector/config/configauth => github.com/open-telemetry/opentelemetry-collector/config/configauth v0.115.0

replace go.opentelemetry.io/collector/extension/auth => github.com/open-telemetry/opentelemetry-collector/extension/auth v0.115.0

replace go.opentelemetry.io/collector/config/configcompression => github.com/open-telemetry/opentelemetry-collector/config/configcompression v1.21.0

replace go.opentelemetry.io/collector/config/configtls => github.com/open-telemetry/opentelemetry-collector/config/configtls v1.21.0

replace go.opentelemetry.io/collector/config/configopaque => github.com/open-telemetry/opentelemetry-collector/config/configopaque v1.21.0

replace go.opentelemetry.io/collector/consumer/consumerprofiles => github.com/open-telemetry/opentelemetry-collector/consumer/consumerprofiles v0.115.0

replace go.opentelemetry.io/collector/consumer/consumertest => github.com/open-telemetry/opentelemetry-collector/consumer/consumertest v0.115.0

replace go.opentelemetry.io/collector/client => github.com/open-telemetry/opentelemetry-collector/client v1.21.0

replace go.opentelemetry.io/collector/component/componentstatus => github.com/open-telemetry/opentelemetry-collector/component/componentstatus v0.115.0

replace go.opentelemetry.io/collector/extension/extensioncapabilities => github.com/open-telemetry/opentelemetry-collector/extension/extensioncapabilities v0.115.0

replace go.opentelemetry.io/collector/receiver/receiverprofiles => github.com/open-telemetry/opentelemetry-collector/receiver/receiverprofiles v0.115.0

replace go.opentelemetry.io/collector/receiver/receivertest => github.com/open-telemetry/opentelemetry-collector/receiver/receivertest v0.115.0

replace go.opentelemetry.io/collector/processor/processorprofiles => github.com/open-telemetry/opentelemetry-collector/processor/processorprofiles v0.115.0

replace go.opentelemetry.io/collector/connector/connectorprofiles => github.com/open-telemetry/opentelemetry-collector/connector/connectorprofiles v0.115.0

replace go.opentelemetry.io/collector/exporter/exporterprofiles => github.com/open-telemetry/opentelemetry-collector/exporter/exporterprofiles v0.115.0

replace go.opentelemetry.io/collector/pipeline => github.com/open-telemetry/opentelemetry-collector/pipeline v0.115.0

replace go.opentelemetry.io/collector/pipeline/pipelineprofiles => github.com/open-telemetry/opentelemetry-collector/pipeline/pipelineprofiles v0.115.0

replace go.opentelemetry.io/collector/exporter/exportertest => github.com/open-telemetry/opentelemetry-collector/exporter/exportertest v0.115.0

replace go.opentelemetry.io/collector/processor/processortest => github.com/open-telemetry/opentelemetry-collector/processor/processortest v0.115.0

replace go.opentelemetry.io/collector/consumer/consumererror => github.com/open-telemetry/opentelemetry-collector/consumer/consumererror v0.115.0

replace go.opentelemetry.io/collector/internal/fanoutconsumer => github.com/open-telemetry/opentelemetry-collector/internal/fanoutconsumer v0.115.0

replace go.opentelemetry.io/collector/extension/extensiontest => github.com/open-telemetry/opentelemetry-collector/extension/extensiontest v0.115.0

replace go.opentelemetry.io/collector/extension/auth/authtest => github.com/open-telemetry/opentelemetry-collector/extension/auth/authtest v0.115.0

replace go.opentelemetry.io/collector/scraper => github.com/open-telemetry/opentelemetry-collector/scraper v0.115.0
