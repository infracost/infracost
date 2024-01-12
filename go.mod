module github.com/infracost/infracost

go 1.21

toolchain go1.21.3

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Rhymond/go-money v1.0.10
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.38
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.24.3
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.21.11
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.17.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.84.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.33.0
	github.com/awslabs/goformation/v4 v4.19.5
	github.com/briandowns/spinner v1.15.0
	github.com/dave/dst v0.27.2
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.15.0
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.3.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/jedib0t/go-pretty/v6 v6.2.1
	github.com/joho/godotenv v1.4.0
	github.com/json-iterator/go v1.1.12
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	github.com/tidwall/gjson v1.17.0
	github.com/zclconf/go-cty v1.14.0
	golang.org/x/crypto v0.17.0
	golang.org/x/mod v0.13.0
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go-v2/service/eks v1.27.0
	github.com/hashicorp/terraform-config-inspect v0.0.0-20210625153042-09f34846faab
	golang.org/x/sys v0.15.0 // indirect
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.36 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.41 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.42 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.7.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.13.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.21.5 // indirect
	github.com/aws/smithy-go v1.14.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.1-vault // indirect
	github.com/hashicorp/hcl/v2 v2.16.2
	github.com/imdario/mergo v0.3.13
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/sanathkr/go-yaml v0.0.0-20170819195128-ed9d249f429b // indirect
	github.com/sanathkr/yaml v0.0.0-20170819201035-0056894fa522 // indirect
	github.com/slack-go/slack v0.12.3
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	golang.org/x/text v0.14.0
	golang.org/x/tools v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.35 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/sync v0.4.0
)

require (
	github.com/alecthomas/jsonschema v0.0.0-20211209230136-e2b41affa5c1
	github.com/channelmeter/iso8601duration v0.0.0-20150204201828-8da3af7a2a61
	github.com/fatih/camelcase v1.0.0
	github.com/go-git/go-billy/v5 v5.5.0
	github.com/go-git/go-git/v5 v5.11.0
	github.com/google/go-github/v41 v41.0.0
	github.com/gruntwork-io/go-commons v0.17.1
	github.com/gruntwork-io/terragrunt v0.52.4
	github.com/hashicorp/go-retryablehttp v0.7.5
	github.com/hashicorp/go-terraform-address v0.0.0-20210506203813-2cc4f0f34da8
	github.com/hashicorp/golang-lru/v2 v2.0.6
	github.com/hashicorp/terraform-svchost v0.1.0
	github.com/maruel/panicparse/v2 v2.3.1
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/pkg/browser v0.0.0-20201207095918-0426ae3fba23
	github.com/pkg/profile v1.2.1
	github.com/rs/zerolog v1.31.0
	github.com/shurcooL/githubv4 v0.0.0-20220115235240-a14260e6f8a2
	github.com/shurcooL/graphql v0.0.0-20200928012149-18c5c3165e3a
	github.com/soongo/path-to-regexp v1.6.4
	github.com/withfig/autocomplete-tools/packages/cobra v1.2.0
	github.com/xanzy/go-gitlab v0.86.0
	golang.org/x/oauth2 v0.8.0
)

require (
	cloud.google.com/go/compute v1.19.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	dario.cat/mergo v1.0.0 // indirect
	filippo.io/age v1.0.0 // indirect
	github.com/Azure/azure-sdk-for-go v63.3.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.26 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/apparentlymart/go-versions v1.0.1 // indirect
	github.com/armon/go-metrics v0.3.10 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.15.5 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/creack/pty v1.1.11 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-github/v35 v35.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/goware/prefixer v0.0.0-20160118172347-395022866408 // indirect
	github.com/gruntwork-io/gruntwork-cli v0.7.0 // indirect
	github.com/gruntwork-io/terratest v0.41.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.10 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.3 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform v0.15.3 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.0 // indirect
	github.com/hashicorp/vault/api v1.5.0 // indirect
	github.com/hashicorp/vault/sdk v0.4.1 // indirect
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jstemmer/go-junit-report v1.0.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/lib/pq v1.10.5 // indirect
	github.com/mattn/go-zglob v0.0.3 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mitchellh/panicwrap v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/owenrumney/go-sarif v1.1.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/skeema/knownhosts v1.2.1 // indirect
	github.com/sourcegraph/go-lsp v0.0.0-20200429204803-219e11d77f5d // indirect
	github.com/sourcegraph/jsonrpc2 v0.2.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/tchap/go-patricia/v2 v2.3.1 // indirect
	github.com/terraform-linters/tflint v0.46.1 // indirect
	github.com/terraform-linters/tflint-plugin-sdk v0.16.1 // indirect
	github.com/terraform-linters/tflint-ruleset-terraform v0.3.0 // indirect
	github.com/urfave/cli/v2 v2.25.5 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.mozilla.org/gopgagent v0.0.0-20170926210634-4d7ea76ff71a // indirect
	go.mozilla.org/sops/v3 v3.7.3 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/storage v1.28.1 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.122 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bmatcuk/doublestar v1.3.4
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.7.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2
	github.com/hashicorp/go-getter v1.7.3
	github.com/hashicorp/go-safetemp v1.0.0
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/go-version v1.6.0
	github.com/heimdalr/dag v1.3.1
	github.com/iancoleman/orderedmap v0.2.0 // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/open-policy-agent/opa v0.46.1
	github.com/otiai10/copy v1.7.0
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.1.0 // indirect
	github.com/zclconf/go-cty-yaml v1.0.3
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	google.golang.org/api v0.114.0
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/jedib0t/go-pretty/v6 => github.com/aliscott/go-pretty/v6 v6.1.1-0.20210226104003-408905a61c8e

replace github.com/spf13/cobra => github.com/spf13/cobra v1.4.0

//replace github.com/gruntwork-io/terragrunt => github.com/infracost/terragrunt v0.47.1-0.20231103101711-77c5e7d4d795
replace github.com/gruntwork-io/terragrunt => github.com/infracost/terragrunt v0.47.1-0.20231211141424-6a52de9a284a

replace github.com/heimdalr/dag => github.com/aliscott/dag v1.3.2-0.20231115114512-4ce18c825f94
