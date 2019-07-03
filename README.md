![MicroBadger Size](https://img.shields.io/microbadger/image-size/ragon000/fossag-fotobox.svg)
![Docker Stars](https://img.shields.io/docker/stars/ragon000/fossag-fotobox.svg)
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/ragon000/fossag-fotobox.svg)
# fossag-fotobox


A webserver to host your cloud fotobox

# Installation

## Docker

```bash
docker run -p "8080:8080" ragon000/fossag-fotobox
```

for SSL you should use a reverse proxy like nginx or traefik

## Bare Metal

```bash
git clone https://github.com/ragon000/fossag-fotobox.git
cd fossag-fotobox
go run
```

for SSL you should use a reverse proxy like nginx or traefik

# Usage

to upload a picture use that curl string

```bash
curl http://<url>:<port>/upload -F "uploadFile=@<absolute path>"
```

your open clients should refresh automatically when a new picture arrives

