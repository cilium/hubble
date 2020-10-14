module github.com/cilium/hubble

go 1.14

require (
	github.com/cilium/cilium v1.9.0-rc0.0.20201007145948-2c42c8c8edcd
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.2
	github.com/gordonklaus/ineffassign v0.0.0-20200809085317-e36bfde3bb78
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
	google.golang.org/grpc v1.33.0
	gopkg.in/yaml.v2 v2.3.0
	honnef.co/go/tools v0.0.1-2019.2.3
)

replace (
	github.com/miekg/dns => github.com/cilium/dns v1.1.4-0.20190417235132-8e25ec9a0ff3
	github.com/optiopay/kafka => github.com/cilium/kafka v0.0.0-20180809090225-01ce283b732b
	k8s.io/client-go => github.com/cilium/client-go v0.0.0-20200917084247-85ed8d558b9c
	sigs.k8s.io/controller-tools => github.com/christarazi/controller-tools v0.3.1-0.20200911184030-7e668c1fb4c2
	sigs.k8s.io/structured-merge-diff/v4 => github.com/christarazi/structured-merge-diff/v4 v4.0.2-0.20200917183246-1cc601931628
)
