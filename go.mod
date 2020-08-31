module github.com/cilium/hubble

go 1.14

require (
	github.com/cilium/cilium v1.8.0-rc1.0.20200805113126-d80e66fd95fe
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.1
	github.com/gordonklaus/ineffassign v0.0.0-20200809085317-e36bfde3bb78
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de
	google.golang.org/grpc v1.27.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
)

replace (
	github.com/miekg/dns => github.com/cilium/dns v1.1.4-0.20190417235132-8e25ec9a0ff3
	github.com/optiopay/kafka => github.com/cilium/kafka v0.0.0-20180809090225-01ce283b732b
	k8s.io/client-go => github.com/cilium/client-go v0.0.0-20200417200322-b77c886899ef
)
