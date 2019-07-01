# fossag-fotobox

A webserver to host your cloud fotobox

# Usage

to upload a picture use that curl string

```bash
curl http://<url>:8080/upload -F "uploadFile=@<absolute path>"
```

your open clients should refresh automatically when a new picture arrives

