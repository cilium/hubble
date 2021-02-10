module github.com/cilium/hubble

go 1.14

require (
	github.com/cilium/cilium v1.9.0-rc1.0.20210209141502-b944040a9ec8
	github.com/google/go-cmp v0.5.4
	github.com/gordonklaus/ineffassign v0.0.0-20210209182638-d0e41b2fc8ed
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.2
	github.com/spf13/pflag v1.0.6-0.20200504143853-81378bbcd8a1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	google.golang.org/grpc v1.35.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0
	honnef.co/go/tools v0.1.1
)

// Replace directives from github.com/cilium/cilium. Keep in sync when updating Cilium!
replace (
	github.com/miekg/dns => github.com/cilium/dns v1.1.4-0.20190417235132-8e25ec9a0ff3
	github.com/optiopay/kafka => github.com/cilium/kafka v0.0.0-20180809090225-01ce283b732b
	sigs.k8s.io/controller-tools => github.com/christarazi/controller-tools v0.3.1-0.20200911184030-7e668c1fb4c2
)
