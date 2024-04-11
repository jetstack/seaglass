# seaglass

⚠️ This tool is currently a proof of concept. There will be bugs and it is very
much not ready for production usage.

Seaglass is a command line tool and Go package that makes it simple to discover
content in container registries.

It implements registry-specific functionality which makes listing repositories
and artifacts more efficient and useful than with the standard v2 registry API.

## Clients

Seaglass will select the most appropriate client based on registry hostname. If
there's no specific client for the host, then it will fallback to using the
`/v2` API directly.

Implemented:

- GitHub Container Registry (`ghcr.io`)
- Docker Hub (`docker.io`, `*.docker.io`)
- Google Container Registry (`gcr.io`, `*.gcr.io`, `*.k8s.io`)
- Google Artifact Registry (`*.pkg.dev`)
- Registry v2 API (`*`)

TODO:

- [ ] Azure Container Registry
- [ ] AWS ECR
- [ ] Harbor
- [ ] Quay
- [ ] Artifactory?
- [ ] Sonatype Nexus?
- ???

## Install

Checkout this repo and build the project locally:

```shell
go build .
```

Either put the resulting `seaglass` binary in your PATH, or run it directly:

```shell
./seaglass --help
```

## Usage

### List repositories

List the repositories directly under a given path:

```shell
$ seaglass repos gcr.io/your-project
gcr.io/your-project/one
gcr.io/your-project/two
gcr.io/your-project/three
```

```shell
$ seaglass repos gcr.io/your-project/one
gcr.io/your-project/one/a
gcr.io/your-project/one/b
```

You can use the `--recursive` flag to list all the repositories under the path:

```shell
$ seaglass repos gcr.io/your-project
gcr.io/your-project/one
gcr.io/your-project/one/a
gcr.io/your-project/one/b
gcr.io/your-project/two
gcr.io/your-project/two/x
gcr.io/your-project/three
gcr.io/your-project/three/y
gcr.io/your-project/three/z
```

Note, Seaglass requires that the registry host is provided in the reference, even for
Docker Hub images:

```shell
$ seaglass repos index.docker.io/jetstack
index.docker.io/jetstack/bio-docker-watcher
index.docker.io/jetstack/cloud-billing-exporter
index.docker.io/jetstack/contain
index.docker.io/jetstack/dockernews-web
index.docker.io/jetstack/echoloop
index.docker.io/jetstack/elasticsearch-pet
index.docker.io/jetstack/hello-world
index.docker.io/jetstack/helloworld
index.docker.io/jetstack/hyperkube
index.docker.io/jetstack/hyperkube-amd64
index.docker.io/jetstack/kube-lego
index.docker.io/jetstack/kubectl
index.docker.io/jetstack/mongodb
index.docker.io/jetstack/mongodb-client
index.docker.io/jetstack/mongodb-replica-set
index.docker.io/jetstack/mongodb-server
index.docker.io/jetstack/mycms-api
index.docker.io/jetstack/mykrobe-predictor
index.docker.io/jetstack/nginx-ingress
index.docker.io/jetstack/nginx-ingress-controller
index.docker.io/jetstack/nginx-proxy
index.docker.io/jetstack/nginx-ssl-proxy
index.docker.io/jetstack/node-cms-file
index.docker.io/jetstack/node-cms-mongo
index.docker.io/jetstack/node-upvote
index.docker.io/jetstack/simple-cms
index.docker.io/jetstack/simple-server
index.docker.io/jetstack/simple-service
index.docker.io/jetstack/slingshot-cp-ansible-k8s-coreos
index.docker.io/jetstack/slingshot-ip-terraform-aws-coreos
index.docker.io/jetstack/slingshot-ip-vagrant-coreos
index.docker.io/jetstack/vault
```

### List Manifests

List all the manifests in a repository.

```shell
$ seaglass manifests ghcr.io/jetstack/tally
ghcr.io/jetstack/tally@sha256:87f4f96fc7493d7e77c628583e0cf776a90bf95fd83168e9c0e8fd6db5624656
ghcr.io/jetstack/tally@sha256:ca977a9d59454e78aae934097a482981398b696bb5d48de9992dd269bd2d6af1
ghcr.io/jetstack/tally@sha256:378faebd92d5baf83230affaf54ac61309bea60d23b38a83af97d6dd6656f5f1
ghcr.io/jetstack/tally@sha256:ce06d36166ca345dbcaf60e77666193cc3aa7cc99850fb8d03fca9e58efe72b1
ghcr.io/jetstack/tally@sha256:66cfb69847a7acd2823022504a23eb3ae3181fe44d4cb07aeb4f0f9c46095a94
ghcr.io/jetstack/tally@sha256:f4e58f42d5f6d724fd059fcac82b3266d149c909c24f822a936ff364547fba53
ghcr.io/jetstack/tally@sha256:b2a73a4fd2a96e860c2595483bcae12dc8d52ae7703eb46bd028a0ddd30066a8
ghcr.io/jetstack/tally@sha256:b794a2e25d51e5771c80da7a78e8ef5fb9e08f04aec2a7a98497dd25e05858fd
ghcr.io/jetstack/tally@sha256:a554d8a23a3a3e7fb497a1d1a2376f40b28c69f0d83aafa24c0531ec09cf37b3
```

List all the manifests in a repository, recursively.

```shell
$ seaglass manifests ghcr.io/jetstack --recursive
ghcr.io/jetstack/platform/deploy-helper@sha256:ccf60b7a872b5e71c8bc0ccd9bacabf79bf6d0730fc2e65061016a14bb3fabe9
ghcr.io/jetstack/platform/deploy-helper@sha256:e8de17335c009f085fb39bbb2c20406a9c8d645c1268c679868f379f78a0c1f3
ghcr.io/jetstack/platform/deploy-helper@sha256:cd48849f9e9e50bf5d1514422d78a055ba4b13b1d7a62e40e4dfba478f654940
ghcr.io/jetstack/platform/deploy-helper@sha256:43d5327f6a702f191a3fd3fdacd479ee7b97da0cd4631ba7e767f751772c3e82
ghcr.io/jetstack/platform/deploy-helper@sha256:be8a4a040a4a563bf7484da32f16787cf39478469fab90050f3f9c76134fa1ad
ghcr.io/jetstack/platform/deploy-helper@sha256:de36c6ad309fdcba8c3a3f8a45bf703660378de34454079110c1f6e8de6521fb
ghcr.io/jetstack/platform/deploy-helper@sha256:ca548cb1932ec287516daca40afb091f815c62d6b9f7aaffe64d64b4479633b7
ghcr.io/jetstack/platform/deploy-helper@sha256:1516e4034f3889b34aa35165188f3a0c69be2914dc887cb3a29ea400289744ef
ghcr.io/jetstack/platform/deploy-helper@sha256:2f016e5aa3c693139d538b0c47ca3e9e1e828fe573b4f6859ea22a2dd3fc8074
ghcr.io/jetstack/platform/deploy-helper@sha256:9c1a74dadb1ddbf1db72b879e2e1f1c2d87cb611091da2e0d664d359f554f29f
ghcr.io/jetstack/platform/deploy-helper@sha256:f4decb5801b5a3b837c666d36a886ef2c399221259ec036fb8f586318bea6b07
ghcr.io/jetstack/platform/deploy-helper@sha256:1ab6a4fbcda7e7fa15efc24428d7bf2fb039d8b24b1a8fcc9ad89cf8885c16ca
ghcr.io/jetstack/platform/deploy-helper@sha256:16969701e82148cdf46be79b7f79fc6e58a8bdcde65c63356f62105e5f40d914
ghcr.io/jetstack/platform/deploy-helper@sha256:cbf89c70d97b1d5cbd080735df9a1a1128892c6893ab021fd23b82e825090256
ghcr.io/jetstack/platform/deploy-helper@sha256:2046c4a6d79349033097c5a72faa8a6d5d86c2728651a9ca61d2c385554dc632
ghcr.io/jetstack/platform/deploy-helper@sha256:8861aa258b69127dc15ca3c5f11d4918ec1c041ecff8b42241bef19f824f10f4
ghcr.io/jetstack/platform/deploy-helper@sha256:7111bd5b2bbe5d97ad75f90b3dd5de767f00a6cbfd53ad8989f10a67d931334f
ghcr.io/jetstack/platform/deploy-helper@sha256:b05948439ec55f546404991874933c5c5a519c81d2f7a423344be2abce31ea3f
ghcr.io/jetstack/dependency-track-exporter@sha256:0a506c1b8bf571fa681eadd70e11b6c1b05c2ba6e4655dea911f1ff79e2f3223
ghcr.io/jetstack/dependency-track-exporter@sha256:0bc7942e77c18363cd116c19385cf24aac4e3a54529c3bc887fc94aeb1dc7f4b
ghcr.io/jetstack/dependency-track-exporter@sha256:3f91577141bdbe85f9fb132134e60ea48f401742c3219cf7da7ec4ac3a9508f6
ghcr.io/jetstack/dependency-track-exporter@sha256:9c830afd1ec44bbbe559788d45ff14695c909ddbe4abe96e75cc947f27662e1c
ghcr.io/jetstack/dependency-track-exporter@sha256:2752003b1fbb73cd0896bb248d11a7bab1246278b22f515b92384fa0126b3084
ghcr.io/jetstack/dependency-track-exporter@sha256:704cfa06c3096dc307d8df14924cb53719fddb952fd97b94565f25d19fa9a3b8
ghcr.io/jetstack/dependency-track-exporter@sha256:41da73349d7f113b00007ffb31514a248596b08e2e0aa403ebf38fed8ab6998a
ghcr.io/jetstack/dependency-track-exporter@sha256:69f8fb7a10873c066b4c37769063038f91454b4f86d17e3fc82a5a6a1e269a26
ghcr.io/jetstack/dependency-track-exporter@sha256:2b01cf77813841758f9845c1727c93c3e305abceaf849c3094598e2de4300d0d
ghcr.io/jetstack/dependency-track-exporter@sha256:a56b6ba879fe75e9a0ed23f062eda5b6969c1e279f3dd4a8b464317fbce56269
ghcr.io/jetstack/dependency-track-exporter@sha256:e35879a321960ce607596a00eff1224b51e2ce142bdb0c127188db60ec1cca0e
ghcr.io/jetstack/dependency-track-exporter@sha256:944c5710c87e4d83a914e8fd7d69a171c0282402003b82aad18b642c4e2bbcdd
ghcr.io/jetstack/zenko-cloudserver-nonroot@sha256:2b373ed44f6b9c02d297f701d8079f8315507ef02e42f37ca3fc51fbea90c1be
```
