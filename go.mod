module github.com/appscodelabs/offline-license-server

go 1.14

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/elazarl/go-bindata-assetfs v1.0.1 // indirect
	github.com/go-macaron/bindata v0.0.0-20200308113348-9fced76aaa6e
	github.com/go-macaron/binding v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.1.2
	github.com/gorilla/schema v1.2.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/mailgun/mailgun-go/v4 v4.3.0
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/rickb777/date v1.14.3
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/yuin/goldmark v1.1.27
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	gomodules.xyz/blobfs v0.1.7
	gomodules.xyz/cert v1.2.0
	gomodules.xyz/email-providers v0.1.3
	gomodules.xyz/freshsales-client-go v0.0.0-20210204000632-49a286fb1763
	gomodules.xyz/x v0.0.0-20201105065653-91c568df6331
	google.golang.org/api v0.30.0
	gopkg.in/macaron.v1 v1.4.0
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	kmodules.xyz/client-go v0.0.0-20200817064010-b2e03dabff6b
)

replace (
	bitbucket.org/ww/goautoneg => gomodules.xyz/goautoneg v0.0.0-20120707110453-a547fc61f48d
	cloud.google.com/go => cloud.google.com/go v0.49.0
	git.apache.org/thrift.git => github.com/apache/thrift v0.13.0
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v35.0.0+incompatible
	github.com/Azure/go-ansiterm => github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0
	github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0
	github.com/Azure/go-autorest/autorest/azure/auth => github.com/Azure/go-autorest/autorest/azure/auth v0.2.0
	github.com/Azure/go-autorest/autorest/date => github.com/Azure/go-autorest/autorest/date v0.1.0
	github.com/Azure/go-autorest/autorest/mocks => github.com/Azure/go-autorest/autorest/mocks v0.2.0
	github.com/Azure/go-autorest/autorest/to => github.com/Azure/go-autorest/autorest/to v0.2.0
	github.com/Azure/go-autorest/autorest/validation => github.com/Azure/go-autorest/autorest/validation v0.1.0
	github.com/Azure/go-autorest/logger => github.com/Azure/go-autorest/logger v0.1.0
	github.com/Azure/go-autorest/tracing => github.com/Azure/go-autorest/tracing v0.5.0
	github.com/codeskyblue/go-sh => github.com/gomodules/go-sh v0.0.0-20200626022009-2c4bcc0f861d
	github.com/go-macaron/binding => github.com/gomodules/binding v0.0.0-20200811095614-c752727d2156
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.5
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.0.0
	github.com/spf13/afero => github.com/gomodules/afero v1.3.4-0.20200809021348-4197decdee8b
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => github.com/kmodules/api v0.18.4-0.20200524125823-c8bc107809b9
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423
	k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625
	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
)
