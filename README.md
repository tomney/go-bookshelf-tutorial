This repository was created as part of the following [exercise](https://cloud.google.com/go/docs/tutorials/bookshelf-on-kubernetes-engine).

A kubernetes engine cluster can be created VIA:
```
gcloud container clusters create bookshelf \
    --scopes "cloud-platform" \
    --num-nodes 2 \
    --enable-basic-auth \
    --issue-client-certificate \
    --enable-ip-alias \
    --zone us-central1-a
```

This code is copied and in some cases changed from this [repository](github.com/GoogleCloudPlatform/golang-samples/getting-started/bookshelf
)