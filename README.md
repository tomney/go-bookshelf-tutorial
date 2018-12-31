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

This code is copied and in some cases changed from this [repository](https://github.com/GoogleCloudPlatform/golang-samples/tree/master/getting-started/bookshelf
)

Local SQL instance listens on 127.0.0.1:3300. You can proxy to the cloudSQL instance when running locally by running the command: 
```
./cloud_sql_proxy -instances="ace-shine-212419:us-east1:library"=tcp:3300
```