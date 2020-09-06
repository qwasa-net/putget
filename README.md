# putget

## putget — my little **go** experiment

**putget** is a super stupid http service to:

* upload series of files into buckets

* get the most recent file from the bucket


### Run *putget* service

```bash
> ./putget -bind "localhost:8900" -files-root "/tmp/putget.files/" -url-root "/"
```


### Use *putget*
```bash

> curl -X POST http://localhost:8900/bucket_name/ --data-binary @image1.jpg
> curl -X POST http://localhost:8900/bucket_name/ --data-binary @imageN.jpg
> curl -X POST http://localhost:8900/another_bucket/ --data-binary @imageM.jpg
…
> curl http://localhost:8900/bucket_name/ -O downloaded_image.jpg
> cmp downloaded_image.jpg imageN.jpg
```

## etc

* Authorization, limits and access control must be handled in the router/proxy service, e.g. **nginx**.

* DataBase module is not implemented (yet?) — dumb memory map is used

* This is a toy project, not intended for production use.

:wq
