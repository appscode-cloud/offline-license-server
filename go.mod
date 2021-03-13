module github.com/appscodelabs/offline-license-server

go 1.14

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/avct/uasurfer v0.0.0-20191028135549-26b5daa857f1
	github.com/davegardnerisme/phonegeocode v0.0.0-20160120101024-a49b977f8889
	github.com/elazarl/go-bindata-assetfs v1.0.1 // indirect
	github.com/go-macaron/auth v0.0.0-20161228062157-884c0e6c9b92
	github.com/go-macaron/bindata v0.0.0-20200308113348-9fced76aaa6e
	github.com/go-macaron/binding v0.0.0-00010101000000-000000000000
	github.com/go-macaron/cache v0.0.0-20200329073519-53bb48172687
	github.com/go-macaron/cors v0.0.0-20210206180111-00b7f53a9308
	github.com/gobuffalo/flect v0.2.2
	github.com/gocarina/gocsv v0.0.0-20201208093247-67c824bc04d4
	github.com/google/uuid v1.1.2
	github.com/gorilla/schema v1.2.0 // indirect
	github.com/himalayan-institute/zoom-lib-golang v1.0.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/k3a/html2text v1.0.7
	github.com/mailgun/mailgun-go/v4 v4.3.3
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/oschwald/geoip2-golang v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/rickb777/date v1.15.3
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/tchap/go-patricia v2.3.0+incompatible // indirect
	github.com/yuin/goldmark v1.2.1
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	gomodules.xyz/blobfs v0.1.7
	gomodules.xyz/cert v1.2.0
	gomodules.xyz/email-providers v0.1.4
	gomodules.xyz/freshsales-client-go v0.0.1
	gomodules.xyz/gdrive-utils v0.0.0-20210313135106-78213dfc3fc4
	gomodules.xyz/homedir v0.0.0-20201104190528-bcd4d5d94b84
	gomodules.xyz/password-generator v0.2.6
	gomodules.xyz/pointer v0.0.0-20201105040656-991dd254b680
	gomodules.xyz/sets v0.0.0-20210218105342-2efe2fb519a2
	gomodules.xyz/x v0.0.0-20201105065653-91c568df6331
	google.golang.org/api v0.39.0
	gopkg.in/macaron.v1 v1.4.0
	k8s.io/apimachinery v0.18.9
	sigs.k8s.io/yaml v1.2.0
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
	google.golang.org/api => google.golang.org/api v0.36.0
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20201203001206-6486ece9c497
	google.golang.org/grpc => google.golang.org/grpc v1.34.0
	k8s.io/api => github.com/kmodules/api v0.18.4-0.20200524125823-c8bc107809b9
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.19.0-alpha.0.0.20200520235721-10b58e57a423
	k8s.io/apiserver => github.com/kmodules/apiserver v0.18.4-0.20200521000930-14c5f6df9625
	k8s.io/client-go => k8s.io/client-go v0.18.3
	k8s.io/kubernetes => github.com/kmodules/kubernetes v1.19.0-alpha.0.0.20200521033432-49d3646051ad
)
