---
title: Set up Storage for Loading Models
---

You will need access to an S3-compatible object storage, for example [MinIO](https://github.com/minio/minio) or [IBM Cloud Object Storage](https://www.ibm.com/cloud/object-storage). To provide access to the object storage, use the `storage-config` secret.

## Deploy a sample model from our shared object storage

A set of example models are shared via an IBM Cloud COS instance to use when getting started with Model-Mesh Serving and experimenting with the provided runtimes. Access to this COS instance needs to be set up in the `storage-config` secret.

For the shared COS instance, the `secretKey` to access the models is `modelmesh-example-models`. If you used the `--quickstart` option of the install script, the `secretKey` to use for the local MinIO storage instance is `localMinIO`. If you use a different key value, be sure to update the `spec.storage.s3.secretKey` value in the Predictor.

It should look something like:

```shell
$ kubectl describe secret storage-config
Name:         storage-config
Namespace:    modelmesh
Labels:       indigo-service=all
Annotations:
Type:         Opaque

Data
====
modelmesh-example-models:  308 bytes
```

If you installed Model-Mesh Serving using the operator, you will have to configure the `storage-config` secret for access:

```shell
$ kubectl patch secret/storage-config -p '{"data": {"wml-serving-example-models": "ewogICJ0eXBlIjogInMzIiwKICAiYWNjZXNzX2tleV9pZCI6ICJlY2I5ODNmMTE4MjI0MjNjYTllNDg3Zjg5OGQ1NGE4ZiIsCiAgInNlY3JldF9hY2Nlc3Nfa2V5IjogImNkYmVmZjZhMzJhZWY2YzIzNzRhZTY5ZWVmNTAzZTZkZDBjOTNkNmE3NGJjMjQ2NyIsCiAgImVuZHBvaW50X3VybCI6ICJodHRwczovL3MzLnVzLXNvdXRoLmNsb3VkLW9iamVjdC1zdG9yYWdlLmFwcGRvbWFpbi5jbG91ZCIsCiAgInJlZ2lvbiI6ICJ1cy1zb3V0aCIsCiAgImRlZmF1bHRfYnVja2V0IjogIndtbC1zZXJ2aW5nLWV4YW1wbGUtbW9kZWxzLXB1YmxpYyIKfQo="}}'
```

For reference the contents of the secret value for the `wml-serving-example-models` entry looks like:

```json
{
  "type": "s3",
  "access_key_id": "ecb983f11822423ca9e487f898d54a8f",
  "secret_access_key": "cdbeff6a32aef6c2374ae69eef503e6dd0c93d6a74bc2467",
  "endpoint_url": "https://s3.us-south.cloud-object-storage.appdomain.cloud",
  "region": "us-south",
  "default_bucket": "wml-serving-example-models-public"
}
```

<InlineNotification>

**Note** After updating the storage config secret, there may be a delay of up to 2 minutes until the change is picked up. You should take this into account when creating/updating Predictors that use storage keys which have just been added or updated - they may fail to load otherwise.

</InlineNotification>

## Deploy a model from your own object storage

1. Download sample model or use an existing model

Here we show an example using an MLeap-format model for [AirBnB Linear Regression](https://github.com/combust/mleap/raw/master/mleap-benchmark/src/main/resources/models/airbnb.model.lr.zip).

2. Add your MLeap saved model to S3-based object storage

A bucket in MinIO needs to be created to copy the model into, which either requires [MinIO Client](https://docs.min.io/docs/minio-client-quickstart-guide.html) or port-forwarding the minio service and logging in using the web interface.

```shell
# Install minio client
$ brew install minio/stable/mc
$ mc --help
NAME:
  mc - MinIO Client for cloud storage and filesystems.
....

# test setup - mc is pre-configured with https://play.min.io, aliased as "play".
# list all buckets in play
$ mc ls play

[2021-06-10 21:04:25 EDT]     0B 2063b651-92a3-4a20-a4a5-03a96e7c5a89/
[2021-06-11 02:40:33 EDT]     0B 5ddfe44282319c500c3a4f9b/
[2021-06-11 05:15:45 EDT]     0B 6dkmmiqcdho1zoloomsj3620cocs6iij/
[2021-06-11 02:39:54 EDT]     0B 9jo5omejcyyr62iizn02ex982eapipjr/
[2021-06-11 02:33:53 EDT]     0B a-test-zip/
[2021-06-11 09:14:28 EDT]     0B aio-ato/
[2021-06-11 09:14:29 EDT]     0B aio-ato-art/
...

# add cloud storage service
$ mc alias set <ALIAS> <YOUR-S3-ENDPOINT> [YOUR-ACCESS-KEY] [YOUR-SECRET-KEY]
# for example if you installed with --quickstart
$ mc alias set myminio http://localhost:9000 EXAMPLE_ACESS_KEY example/secret/EXAMPLEKEY
Added `myminio` successfully.

# create bucket
$ mc mb myminio/models/mleap
Bucket created successfully myminio/models/mleap.

$ mc tree myminio
myminio
└─ models
   └─ mleap

# copy object -- must copy into an existing bucket
$ mc cp ~/Downloads/airbnb.model.lr.zip myminio/models/mleap
...model.lr.zip:  14.90 KiB / 14.90 KiB  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  2.74 MiB/s 0s

$ mc ls myminio/models/mleap
[2021-06-11 11:55:48 EDT]  15KiB airbnb.model.lr.zip
```

3. Add a storage entry to the `storage-config` secret

Ensure there is a key defined in the common `storage-config` secret corresponding to the S3-based storage instance holding your model. The value of this secret key should be JSON like the following, `default_bucket` is optional.

Users can specify use of a custom certificate via the storage config `certificate` parameter. The custom certificate should be in the form of an embedded Certificate Authority (CA) bundle in PEM format.

Using MinIO the JSON contents look like:

```json
{
  "type": "s3",
  "access_key_id": "minioadmin",
  "secret_access_key": "minioadmin/K7JTCMP/EXAMPLEKEY",
  "endpoint_url": "http://127.0.0.1:9000:9000",
  "default_bucket": "",
  "region": "us-east"
}
```

Remember that after updating the storage config secret, there may be a delay of up to 2 minutes until the change is picked up. You should take this into account when creating/updating Predictors that use storage keys which have just been added or updated - they may fail to load otherwise.
